package routine_test

import (
	"context"
	"testing"

	"github.com/rnovatorov/go-routine"
)

func TestStarted(t *testing.T) {
	g := routine.NewGroup(context.Background())
	setup := false
	r := g.Go(func(ctx context.Context) error {
		setup = true
		routine.Started(ctx)
		return nil
	})
	if err := r.WaitStarted(); err != nil {
		t.Logf("unexpected error: %v", err)
		t.FailNow()
	}
	if !setup {
		t.Log("setup not completed")
		t.FailNow()
	}
}
