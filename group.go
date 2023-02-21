package routine

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

func WaitGroup(ctx context.Context, spawn func(g *Group)) error {
	g := NewGroup(ctx)
	spawn(g)
	return g.Wait()
}

type Group struct {
	ctx    context.Context
	cancel context.CancelCauseFunc
	wg     sync.WaitGroup
	mu     sync.Mutex
	errs   []error
}

func NewGroup(ctx context.Context) *Group {
	ctx, cancel := context.WithCancelCause(ctx)

	return &Group{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (g *Group) Wait() error {
	defer g.cancel(nil)

	g.wg.Wait()

	return errors.Join(g.errs...)
}

func (g *Group) Go(name string, run func(context.Context) error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()

		defer func() {
			if v := recover(); v != nil {
				g.panicHook(v)
				panic(v)
			}
		}()

		if err := run(g.ctx); err != nil {
			err = fmt.Errorf("%s: %w", name, err)
			g.cancel(err)

			g.mu.Lock()
			g.errs = append(g.errs, err)
			g.mu.Unlock()
		}
	}()
}

func (g *Group) panicHook(v interface{}) {
	if hook, ok := g.ctx.Value(panicHookContextKey).(PanicHook); ok {
		hook(v)
	}
}

type PanicHook func(interface{})

func NewPanicHookContext(ctx context.Context, hook PanicHook) context.Context {
	return context.WithValue(ctx, panicHookContextKey, hook)
}

type contextKey struct{}

var panicHookContextKey contextKey
