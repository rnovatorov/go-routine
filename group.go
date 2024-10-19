package routine

import (
	"context"
	"errors"
	"sync"
)

type Group struct {
	*Routine
	wg   sync.WaitGroup
	mu   sync.Mutex
	errs []error
}

func StartGroup(ctx context.Context) *Group {
	g := &Group{}
	g.Routine = startRoutine(ctx, g.run)
	return g
}

func (g *Group) Go(run Run) *Routine {
	g.wg.Add(1)
	return startRoutine(g.ctx, func(ctx context.Context) error {
		defer g.wg.Done()

		if err := run(ctx); err != nil {
			g.Cancel()

			g.mu.Lock()
			g.errs = append(g.errs, err)
			g.mu.Unlock()
		}

		return nil
	})
}

func (g *Group) run(ctx context.Context) error {
	<-ctx.Done()

	g.wg.Wait()

	g.mu.Lock()
	defer g.mu.Unlock()

	return errors.Join(g.errs...)
}
