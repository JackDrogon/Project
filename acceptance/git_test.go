//go:build acceptance

package acceptance

import (
	"path/filepath"
	"strings"
	"testing"
)

// isolatedGitEnv gives git a self-contained identity and detaches it from the
// host's global and system config, so the result is identical on a bare CI
// runner and on a developer machine.
func isolatedGitEnv() []string {
	return []string{
		"GIT_AUTHOR_NAME=Acceptance Test",
		"GIT_AUTHOR_EMAIL=acceptance@example.test",
		"GIT_COMMITTER_NAME=Acceptance Test",
		"GIT_COMMITTER_EMAIL=acceptance@example.test",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
	}
}

func TestAcceptance_GitInitCommitCreatesSingleCommit(t *testing.T) {
	requireTool(t, "git")

	gitEnv := isolatedGitEnv()
	workDir := scaffold(t, gitEnv,
		"new", "--lang", "go", "--git", "init+commit",
		"--module", "example.com/acceptgit", "acceptgit",
	)
	projectDir := filepath.Join(workDir, "acceptgit")

	requireFiles(t, projectDir, []string{".git", "go.mod", "cmd/acceptgit/main.go"})

	count := strings.TrimSpace(runTool(t, projectDir, gitEnv, "git", "rev-list", "--count", "HEAD"))
	if count != "1" {
		t.Fatalf("commit count = %q, want %q", count, "1")
	}

	subject := strings.TrimSpace(runTool(t, projectDir, gitEnv, "git", "log", "-1", "--pretty=%s"))
	if subject != "Initial commit" {
		t.Fatalf("commit subject = %q, want %q", subject, "Initial commit")
	}

	status := runTool(t, projectDir, gitEnv, "git", "status", "--porcelain")
	if strings.TrimSpace(status) != "" {
		t.Fatalf("working tree is dirty after scaffolding:\n%s", status)
	}
}

func TestAcceptance_GitSignoffAddsTrailer(t *testing.T) {
	requireTool(t, "git")

	gitEnv := isolatedGitEnv()
	workDir := scaffold(t, gitEnv,
		"new", "--lang", "go", "--git", "init+commit", "--signoff",
		"--module", "example.com/acceptsignoff", "acceptsignoff",
	)
	projectDir := filepath.Join(workDir, "acceptsignoff")

	body := runTool(t, projectDir, gitEnv, "git", "log", "-1", "--pretty=%B")
	if !strings.Contains(body, "Signed-off-by: Acceptance Test <acceptance@example.test>") {
		t.Fatalf("commit message = %q, want a Signed-off-by trailer", body)
	}
}

func TestAcceptance_GitInitOnlyLeavesNoCommit(t *testing.T) {
	requireTool(t, "git")

	gitEnv := isolatedGitEnv()
	workDir := scaffold(t, gitEnv,
		"new", "--lang", "go", "--git", "init-only",
		"--module", "example.com/acceptinitonly", "acceptinitonly",
	)
	projectDir := filepath.Join(workDir, "acceptinitonly")

	requireFiles(t, projectDir, []string{".git"})

	status := runTool(t, projectDir, gitEnv, "git", "status", "--porcelain")
	if strings.TrimSpace(status) == "" {
		t.Fatal("git status is empty, want the scaffolded files to be uncommitted")
	}
}
