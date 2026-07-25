package create

import (
	"strings"
	"testing"
)

// The --set flag is wired through internal/cli/scaffold and exercised end to
// end by the cmd/project integration tests; what is verified here is the
// validation rule set itself, which lives in this package.
func TestParseSetValues_RejectsMalformedDuplicateAndReservedKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setValues []string
		wantErr   string
	}{
		{name: "missing equals separator", setValues: []string{"author"}, wantErr: "must be key=value"},
		{name: "empty key", setValues: []string{"=alice"}, wantErr: "key must not be empty"},
		{name: "duplicate key", setValues: []string{"author=alice", "author=bob"}, wantErr: "specified more than once"},
		{name: "reserved lang key", setValues: []string{"lang=go"}, wantErr: "reserved for command options"},
		{name: "reserved module path key", setValues: []string{"module_path=example.com/demo"}, wantErr: "reserved for command options"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewService().parseSetValues(Flags{SetValues: tt.setValues})
			if err == nil {
				t.Fatal("parseSetValues() expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("parseSetValues() error = %v, want contains %q", err, tt.wantErr)
			}
		})
	}
}
