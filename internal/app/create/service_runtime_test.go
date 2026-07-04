package create

import (
	"strings"
	"testing"
)

func TestParseSetValues(t *testing.T) {
	t.Run("rejects reserved keys including module aliases", func(t *testing.T) {
		for _, key := range []string{"lang", "project_name", "target_dir", "module", "module_path", "git_mode", "signoff", "force", "dry_run"} {
			_, err := NewService().ParseSetValues(Flags{SetValues: []string{key + "=value"}})
			if err == nil || !strings.Contains(err.Error(), "reserved for command options") {
				t.Fatalf("ParseSetValues(--set %s=value) error = %v, want reserved-key rejection", key, err)
			}
		}
	})

	t.Run("rejects malformed and duplicate entries", func(t *testing.T) {
		for _, tc := range []struct {
			values []string
			want   string
		}{
			{[]string{"missing-separator"}, "must be key=value"},
			{[]string{"=value"}, "key must not be empty"},
			{[]string{"author=a", "author=b"}, "specified more than once"},
		} {
			_, err := NewService().ParseSetValues(Flags{SetValues: tc.values})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("ParseSetValues(%v) error = %v, want %q", tc.values, err, tc.want)
			}
		}
	})

	t.Run("accepts template input keys", func(t *testing.T) {
		values, err := NewService().ParseSetValues(Flags{SetValues: []string{"author=alice", "go_version=1.25"}})
		if err != nil {
			t.Fatalf("ParseSetValues() error = %v", err)
		}
		if values["author"] != "alice" || values["go_version"] != "1.25" {
			t.Fatalf("ParseSetValues() = %#v, want author/go_version preserved", values)
		}
	})
}
