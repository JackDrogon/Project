package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	os.Exit(run(newDependencies(), os.Args[1:], os.Stdout, os.Stderr))
}

// run owns the process-level concerns - argument source, output streams, and
// the exit code - so nothing below it writes to os.Stdout/os.Stderr or calls
// os.Exit. The failing command's error and usage are printed once, on stderr.
func run(deps dependencies, args []string, stdout, stderr io.Writer) int {
	rootCmd := newRootCmd(deps)
	rootCmd.SetArgs(args)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)

	failed, err := rootCmd.ExecuteC()
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
