package main

import (
	"fmt"
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

		res, err := rev.SendRequest(r.Context(), "merda", req)
		fmt.Println(err)
		rev.CopyResponse(res, w)
	})

	http.ListenAndServe(":1313", router)
}
