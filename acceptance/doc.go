// Package acceptance holds end-to-end checks that build the real CLI binary,
// scaffold projects from the real embedded templates, and validate the result
// with each language's own toolchain.
//
// The tests are gated behind the `acceptance` build tag so the default
// `go test ./...` run stays fast and toolchain-independent. Run them with:
//
//	just acceptance
package acceptance
