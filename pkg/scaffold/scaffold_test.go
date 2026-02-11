package scaffold

import (
	"bytes"
	"testing"
	"testing/fstest"
)

func TestListLangs(t *testing.T) {
	fsys := fstest.MapFS{
		"go/Makefile":  {Data: []byte("build:")},
		"cpp/Makefile": {Data: []byte("build:")},
	}

	c := NewCreator(fsys, &bytes.Buffer{})
	langs, err := c.ListLangs()
	if err != nil {
		t.Fatalf("ListLangs() error = %v", err)
	}

	if len(langs) != 2 {
		t.Fatalf("ListLangs() returned %d languages, want 2", len(langs))
	}

	// fstest.MapFS returns entries sorted alphabetically
	want := []string{"cpp", "go"}
	for i, got := range langs {
		if got != want[i] {
			t.Errorf("ListLangs()[%d] = %q, want %q", i, got, want[i])
		}
	}
}
