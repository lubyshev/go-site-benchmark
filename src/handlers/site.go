package handlers

import (
	"errors"
	"fmt"
	"log"
	"lubyshev/go-site-benchmark/src/benchmark"
	"lubyshev/go-site-benchmark/src/conf"
	"lubyshev/go-site-benchmark/src/dataProvider"
	"net/http"
	"sort"
	"strings"
)

func Site(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in handlers.Site()", r)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintf(w, "Internal error: %v", r)
		}

	}()
	log.Printf("START REQUEST FROM: %s\n", req.RemoteAddr)
	searchPhrase := strings.Trim(req.FormValue("search"), " ")
	if searchPhrase == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "Empty search param")
		return
	}

	sites, err := dataProvider.GetAdapter(dataProvider.DataProviderYandex).GetData(searchPhrase)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "Yandex search failed: %s", err.Error())
		return
	}

	test := benchmark.GetManager().GetTest(benchmark.BenchOverload).(benchmark.OverloadTest)

	result, err := test.Benchmark(sites, conf.GetConfig().CacheTtl)
	if err != nil || result == nil {
		if err == nil {
			err = errors.New("unexpected error")
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "bencmark failed: %s", err.Error())
		return
	}

	keys := make([]string, 0, len(result.Items))
	for k := range result.Items {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, hostName := range keys {
		count := 0
		for _, url := range result.Items[hostName].Urls {
			count += url.Count
		}
		_, _ = fmt.Fprintf(w, "%3d: %s\n", count, hostName)
	}
	log.Printf("FINISH REQUEST FROM: %s\n===\n", req.RemoteAddr)
}
