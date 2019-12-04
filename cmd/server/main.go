package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"

	"github.com/function61/holepunch-server/pkg/wsconnadapter"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/gustavosbarreto/revwebsocketdial/pkg/connman"
	"github.com/gustavosbarreto/revwebsocketdial/pkg/revdial"
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
	manager := connman.New()
	router := mux.NewRouter()

	router.HandleFunc("/connection", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
		}

		manager.Set("merda", wsconnadapter.New(conn))
		fmt.Println("/connection")
	}).Methods("GET")

	router.Handle("/revdial", revdial.ConnHandler(upgrader)).Methods("GET")

	router.HandleFunc("/go", func(w http.ResponseWriter, r *http.Request) {
		deviceConn, err := manager.Dial(r.Context(), "merda")
		fmt.Println(err)

		req, _ := http.NewRequest(
			"GET", "/merda",
			nil,
		)

		req.Write(deviceConn)

		resp, _ := http.ReadResponse(bufio.NewReader(deviceConn), req)

		ProxyResponseFromDevice(w, resp)
	})

	http.ListenAndServe(":1313", router)
}
