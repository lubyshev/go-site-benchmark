package benchmark

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
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

const (
	methodSimple = "simple"
	methodStrong = "strong"
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
	method               string
}

func (q *overloadQueue) start(
	count int,
	connectionsCount int,
	limit int,
	connections int,
	method string,
) error {
	if q.state == stateQueueStarted {
		return ErrAlreadyStarted
	}
	switch method {
	case methodSimple:
		q.method = methodSimple
	case methodStrong:
		q.method = methodStrong
	default:
		return fmt.Errorf("invalid overload metod: %s", method)
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

var (
	connectionCount      int
	connectionCountMutex sync.Mutex
)

func (q *overloadQueue) allocateConnections(count int) bool {
	defer connectionCountMutex.Unlock()
	connectionCountMutex.Lock()
	if q.maxConnections-connectionCount < count {
		return false
	}
	connectionCount += count
	return true
}

func (q *overloadQueue) releaseConnections(count int) bool {
	defer connectionCountMutex.Unlock()
	connectionCountMutex.Lock()
	connectionCount -= count
	return true
}

func (q *overloadQueue) testUrl(_ int, url *Url) {
	if url.state != stateUrlInProgress {
		return
	}
	if url.errors >= 0 {
		url.state, url.Count, url.attempts = q.nextStep(url)
		if url.state != stateUrlInProgress {
			cache.GetCache().Set(url.Url, url, url.ttl)
			log.Printf("url tested: %s", url.Url)
			return
		}
		url.errors = -1
	}

	errorsCount := int32(0)
	if q.allocateConnections(url.attempts) {
		wg := sync.WaitGroup{}
		for i := 0; i < url.attempts; i++ {
			wg.Add(1)
			go q.loadUrl(url.Url, &errorsCount, &wg)
		}
		wg.Wait()
		q.releaseConnections(url.attempts)
	} else {
		q.pushForced(url)
		return
	}
	log.Printf(
		"%s tested on %d connections and has %d errors",
		url.Url,
		url.attempts,
		errorsCount,
	)
	url.errors = int(errorsCount)

	cache.GetCache().Set(url.Url, url, url.ttl)
	time.Sleep(20 * time.Millisecond)
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
		wg.Done()
	}()
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				Deadline:  time.Now().Add(4 * time.Second),
				KeepAlive: -1,
			}).DialContext,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			TLSHandshakeTimeout: 3 * time.Second,
		}}
	defer func() {
		client.CloseIdleConnections()
		client = nil
	}()
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("ERROR: %s %s", err.Error(), url)
		atomic.AddInt32(errorsCount, 1)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		// log.Printf("ERROR: http status %d %s", resp.StatusCode, url)
		atomic.AddInt32(errorsCount, 1)
	} else {
		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("ERROR: %s %s", err.Error(), url)
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

func (q *overloadQueue) nextStep(url *Url) (nextState string, nextCount int, nextAttempts int) {
	switch q.method {
	case methodSimple:
		return q.nextStepSimple(url.Count, url.attempts, url.errors)
	case methodStrong:
		return q.nextStepStrong(url.Count, url.attempts, url.errors)
	}
	return
}

func (q *overloadQueue) nextStepSimple(count int, attempts int, errs int) (nextState string, nextCount int, nextAttempts int) {
	nextAttempts = 0
	nextState = stateUrlInProgress
	if count == 0 && attempts == 0 {
		nextAttempts = q.initConnectionsCount
	} else {
		if errs == attempts {
			if count != q.initConnectionsCount {
				nextState, nextCount, nextAttempts = stateUrlReady, count, 0
			} else {
				nextState, nextCount, nextAttempts = stateUrlFailed, 0, 0
			}
			return
		}
		if errs > 0 {
			nextCount = attempts - errs
			if count < 0 {
				nextCount = 0
			}
			nextState = stateUrlReady
		} else {
			nextCount = attempts
			nextAttempts = attempts * 2
			if nextCount >= q.maxLimit {
				nextState = stateUrlReady
			}
			if attempts > q.maxConnections {
				nextAttempts = q.maxConnections
			}
		}
	}

	return
}

func (q *overloadQueue) nextStepStrong(count int, attempts int, errs int) (nextState string, nextCount int, nextAttempts int) {
	return
}

var overloadBg *overloadQueue
var overloadBgOnce sync.Once

func (o overload) StartBackground(
	workersCount int,
	initConnectionsCount int,
	maxLimit int,
	maxConnections int,
	method string,
) error {
	return getQueue().start(
		workersCount,
		initConnectionsCount,
		maxLimit,
		maxConnections,
		method,
	)
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
