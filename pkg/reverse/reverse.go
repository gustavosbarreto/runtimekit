package reverse

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/gustavosbarreto/revwebsocketdial/pkg/connman"
	"github.com/gustavosbarreto/revwebsocketdial/pkg/revdial"
	"github.com/gustavosbarreto/revwebsocketdial/pkg/wsconnadapter"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		Subprotocols:    []string{"binary"},
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

const ProxiedFromDeviceHeader = "proxied-from-device"

func ProxyResponseFromDevice(w http.ResponseWriter, resp *http.Response) {
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.Header().Set(ProxiedFromDeviceHeader, "")

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	resp.Body.Close()
}

type Reverse struct {
	ConnectionPath    string
	DialerPath        string
	ConnectionHandler func(*http.Request) (string, error)
	connman           *connman.ConnectionManager
}

func NewReverse(connectionPath, dialerPath string) *Reverse {
	return &Reverse{
		ConnectionPath: connectionPath,
		DialerPath:     dialerPath,
		ConnectionHandler: func(r *http.Request) (string, error) {
			fmt.Println("default connection handler")

			id := r.Header.Get("X-CLIENT-ID")
			if id == "" {
				return id, errors.New("invalid id")
			}

			return id, nil
		},
		connman: connman.New(),
	}
}

func (r *Reverse) Router() http.Handler {
	router := mux.NewRouter()

	router.HandleFunc(r.ConnectionPath, func(res http.ResponseWriter, req *http.Request) {
		conn, err := upgrader.Upgrade(res, req, nil)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		id, err := r.ConnectionHandler(req)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			defer conn.Close()
			return
		}

		r.connman.Set(id, wsconnadapter.New(conn))
	}).Methods("GET")

	router.Handle(r.DialerPath, revdial.ConnHandler(upgrader)).Methods("GET")

	return router
}

func (r *Reverse) ProxyRequest(id string, ctx context.Context, res http.ResponseWriter, req *http.Request) {
	deviceConn, err := r.connman.Dial(ctx, id)
	if err != nil {
		return
	}

	if err := req.Write(deviceConn); err != nil {
		return
	}

	resp, err := http.ReadResponse(bufio.NewReader(deviceConn), req)
	if err != nil {
		return
	}

	ProxyResponseFromDevice(res, resp)
}
