package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/gustavosbarreto/httptunnel/pkg/revdial"
	"github.com/gustavosbarreto/httptunnel/pkg/wsconnadapter"
)

func Revdial(ctx context.Context, path string) (*websocket.Conn, *http.Response, error) {
	fmt.Println(path)
	return websocket.DefaultDialer.Dial(strings.Join([]string{"ws://localhost:1313", path}, ""), nil)
}

func main() {
	router := mux.NewRouter()

	req, _ := http.NewRequest("", "", nil)
	req.Header.Set("X-CLIENT-ID", "merda")

	wsConn, _, _ := websocket.DefaultDialer.Dial("ws://localhost:1313/connection", req.Header)

	conn := wsconnadapter.New(wsConn)

	listener := revdial.NewListener(conn, func(ctx context.Context, path string) (*websocket.Conn, *http.Response, error) {
		fmt.Println("listener")
		return Revdial(ctx, path)
	})

	router.HandleFunc("/merda", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("/merda")
		w.WriteHeader(http.StatusOK)

		io.WriteString(w, `{"alive": true}`)
	})

	sv := http.Server{
		Handler: router,
	}

	sv.Serve(listener)
}
