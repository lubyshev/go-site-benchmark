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
	"os"
	"os/signal"
	"syscall"
	"time"
)

func sites(w http.ResponseWriter, req *http.Request) {
	handlers.Site(w, req)
}

var pid int

type logWriter struct{}

func (writer logWriter) Write(bytes []byte) (int, error) {
	tm := time.Now().UTC().Format("2006-01-02 15:04:05.999")
	if len([]rune(tm)) < 23 {
		tm += "0"
	}
	return fmt.Printf("[%d] [%s] %s", pid, tm, string(bytes))
}

func main() {
	pid = os.Getpid()
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
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

	ctxCache, ctxCacheCancelFunc := context.WithCancel(context.Background())
	err = cache.GetCache().StartBackground(ctxCache, config.CacheBgFrequency, config.CacheDebug)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}

	log.Printf("Listen on http://localhost:%d with config %+v", config.ServerPort, config)

	signals := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	server := &http.Server{Addr: fmt.Sprintf(":%d", config.ServerPort)}
	http.HandleFunc("/sites", sites)
	go func() {
		err = server.ListenAndServe()
		if err != nil {
			log.Printf("error listen & serve: %s", err.Error())
			done <- true
		}
	}()

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		sig := <-signals
		fmt.Println()
		log.Printf("got signal: %s", sig)
		done <- true
	}()

	<-done
	overload.StopBackground()
	ctxCacheCancelFunc()
	_ = server.Close()
	time.Sleep(1 * time.Second)
	log.Println("Stop background ...")
}
