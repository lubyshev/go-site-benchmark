package benchmarks

import (
	"fmt"
	"io/ioutil"
	"log"
	"lubyshev/go-site-benchmark/src/conf"
	"math/rand"
	"net/http"
	"net/url"
	"testing"
)

var queries = [...]string{
	"купить самбуку",
	"купить playstation",
	"купить xbox",
	"купить ириску",
}

func BenchmarkHiLoad(b *testing.B) {
	u := fmt.Sprintf(
		"http://localhost:%d/sites?search=%s",
		conf.GetConfig().ServerPort,
		url.QueryEscape(queries[rand.Intn(4)]),
	)
	log.Printf("load url: %s", u)
	resp, err := http.Get(u)
	if err != nil {
		log.Printf("error: %s", err.Error())
	} else {
		if resp.StatusCode != http.StatusOK {
			log.Printf("http error: %d", resp.StatusCode)
		} else {
			_, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("read error: %s", err.Error())
			}
		}
	}
}
