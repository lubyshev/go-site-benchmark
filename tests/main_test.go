package tests

import (
	"context"
	"log"
	"lubyshev/go-site-benchmark/src/cache"
	"lubyshev/go-site-benchmark/src/conf"
	"os"
	"sync"
	"testing"
)

var config *conf.TestConfig
var onceConfig sync.Once

func getConfig() *conf.TestConfig {
	onceConfig.Do(func() {
		config = conf.GetTestConfig()
	})
	return config
}

func TestMain(m *testing.M) {
	ctxCache, ctxCacheCancelFunc := context.WithCancel(context.Background())
	err := cache.GetCache().StartBackground(ctxCache, getConfig().CacheBgFrequency, false)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err.Error())
	}
	defer ctxCacheCancelFunc()
	os.Exit(m.Run())
}
