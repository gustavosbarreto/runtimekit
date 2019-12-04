package main

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/gustavosbarreto/revwebsocketdial/pkg/reverse"
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

func main() {
	rev := reverse.NewReverse(reverse.DefaultConnectionURL, reverse.DefaultRevdialURL)
	router := rev.Router().(*mux.Router)

	router.HandleFunc("/go", func(w http.ResponseWriter, r *http.Request) {
		req, _ := http.NewRequest(
			"GET", "/merda",
			nil,
		)

		rev.ProxyRequest("merda", r.Context(), w, req)
	})

	http.ListenAndServe(":1313", router)
}
