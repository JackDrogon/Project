package catalog

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func TestServiceListLangs(t *testing.T) {
	svc := NewService(fstest.MapFS{
		"go/.project-template-manifest.toml":   {Data: []byte("version = 2\nname = \"go\"\n")},
		"cpp/.project-template-manifest.toml":  {Data: []byte("version = 2\nname = \"cpp\"\n")},
		"rust/.project-template-manifest.toml": {Data: []byte("version = 2\nname = \"rust\"\n")},
	}, nil)

	got, err := svc.ListLangs()
	if err != nil {
		t.Fatalf("ListLangs() error = %v", err)
	}
	if !reflect.DeepEqual(got, []string{"cpp", "go", "rust"}) {
		t.Fatalf("ListLangs() = %v, want [cpp go rust]", got)
	}
}

func TestServiceListLangs_PropagatesReadErrors(t *testing.T) {
	svc := NewService(failingFS{err: errors.New("boom")}, nil)
	_, err := svc.ListLangs()
	if err == nil {
		t.Fatal("ListLangs() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read templates") {
		t.Fatalf("ListLangs() error = %v, want template read error", err)
	}
}
