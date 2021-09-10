package benchmark

import (
	"bytes"
	"lubyshev/go-site-benchmark/src/conf"
	"lubyshev/go-site-benchmark/src/yandex"
	"net/http"
	"sync"
	"time"
)

type OverloadTest interface {
	Benchmark(sites *yandex.ResponseStruct) (*OverloadTestResult, error)
}

type OverloadSiteResult struct {
	Host        string
	Connections int
}

type OverloadTestResult struct {
	Items []OverloadSiteResult
	lock  sync.Mutex
}

func (otr *OverloadTestResult) push(host string, count int) {
	defer otr.lock.Unlock()
	otr.lock.Lock()

	otr.Items = append(otr.Items, OverloadSiteResult{
		Host:        host,
		Connections: count,
	})
}

type overload struct {
	started bool
	config  *conf.AppConfig
}

var overloadManager overload

func (o *overload) init() error {
	config, err := conf.GetConfig()
	if err != nil {
		return err
	}
	o.config = config
	return nil
}

func (o overload) Benchmark(sites *yandex.ResponseStruct) (*OverloadTestResult, error) {
	if !o.started {
		err := o.init()
		if err != nil {
			return nil, err
		}
	}

	result := new(OverloadTestResult)

	var wg sync.WaitGroup
	for _, item := range sites.Items {
		wg.Add(1)
		go o.testSite(item.Host, item.Url, result, &wg)
	}
	wg.Wait()

	return result, nil
}

func (o *overload) testSite(host string, url string, result *OverloadTestResult, wg *sync.WaitGroup) {
	defer wg.Done()
	maxConnections := o.config.MaxConnections
	tm := time.Now()

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = maxConnections*100
	t.MaxConnsPerHost = maxConnections
	t.MaxIdleConnsPerHost = maxConnections


	httpClient := &http.Client{
		Timeout:   3 * time.Second,
		Transport: t,
	}

	var resp []*http.Response
	defer func() {
		for _, item := range resp {
			_ = item.Body.Close()
		}
	}()

	for i := 0; i < maxConnections; i++ {
		if time.Since(tm) > 20*time.Second {
			break
		}
		req, err := http.NewRequest(http.MethodGet, url, bytes.NewBuffer([]byte("")))
		if err != nil {
			break
		}
		response, err := httpClient.Do(req)
		if err != nil {
			break
		}
		if response.StatusCode != http.StatusOK {
			break
		}
		resp = append(resp, response)
	}
	count := len(resp)

	result.push(host, count)
}
