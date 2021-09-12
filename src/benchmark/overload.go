package benchmark

import (
	"log"
	"lubyshev/go-site-benchmark/src/sockets"
	"lubyshev/go-site-benchmark/src/yandex"
	"net"
	"sync"
	"time"
)

type OverloadTest interface {
	Benchmark(sites *yandex.ResponseStruct) (*OverloadTestResult, error)
}

type OverloadTestResult struct {
	Items map[string]int
	lock  sync.Mutex
}

func (otr *OverloadTestResult) set(host string, count int) {
	defer otr.lock.Unlock()
	otr.lock.Lock()
	otr.Items[host] = count
}

type overload struct{}

var overloadManager overload

func (o overload) Benchmark(sites *yandex.ResponseStruct) (*OverloadTestResult, error) {
	result := new(OverloadTestResult)
	result.Items = make(map[string]int)

	var wg sync.WaitGroup
	for _, item := range sites.Items {
		if _, ok := result.Items[item.Host]; !ok {
			result.Items[item.Host] = 0
			wg.Add(1)
			go o.testSite(item.Host, item.Url, result, &wg)
		}
	}
	wg.Wait()

	return result, nil
}

func (o *overload) testSite(host string, url string, result *OverloadTestResult, wg *sync.WaitGroup) {
	var cons []net.Conn
	defer func() {
		for _, conn := range cons {
			_ = conn.Close()
		}
		log.Printf("finish %s connection\n", host)
		wg.Done()
	}()
	log.Printf("start %s connection\n", host)

	isSecure := url[0:5] == "https"
	addr, err := net.LookupIP(host)
	if err != nil {
		return
	}
	ip := addr[0].String()

	tm := time.Now()
	for {
		if time.Since(tm) > 20*time.Second {
			break
		}
		conn, err := sockets.GetSocketsManager().GetHttpConnection(ip, isSecure)
		if err != nil {
			log.Printf("ERROR %s: connection error: %s\n", host, err.Error())
			break
		}
		cons = append(cons, conn)
	}

	result.set(host, len(cons))
}
