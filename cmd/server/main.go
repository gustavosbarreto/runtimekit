package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gustavosbarreto/revwebsocketdial/pkg/reverse"
)

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
