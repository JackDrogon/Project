package config

// LoadOptions selects which config file to load. It was previously called
// Context, which is what pulled this package toward stuffing config into a
// context.Context: methods took it as a parameter named ctx, and the file that
// did the stuffing had to import the standard library as stdcontext to avoid
// the collision.
type LoadOptions struct {
	ExplicitPath string
}

// Resolved is the config state the root command works out once, in
// PersistentPreRunE, and that every subcommand reads while running.
//
// This used to travel as three separate context.Context values - the active
// config, the load options, and the load error - retrieved by type assertion
// and silently defaulting to the zero value when a key was absent, with the
// root command walking the whole command tree calling SetContext to propagate
// them. Go's own documentation reserves context values for request-scoped data
// crossing API boundaries, not for optional parameters, and an error is
// especially out of place there. Commands now receive a *Resolved explicitly,
// so cmd.Context() is left to mean cancellation and nothing else.
//
// The pointer is what bridges the timing gap: subcommands are constructed
// before the root command has parsed --config, so they capture the pointer at
// build time and read through it at run time.
type Resolved struct {
	Active  ActiveConfig
	Options LoadOptions
	LoadErr error
}
