package benchmark

import (
	"fmt"
	"log"
	"lubyshev/go-site-benchmark/src/cache"
	"lubyshev/go-site-benchmark/src/dataProvider"
	"sync"
	"time"
)

type OverloadTest interface {
	Benchmark(sites *dataProvider.HostsToCheck, ttl time.Duration) (map[string]int, error)
	StartBackground(
		workersCount int,
		initConnectionsCount int,
		maxLimit int,
		maxConnections int,
		method string,
	) error
	StopBackground()
}

const (
	stateUrlInProgress = "in progress"
	stateUrlReady      = "ready"
	stateUrlFailed     = "failed"
)

type Url struct {
	Url      string
	Count    int
	state    string
	ttl      time.Duration
	attempts int
	errors   int
}

type Host struct {
	Urls map[string]*Url
}

type OverloadTestResult struct {
	Items map[string]*Host
	lock  sync.RWMutex
}

func (otr *OverloadTestResult) set(host string, url *Url) error {
	defer otr.lock.Unlock()
	otr.lock.Lock()
	if _, ok := otr.Items[host]; !ok {
		return fmt.Errorf("host %s is not initialized", host)
	}
	if otr.Items[host].Urls == nil {
		otr.Items[host].Urls = make(map[string]*Url, 0)
	}
	otr.Items[host].Urls[url.Url] = url

	return nil
}

func (otr *OverloadTestResult) clone() *OverloadTestResult {
	defer otr.lock.Unlock()
	otr.lock.Lock()
	clone := new(OverloadTestResult)
	if otr.Items != nil {
		clone.Items = make(map[string]*Host)
		for hostName, host := range otr.Items {
			clone.Items[hostName] = new(Host)
			if otr.Items[hostName].Urls != nil {
				urls := make(map[string]*Url)
				for urlName, url := range host.Urls {
					urls[urlName] = &Url{
						Url:   url.Url,
						Count: url.Count,
					}
				}
				clone.Items[hostName].Urls = urls
			}
		}
	}

	return clone
}

type overload struct{}

var overloadManager overload

func (o overload) Benchmark(
	sites *dataProvider.HostsToCheck,
	ttl time.Duration,
) (res map[string]int, err error) {
	result := new(OverloadTestResult)
	result.Items = make(map[string]*Host)

	var wg sync.WaitGroup
	for host, url := range sites.Items {
		if _, ok := result.Items[host]; !ok {
			result.lock.Lock()
			result.Items[host] = new(Host)
			result.lock.Unlock()
			wg.Add(1)
			go o.testSite(host, url, ttl, result, &wg)
		}
	}
	wg.Wait()
	res = make(map[string]int)
	for hostName, host := range result.Items {
		res[hostName] = 0
		if l := len(host.Urls); l > 0 {
			for _, url := range host.Urls {
				res[hostName] += url.Count
			}
			res[hostName] /= l
		}
	}

	return
}

func (o *overload) testSite(
	host string,
	urls []string,
	ttl time.Duration,
	result *OverloadTestResult,
	wg *sync.WaitGroup,
) {
	defer func() {
		wg.Done()
	}()

	for _, url := range urls {
		cachedUrl, err := getQueue().getUrl(url)
		if err == cache.ErrNotExists {
			// move to queue
			getQueue().push(&Url{
				state: stateUrlInProgress,
				ttl:   ttl,
				Url:   url,
			})
			continue
		}
		if err != nil {
			log.Printf("ERROR: %s\n", err.Error())
			continue
		}

		err = result.set(host, cachedUrl)
		if err != nil {
			log.Printf("ERROR: %s\n", err.Error())
		}
	}
}
