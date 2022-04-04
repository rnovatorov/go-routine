package routine

import "context"

type Parent interface {
	context() context.Context
}

type RunFunc func(context.Context) error

func Go(parent Parent, child RunFunc) *Routine {
	return start(parent.context(), child)
}

type Routine struct {
	err     error
	ctx     context.Context
	cancel  context.CancelFunc
	stopped chan struct{}
}

func start(ctx context.Context, run RunFunc) *Routine {
	ctx, cancel := context.WithCancel(ctx)

	r := &Routine{
		err:     nil,
		ctx:     ctx,
		cancel:  cancel,
		stopped: make(chan struct{}),
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

func (r *Routine) Error() error {
	return r.err
}

func (r *Routine) Stop() error {
	r.Cancel()
	return r.Wait()
}

func (r *Routine) Cancel() {
	r.cancel()
}

func (r *Routine) Wait() error {
	<-r.stopped
	return r.err
}

func (r *Routine) context() context.Context {
	return r.ctx
}

type main struct {
	ctx context.Context
}

func Main(ctx context.Context) Parent {
	return &main{
		ctx: ctx,
	}
}

func (m *main) context() context.Context {
	return m.ctx
}
