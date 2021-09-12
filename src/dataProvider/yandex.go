package dataProvider

import (
	yandex2 "lubyshev/go-site-benchmark/src/yandex"
)

var yandexProvider yandex

type yandex struct{}

func (y yandex) GetData(query string) (*HostsToCheck, error) {
	data, err := yandex2.GetYandexSearchResult(query)
	if err != nil {
		return nil, err
	}
	result := make(map[string]bool)
	for _, item := range data.Items {
		result[item.Host] = item.Url[0:5] == "https"
	}
	return &HostsToCheck{
		Items: result,
	}, nil
}
