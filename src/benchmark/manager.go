package benchmark

const BenchOverload = "overload"

type Manager struct {

}

var manager *Manager

func GetManager() (*Manager, error) {
	if manager == nil {
		manager = new(Manager)
	}

	return manager, nil
}

func (m *Manager) GetTest(name string) interface{} {
	switch name {
	case BenchOverload:
		return overloadManager
	}
	return nil
}
