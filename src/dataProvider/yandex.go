package dataProvider

import (
	yandex2 "lubyshev/go-site-benchmark/src/yandex"
	"sort"
)

var yandexProvider yandex

type yandex struct{}

func (y yandex) GetData(query string) (*HostsToCheck, error) {
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
	return &HostsToCheck{
		Items: result,
	}, nil
}
