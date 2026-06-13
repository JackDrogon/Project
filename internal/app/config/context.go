package config

import stdcontext "context"

type Context struct {
	ExplicitPath string
}

type (
	activeConfigContextKey struct{}
	loadContextKey         struct{}
	loadErrorContextKey    struct{}
)

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

func WithLoadContext(ctx stdcontext.Context, loadCtx Context) stdcontext.Context {
	if ctx == nil {
		ctx = stdcontext.Background()
	}

	return stdcontext.WithValue(ctx, loadContextKey{}, loadCtx)
}

func LoadContextFromContext(ctx stdcontext.Context) (Context, bool) {
	if ctx == nil {
		return Context{}, false
	}

	loadCtx, ok := ctx.Value(loadContextKey{}).(Context)
	return loadCtx, ok
}

func WithLoadError(ctx stdcontext.Context, err error) stdcontext.Context {
	if ctx == nil {
		ctx = stdcontext.Background()
	}

	return stdcontext.WithValue(ctx, loadErrorContextKey{}, err)
}

func LoadErrorFromContext(ctx stdcontext.Context) error {
	if ctx == nil {
		return nil
	}

	err, _ := ctx.Value(loadErrorContextKey{}).(error)
	return err
}
