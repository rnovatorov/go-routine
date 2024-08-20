package routine

import "context"

type Run func(context.Context) error

type Routine struct {
	err     error
	ctx     context.Context
	cancel  context.CancelFunc
	stopped chan struct{}
}

func Go(ctx context.Context, run Run) *Routine {
	ctx, cancel := context.WithCancel(ctx)

	r := &Routine{
		err:     nil,
		ctx:     ctx,
		cancel:  cancel,
		stopped: make(chan struct{}),
	}

	if middleware, ok := MiddlewareFromContext(ctx); ok {
		run = middleware(run)
	}

	go func() {
		defer close(r.stopped)
		defer r.cancel()

		r.err = run(r.ctx)
	}()

	return r
}

func (r *Routine) Stopped() <-chan struct{} {
	return r.stopped
}

func (r *Routine) Stop() error {
	r.cancel()
	return r.Wait()
}

func (r *Routine) Cancel() {
	r.cancel()
}

func (r *Routine) Wait() error {
	<-r.stopped
	return r.err
}
