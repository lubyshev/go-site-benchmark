package handlers

import (
	"errors"
	"fmt"
	"lubyshev/go-site-benchmark/src/benchmark"
	"lubyshev/go-site-benchmark/src/yandex"
	"net/http"
	"strings"
)

func Site(w http.ResponseWriter, req *http.Request) {
	searchPhrase := strings.Trim(req.FormValue("search"), " ")
	if searchPhrase == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "Empty search param")
		return
	}

	sites, err := yandex.GetYandexSearchResult(searchPhrase)
	if err != nil || sites == nil {
		if err == nil {
			err = errors.New("unexpected error")
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "Yandex search failed: %s", err.Error())
		return
	}

	test := benchmark.GetManager().GetTest(benchmark.BenchOverload).(benchmark.OverloadTest)

	result, err := test.Benchmark(sites)
	if err != nil || result == nil {
		if err == nil {
			err = errors.New("unexpected error")
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "bencmark failed: %s", err.Error())
		return
	}

	for host, count := range result.Items {
		_, _ = fmt.Fprintf(w, "%3d: %s\n", count, host)
	}
}
