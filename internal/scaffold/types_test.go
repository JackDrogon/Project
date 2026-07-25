package scaffold

import "testing"

func TestCreateRequestDestinationDir(t *testing.T) {
	t.Parallel()

	t.Run("uses target dir when set", func(t *testing.T) {
		req := CreateRequest{ProjectName: "demo", TargetDir: "workspace/output"}
		if got := req.DestinationDir(); got != "workspace/output" {
			t.Fatalf("DestinationDir() = %q, want %q", got, "workspace/output")
		}
	})

	t.Run("falls back to project name", func(t *testing.T) {
		req := CreateRequest{ProjectName: "demo"}
		if got := req.DestinationDir(); got != "demo" {
			t.Fatalf("DestinationDir() = %q, want %q", got, "demo")
		}
	})
}

func TestDryRunActionKinds(t *testing.T) {
	t.Parallel()

	got := []DryRunActionKind{DryRunActionCreateDir, DryRunActionRenderFile, DryRunActionCopyFile}
	want := []DryRunActionKind{"create_dir", "render_file", "copy_file"}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("DryRunActionKind[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
