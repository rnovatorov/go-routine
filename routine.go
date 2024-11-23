package routine

import "context"

type Routine struct {
	err     error
	ctx     context.Context
	cancel  context.CancelFunc
	started chan struct{}
	stopped chan struct{}
}

func Go(ctx context.Context, run Run) *Routine {
	ctx, cancel := context.WithCancel(ctx)

	r := &Routine{
		err:     nil,
		ctx:     ctx,
		cancel:  cancel,
		started: make(chan struct{}),
		stopped: make(chan struct{}),
	}

	if middleware, ok := MiddlewareFromContext(ctx); ok {
		run = middleware(run)
	}

	go func() {
		defer close(r.stopped)
		defer r.cancel()

		r.err = run(newContext(r.ctx, r))
	}()

	return r
}

func Started(ctx context.Context) {
	r := fromContext(ctx)
	close(r.started)
}

func (r *Routine) WaitStarted() error {
	select {
	case <-r.started:
		return nil
	case <-r.stopped:
		return r.err
	}
}

func (r *Routine) Stopped() <-chan struct{} {
	return r.stopped
}

func (r *Routine) Stop() error {
	r.cancel()
	return r.WaitStopped()
}

func (r *Routine) Cancel() {
	r.cancel()
}

func (r *Routine) WaitStopped() error {
	<-r.stopped
	return r.err
}

type contextKey struct{}

func newContext(ctx context.Context, r *Routine) context.Context {
	return context.WithValue(ctx, contextKey{}, r)
}

func fromContext(ctx context.Context) *Routine {
	return ctx.Value(contextKey{}).(*Routine)
}
