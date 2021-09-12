package main

import (
	"fmt"
	"log"
	"lubyshev/go-site-benchmark/src/conf"
	"lubyshev/go-site-benchmark/src/handlers"
	"net/http"
)

func sites(w http.ResponseWriter, req *http.Request) {
	handlers.Site(w, req)
}

func main() {
	config := conf.GetConfig()
	log.Printf("Listen on http://localhost:%d", config.ServerPort)
	log.Printf("With config: %+v\n", *config)

	http.HandleFunc("/sites", sites)
	_ = http.ListenAndServe(fmt.Sprintf(":%d", config.ServerPort), nil)
}
