package dataProvider

import (
	"lubyshev/go-site-benchmark/src/cache"
	yandex2 "lubyshev/go-site-benchmark/src/yandex"
	"sort"
	"time"
)

var yandexProvider yandex

type yandex struct{}

func (y yandex) GetData(query string) (*HostsToCheck, error) {
	iRes, err := cache.GetCache().Get("yandex::" + query)
	if err == nil {
		res := iRes.(*HostsToCheck)
		return res, nil
	}
	data, err := yandex2.GetYandexSearchResult(query)
	if err != nil {
		return nil, err
	}
	result := make(map[string][]string)
	for _, item := range data.Items {
		if _, ok := result[item.Host]; !ok {
			result[item.Host] = make([]string, 0)
		}
		l := len(result[item.Host])
		sort.Strings(result[item.Host])
		i := sort.Search(l, func(i int) bool { return result[item.Host][i] >= item.Url })
		if i >= len(result[item.Host]) || result[item.Host][i] != item.Url {
			result[item.Host] = append(result[item.Host], item.Url)
		}
	}
	res := &HostsToCheck{
		Items: result,
	}
	cache.GetCache().Set("yandex::"+query, res, 600*time.Second)

	return res, nil
}
