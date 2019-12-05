package httptunnel

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/gustavosbarreto/httptunnel/pkg/connman"
	"github.com/gustavosbarreto/httptunnel/pkg/revdial"
	"github.com/gustavosbarreto/httptunnel/pkg/wsconnadapter"
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

type Tunnel struct {
	ConnectionPath    string
	DialerPath        string
	ConnectionHandler func(*http.Request) (string, error)
	connman           *connman.ConnectionManager
}

func NewTunnel(connectionPath, dialerPath string) *Tunnel {
	return &Tunnel{
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

func (t *Tunnel) Router() http.Handler {
	router := mux.NewRouter()

	router.HandleFunc(t.ConnectionPath, func(res http.ResponseWriter, req *http.Request) {
		conn, err := upgrader.Upgrade(res, req, nil)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		id, err := t.ConnectionHandler(req)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			defer conn.Close()
			return
		}

		t.connman.Set(id, wsconnadapter.New(conn))
	}).Methods(http.MethodGet)

	router.Handle(t.DialerPath, revdial.ConnHandler(upgrader)).Methods(http.MethodGet)

	return router
}

func (t *Tunnel) SendRequest(ctx context.Context, id string, req *http.Request) (*http.Response, error) {
	conn, err := t.connman.Dial(ctx, id)
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

func (t *Tunnel) ForwardResponse(resp *http.Response, w http.ResponseWriter) {
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	resp.Body.Close()
}
