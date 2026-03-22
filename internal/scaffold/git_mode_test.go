package scaffold

import "testing"

func TestResolveGitMode(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateRequest
		want    GitMode
		wantErr bool
	}{
		{name: "defaults to init commit", req: CreateRequest{}, want: GitModeInitCommit},
		{name: "no git wins", req: CreateRequest{NoGit: true}, want: GitModeNone},
		{name: "explicit init only", req: CreateRequest{GitMode: GitModeInitOnly}, want: GitModeInitOnly},
		{name: "conflict", req: CreateRequest{NoGit: true, GitMode: GitModeInitOnly}, wantErr: true},
		{name: "invalid mode", req: CreateRequest{GitMode: GitMode("invalid")}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveGitMode(tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ResolveGitMode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("ResolveGitMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProjectNameFromGoModulePath(t *testing.T) {
	tests := []struct {
		name       string
		modulePath string
		want       string
	}{
		{name: "plain final segment", modulePath: "github.com/acme/agent-village", want: "agent-village"},
		{name: "major version uses parent segment", modulePath: "github.com/acme/agent-village/v2", want: "agent-village"},
		{name: "bare major version keeps suffix", modulePath: "v2", want: "v2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ProjectNameFromGoModulePath(tt.modulePath); got != tt.want {
				t.Fatalf("ProjectNameFromGoModulePath(%q) = %q, want %q", tt.modulePath, got, tt.want)
			}
		})
	}
}

func TestDefaultModulePath(t *testing.T) {
	req := CreateRequest{ProjectName: "demo"}
	if got := DefaultModulePath(req); got != "demo" {
		t.Fatalf("DefaultModulePath() = %q, want %q", got, "demo")
	}

	req.ModulePath = "example.com/demo"
	if got := DefaultModulePath(req); got != "example.com/demo" {
		t.Fatalf("DefaultModulePath() = %q, want %q", got, "example.com/demo")
	}
}
