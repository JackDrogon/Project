package create

import (
	"fmt"

	domain "github.com/JackDrogon/project/internal/scaffold"
)

const initialCommitMessage = "Initial commit"

func (c *Creator) validateGitOptions(opts Options) error {
	mode, err := domain.ResolveGitMode(opts)
	if err != nil {
		return err
	}

	if opts.Signoff && mode != domain.GitModeInitCommit {
		return fmt.Errorf("--signoff requires --git=%s", domain.GitModeInitCommit)
	}

	return nil
}

func (c *Creator) initGitRepo(opts Options) error {
	commitArgs := []string{"commit", "-m", initialCommitMessage}
	if opts.Signoff {
		commitArgs = []string{"commit", "-s", "-m", initialCommitMessage}
	}

	for _, args := range [][]string{{"init"}, {"add", "."}, commitArgs} {
		if err := c.runGit(opts.DestinationDir(), args...); err != nil {
			return err
		}
	}

	return nil
}

func (c *Creator) maybeInitGitRepo(opts Options) error {
	mode, err := domain.ResolveGitMode(opts)
	if err != nil {
		return err
	}

	if mode == domain.GitModeNone {
		_, _ = fmt.Fprintln(c.w, "Skipping git initialization (--git none)")
		return nil
	}

	if mode == domain.GitModeInitOnly {
		_, _ = fmt.Fprintln(c.w, "Initializing git repository (--git init-only)")
		return c.runGit(opts.DestinationDir(), "init")
	}

	return c.initGitRepo(opts)
}
