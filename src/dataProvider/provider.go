package dataProvider

const DataProviderYandex = "yandex"

type OverloadSitesToCheck interface {
	GetData(query string) (*HostsToCheck, error)
}

type HostsToCheck struct {
	// [host]isSecure
	Items map[string]bool
}

func GetAdapter(name string) OverloadSitesToCheck {
	switch name {
	case DataProviderYandex:
		return yandexProvider
	}

	return nil
}
