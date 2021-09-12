package benchmark

import (
	"log"
	"lubyshev/go-site-benchmark/src/dataProvider"
	"lubyshev/go-site-benchmark/src/sockets"
	"net"
	"sync"
	"time"
)

type OverloadTest interface {
	Benchmark(sites *dataProvider.HostsToCheck) (*OverloadTestResult, error)
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

func (o overload) Benchmark(sites *dataProvider.HostsToCheck) (*OverloadTestResult, error) {
	result := new(OverloadTestResult)
	result.Items = make(map[string]int)

	var wg sync.WaitGroup
	for host, isSecure := range sites.Items {
		if _, ok := result.Items[host]; !ok {
			result.Items[host] = 0
			wg.Add(1)
			go o.testSite(host, isSecure, result, &wg)
		}
	}
	wg.Wait()

	return result, nil
}

func (o *overload) testSite(host string, isSecure bool, result *OverloadTestResult, wg *sync.WaitGroup) {
	var cons []net.Conn
	defer func() {
		for _, conn := range cons {
			_ = conn.Close()
		}
		log.Printf("finish %s connection\n", host)
		wg.Done()
	}()
	log.Printf("start %s connection\n", host)

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
