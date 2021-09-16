package main

import (
	"context"
	"fmt"
	"log"
	"lubyshev/go-site-benchmark/src/benchmark"
	"lubyshev/go-site-benchmark/src/cache"
	"lubyshev/go-site-benchmark/src/conf"
	"lubyshev/go-site-benchmark/src/handlers"
	"net/http"
)

func sites(w http.ResponseWriter, req *http.Request) {
	handlers.Site(w, req)
}

func main() {
	config := conf.GetConfig()
	log.Println("Starting background ...")

	overload := benchmark.GetManager().GetTest(benchmark.BenchOverload).(benchmark.OverloadTest)
	err := overload.StartBackground(
		config.OverloadWorkers,
		config.OverloadInitConnections,
		config.OverloadMaxLimit,
		config.OverloadMaxConnections,
	)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}
	defer func() {
		overload.StopBackground()
	}()

	ctxCache, ctxCacheCancelFunc := context.WithCancel(context.Background())
	cache.GetCache().StartBackground(ctxCache)
	defer ctxCacheCancelFunc()

	log.Printf("Listen on http://localhost:%d", config.ServerPort)
	log.Printf("With config: %+v\n", *config)
	http.HandleFunc("/sites", sites)
	_ = http.ListenAndServe(fmt.Sprintf(":%d", config.ServerPort), nil)
	log.Println("Stop background ...")
}
