package benchmark

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"lubyshev/go-site-benchmark/src/cache"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	stateQueueStarted = "started"
	stateQueueStopped = "stopped"
)

var (
	ErrAlreadyStarted = errors.New("already started")
)

type overloadQueue struct {
	state                string
	urls                 []*Url
	mxUrls               sync.Mutex
	chUrls               chan *Url
	ctx                  context.Context
	cancel               context.CancelFunc
	workersCount         int
	initConnectionsCount int
	maxLimit             int
	maxConnections       int
}

func (q *overloadQueue) start(count int, connectionsCount int, limit int, connections int) error {
	if q.state == stateQueueStarted {
		return ErrAlreadyStarted
	}
	q.state = stateQueueStarted
	q.workersCount = count
	q.initConnectionsCount = connectionsCount
	q.maxLimit = limit
	q.maxConnections = connections
	q.ctx, q.cancel = context.WithCancel(context.Background())

	go q._pusher(q.ctx)
	go q._start()

	return nil
}

func (q *overloadQueue) _start() {
	wg := sync.WaitGroup{}
	wg.Add(q.workersCount)
	for i := 0; i < q.workersCount; i++ {
		go q.worker(i, q.chUrls, q.ctx, &wg)
	}
	wg.Wait()
	q.state = stateQueueStarted
}

func (q *overloadQueue) stop() {
	q.cancel()
}

func (q *overloadQueue) worker(
	i int,
	chUrls <-chan *Url,
	ctx context.Context,
	wg *sync.WaitGroup,
) {
	defer func() {
		log.Printf("Finish overload queue worker %2d\n", i)
		wg.Done()
	}()
	log.Printf("Start overload queue worker %2d\n", i)
	for {
		select {
		case url := <-chUrls:
			q.testUrl(i, url)
			break
		case <-ctx.Done():
			return
		}
	}
}

var connectionCount int32

func (q *overloadQueue) testUrl(_ int, url *Url) {
	if url.state != stateUrlInProgress {
		return
	}
	if url.errors >= 0 {
		if url.Count == 0 && url.currentCount == 0 {
			url.currentCount = q.initConnectionsCount
		} else {
			if url.errors > 0 {
				url.Count = url.currentCount - url.errors
				if url.Count < 0 {
					url.Count = 0
				}
				url.state = stateUrlReady
			} else {
				url.Count = url.currentCount
				url.currentCount *= 2
				if url.Count >= q.maxLimit {
					url.state = stateUrlReady
				}
				if url.currentCount > q.maxConnections {
					url.currentCount = q.maxConnections
				}
			}

			if url.state == stateUrlReady {
				cache.GetCache().Set(url.Url, url, url.ttl)
				log.Printf("url tested: %s", url.Url)
				return
			}
			url.errors = -1
		}
	}

	errorsCount := int32(0)
	if int(connectionCount) < (q.maxConnections - url.currentCount) {
		wg := sync.WaitGroup{}
		for i := 0; i < url.currentCount; i++ {
			wg.Add(1)
			go q.loadUrl(url.Url, &errorsCount, &wg)
		}
		wg.Wait()
	} else {
		q.pushForced(url)
		return
	}

	log.Printf("%s tested on %d connections and has %d errors",
		url.Url,
		url.currentCount,
		errorsCount,
	)

	if int(errorsCount) == url.currentCount {
		url.Count = 0
		url.state = stateUrlFailed
	}
	url.errors = int(errorsCount)
	cache.GetCache().Set(url.Url, url, url.ttl)
	time.Sleep(200 * time.Millisecond)
	if url.state == stateUrlInProgress {
		q.pushForced(url)
	}
}

func (q *overloadQueue) push(url *Url) {
	if cache.GetCache().Exists(url.Url) {
		return
	}
	cache.GetCache().Set(url.Url, url, url.ttl)
	q.pushForced(url)
	log.Printf("url pushed to queue: %v", url)
}

func (q *overloadQueue) pushForced(url *Url) {
	defer q.mxUrls.Unlock()
	q.mxUrls.Lock()
	q.urls = append(q.urls, url)
}

func (q *overloadQueue) loadUrl(url string, errorsCount *int32, wg *sync.WaitGroup) {
	defer func() {
		atomic.AddInt32(&connectionCount, -1)
		wg.Done()
	}()
	atomic.AddInt32(&connectionCount, 1)

	var dialer = &net.Dialer{
		Timeout:  3 * time.Second,
		Deadline: time.Now().Add(4 * time.Second),
	}
	client := http.Client{Transport: &http.Transport{
		DialContext:         dialer.DialContext,
		TLSHandshakeTimeout: 2 * time.Second,
		MaxConnsPerHost:     q.maxConnections * 2,
		MaxIdleConns:        q.maxConnections * 4,
	}}
	resp, err := client.Get(url)
	if err != nil {
		atomic.AddInt32(errorsCount, 1)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		atomic.AddInt32(errorsCount, 1)
	} else {
		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			atomic.AddInt32(errorsCount, 1)
		}
	}
}

func (q *overloadQueue) _pusher(ctx context.Context) {
	for {
		select {
		case <-time.After(10 * time.Millisecond):
			var tmp *Url
			q.mxUrls.Lock()
			l := len(q.urls)
			if l > 0 {
				tmp, q.urls[0] = q.urls[0], nil
				q.urls = q.urls[1:l]
			}
			q.mxUrls.Unlock()
			if tmp != nil {
				q.chUrls <- tmp
			}
		case <-ctx.Done():
			return
		}
	}
}

func (q *overloadQueue) getUrl(url string) (*Url, error) {
	c := cache.GetCache()
	if !c.Exists(url) {
		return nil, cache.ErrNotExists
	}
	c.RLock()
	defer c.RUnlock()
	cachedUrl, err := c.GetRaw(url)
	if err != nil {
		return nil, err
	}

	return &Url{
		Url:   cachedUrl.(*Url).Url,
		Count: cachedUrl.(*Url).Count,
	}, nil
}

var overloadBg *overloadQueue
var overloadBgOnce sync.Once

func (o overload) StartBackground(
	workersCount int,
	initConnectionsCount int,
	maxLimit int,
	maxConnections int,
) error {
	return getQueue().start(workersCount, initConnectionsCount, maxLimit, maxConnections)
}

func (o overload) StopBackground() {
	getQueue().stop()
}

func getQueue() *overloadQueue {
	overloadBgOnce.Do(func() {
		overloadBg = new(overloadQueue)
		overloadBg.state = stateQueueStopped
		overloadBg.chUrls = make(chan *Url)
		overloadBg.urls = make([]*Url, 0)
	})
	return overloadBg
}
