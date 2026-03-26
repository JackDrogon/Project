package config

import stdcontext "context"

type Context struct {
	ExplicitPath string
}

type activeConfigContextKey struct{}

func WithActiveConfig(ctx stdcontext.Context, active ActiveConfig) stdcontext.Context {
	if ctx == nil {
		ctx = stdcontext.Background()
	}

	return stdcontext.WithValue(ctx, activeConfigContextKey{}, active)
}

func ActiveConfigFromContext(ctx stdcontext.Context) (ActiveConfig, bool) {
	if ctx == nil {
		return ActiveConfig{}, false
	}

	active, ok := ctx.Value(activeConfigContextKey{}).(ActiveConfig)
	return active, ok
}
