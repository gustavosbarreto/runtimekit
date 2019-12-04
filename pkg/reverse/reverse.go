package reverse

import (
	"bufio"
	"context"
	"errors"
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

const (
	DefaultConnectionURL = "/connection"
	DefaultRevdialURL    = "/revdial"
)

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
			id := r.Header.Get("X-CLIENT-ID")
			if id == "" {
				return id, errors.New("X-CLIENT-ID header is missing")
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

func (r *Reverse) SendRequest(ctx context.Context, id string, req *http.Request) (*http.Response, error) {
	conn, err := r.connman.Dial(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := req.Write(conn); err != nil {
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (r *Reverse) CopyResponse(resp *http.Response, w http.ResponseWriter) {
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	resp.Body.Close()
}
