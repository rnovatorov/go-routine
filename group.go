package routine

import (
	"context"
	"fmt"
	"sync"
)

type Group struct {
	ctx       context.Context
	cancel    context.CancelFunc
	panicHook PanicHook
	wg        sync.WaitGroup
	mu        sync.Mutex
	err       error
}

func NewGroup(ctx context.Context) *Group {
	ctx, cancel := context.WithCancel(ctx)

	panicHook, _ := panicHookFromContext(ctx)

	return &Group{
		ctx:       ctx,
		cancel:    cancel,
		panicHook: panicHook,
	}
}

func (g *Group) Wait() error {
	defer g.cancel()

	g.wg.Wait()

	g.mu.Lock()
	defer g.mu.Unlock()

	return g.err
}

func (g *Group) Go(name string, run func(context.Context) error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()

		if err := func() (retErr error) {
			defer g.recover()
			return run(g.ctx)
		}(); err != nil {
			g.cancel()

			g.mu.Lock()
			defer g.mu.Unlock()

			if g.err == nil {
				g.err = fmt.Errorf("%s: %w", name, err)
			} else {
				g.err = fmt.Errorf("%w; %s: %w", g.err, name, err)
			}
		}
	}()
}

func (g *Group) recover() {
	if v := recover(); v != nil {
		if hook := g.panicHook; hook != nil {
			hook(v)
		}
		panic(v)
	}
}

type contextKey struct{}

var panicHookContextKey contextKey

type PanicHook func(interface{})

func NewPanicHookContext(ctx context.Context, hook PanicHook) context.Context {
	return context.WithValue(ctx, panicHookContextKey, hook)
}

func panicHookFromContext(ctx context.Context) (PanicHook, bool) {
	hook, ok := ctx.Value(panicHookContextKey).(PanicHook)
	return hook, ok
}

func WaitGroup(ctx context.Context, spawn func(g *Group)) error {
	g := NewGroup(ctx)
	spawn(g)
	return g.Wait()
}
