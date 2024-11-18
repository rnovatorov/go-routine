package routine

import (
	"context"
	"errors"
	"sync"
)

type Group struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.Mutex
	errs   []error
}

func NewGroup(ctx context.Context) *Group {
	ctx, cancel := context.WithCancel(ctx)

	return &Group{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (g *Group) Go(run Run) *Routine {
	g.wg.Add(1)
	return startRoutine(g.ctx, func(ctx context.Context) error {
		defer g.wg.Done()

		if err := run(ctx); err != nil {
			g.cancel()

			g.mu.Lock()
			g.errs = append(g.errs, err)
			g.mu.Unlock()

			return err
		}

		return nil
	})
}

func (g *Group) Stop() error {
	g.cancel()
	return g.Wait()
}

func (g *Group) Wait() error {
	defer g.cancel()
	g.wg.Wait()

	g.mu.Lock()
	defer g.mu.Unlock()

	return errors.Join(g.errs...)
}
