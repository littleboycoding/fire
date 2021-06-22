package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()

	println("Discoverable on port 2001")

	router.GET("/", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		h := w.Header()
		h.Add("Content-Type", "application/json")
		fmt.Fprint(w, `{ "device": "littlelaptop", "accept": true, "host": "192.168.1.124", "port": "2001" }`)
	})

	log.Fatal(http.ListenAndServe(":2001", router))
}
