package catalog

import "testing"

func TestDependenciesRequireNewService(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("recover() = nil, want panic")
		}
	}()

	Dependencies{}.newService()
}
