package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	os.Exit(runMain())
}

// runMain exists so the signal handler is torn down before os.Exit skips
// deferred calls.
func runMain() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return run(ctx, newDependencies(), os.Args[1:], os.Stdout, os.Stderr)
}

// run owns the process-level concerns - argument source, output streams, and
// the exit code - so nothing below it writes to os.Stdout/os.Stderr or calls
// os.Exit. The failing command's error and usage are printed once, on stderr.
func run(ctx context.Context, deps dependencies, args []string, stdout, stderr io.Writer) int {
	rootCmd := newRootCmd(deps)
	rootCmd.SetArgs(args)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)

	failed, err := rootCmd.ExecuteContextC(ctx)
	if err == nil {
		return 0
	}
	if failed == nil {
		failed = rootCmd
	}

	_, _ = fmt.Fprintln(stderr, failed.ErrPrefix(), err.Error())
	_, _ = fmt.Fprint(stderr, failed.UsageString())
	return 1
}
