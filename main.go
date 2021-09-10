package main

import (
	"fmt"
	"lubyshev/go-site-benchmark/src/conf"
	"lubyshev/go-site-benchmark/src/handlers"
	"net/http"
)

func sites(w http.ResponseWriter, req *http.Request) {
	handlers.Site(w, req)
}

func main() {
	fmt.Println("Listen on http://localhost:8090")
	config, _ := conf.GetConfig()
	fmt.Printf("With config: %+v\n", *config)

	http.HandleFunc("/sites", sites)
	_ = http.ListenAndServe(fmt.Sprintf(":%d", config.ServerPort), nil)
}
