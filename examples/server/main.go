package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gustavosbarreto/httptunnel"
)

func main() {
	tunnel := httptunnel.NewTunnel(httptunnel.DefaultConnectionURL, httptunnel.DefaultRevdialURL)
	router := tunnel.Router().(*mux.Router)

	router.HandleFunc("/go", func(w http.ResponseWriter, r *http.Request) {
		req, _ := http.NewRequest(
			"GET", "/merda",
			nil,
		)

		res, err := tunnel.SendRequest(r.Context(), "merda", req)
		fmt.Println(err)
		tunnel.ForwardResponse(res, w)
	})

	http.ListenAndServe(":1313", router)
}
