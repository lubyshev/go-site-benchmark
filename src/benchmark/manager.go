package benchmark

const BenchOverload = "overload"

type Manager struct{}

var manager *Manager

func GetManager() *Manager {
	return manager
}

func (m *Manager) GetTest(name string) interface{} {
	switch name {
	case BenchOverload:
		return overloadManager
	}
	return nil
}
