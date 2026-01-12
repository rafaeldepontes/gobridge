package main

import (
	"log"
	"net/http"

	"github.com/rafaeldepontes/gobridge/internal/proxy"
)

func main() {
	rp := proxy.NewReverseProxy()

	http.Handle("/", rp)

	log.Fatalln(http.ListenAndServe(":8080", nil))
}
