package templatefs

import "testing"

func TestNew(t *testing.T) {
	adapter := New()
	if adapter == nil {
		t.Fatal("New() = nil")
	}
}
