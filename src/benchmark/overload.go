package benchmark

import (
	"log"
	"lubyshev/go-site-benchmark/src/conf"
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
	result.Items = make(map[string]int)

	var wg sync.WaitGroup
	for _, item := range sites.Items {
		if _, ok := result.Items[item.Host]; !ok {
			wg.Add(1)
			go o.testSite(item.Host, item.Url, result, &wg)
		}
	}

	wg.Wait()

	return result, nil
}

func (o *overload) testSite(host string, url string, result *OverloadTestResult, wg *sync.WaitGroup) {
	var conns []net.Conn

	defer func() {
		for _, conn := range conns {
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
		var (
			conn net.Conn
			err  error
		)
		if time.Since(tm) > 20*time.Second {
			break
		}
		if isSecure {
			conn, err = sockets.GetHttpsConnection(ip)
		} else {
			conn, err = sockets.GetHttpConnection(ip)
		}
		if err != nil {
			break
		}
		conns = append(conns, conn)
	}
	count := len(conns)

	result.set(host, count)
}
