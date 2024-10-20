package routine

import (
	"context"
	"fmt"
)

type Middleware func(Run) Run

type middlewareContextKey struct{}

func WithMiddleware(ctx context.Context, outer Middleware) context.Context {
	m := outer
	if inner, ok := MiddlewareFromContext(ctx); ok {
		m = func(run Run) Run {
			return outer(inner(run))
		}
	}
	return context.WithValue(ctx, middlewareContextKey{}, m)
}

func MiddlewareFromContext(ctx context.Context) (Middleware, bool) {
	m, ok := ctx.Value(middlewareContextKey{}).(Middleware)
	return m, ok
}

func NewRecoverMiddleware(recoverer func(any)) Middleware {
	return func(run Run) Run {
		return func(ctx context.Context) (ret error) {
			defer func() {
				if v := recover(); v != nil {
					recoverer(v)
					ret = PanicRecoveredError{Value: v}
				}
			}()
			return run(ctx)
		}
	}
}

type PanicRecoveredError struct {
	Value interface{}
}

func (e PanicRecoveredError) Error() string {
	return fmt.Sprintf("panic recovered: %v", e.Value)
}
