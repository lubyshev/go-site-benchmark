package tests

import (
	"github.com/stretchr/testify/assert"
	"lubyshev/go-site-benchmark/src/cache"
	"lubyshev/go-site-benchmark/src/conf"
	"testing"
	"time"
)

func Test_Cache_NotExists(t *testing.T) {
	_, err := cache.GetCache().Get("blabla")
	assert.Error(t, err)
	assert.Equal(t, cache.ErrNotExists, err)
}

func Test_Cache_Ttl(t *testing.T) {
	v, err := cache.GetCache().Set("blabla", getConfig(), time.Second).Get("blabla")
	assert.NoError(t, err)
	assert.Equal(t, getConfig(), v.(*conf.TestConfig))
	time.Sleep(2 * time.Second)
	_, err = cache.GetCache().Get("blabla")
	assert.Error(t, err)
	assert.Equal(t, cache.ErrExpired, err)
	time.Sleep(2 * time.Second)
	_, err = cache.GetCache().Get("blabla")
	assert.Error(t, err)
	assert.Equal(t, cache.ErrNotExists, err)
}
