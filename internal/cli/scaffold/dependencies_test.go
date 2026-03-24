package scaffold

import (
	"testing"

	appcreate "github.com/JackDrogon/project/internal/app/create"
)

func TestDependenciesRequireCreator(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("recover() = nil, want panic")
		}
	}()

	Dependencies{}.creator()
}

func TestDependenciesRequireNewService(t *testing.T) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("recover() = nil, want panic")
		}
	}()

	Dependencies{Creator: &appcreate.Creator{}}.newService()
}
