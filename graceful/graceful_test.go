package graceful_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/wafer-bw/go-toolbox/graceful"
)

func TestGroup_Start(t *testing.T) {
	t.Parallel()

	t.Run("starts all runners", func(t *testing.T) {
		t.Parallel()

		runners := []graceful.Runner{
			graceful.RunnerType{StartFunc: func(ctx context.Context) error { return nil }},
			graceful.RunnerType{StartFunc: func(ctx context.Context) error { return nil }},
		}
		g := graceful.Group{Runners: runners}
		g.Start(context.Background())
	})

	t.Run("sends errors returned by runners start functions to group error channel", func(t *testing.T) {
		t.Parallel()

		startErr := errors.New("start failed")
		runners := []graceful.Runner{
			graceful.RunnerType{StartFunc: func(ctx context.Context) error { return startErr }},
			graceful.RunnerType{StartFunc: func(ctx context.Context) error { return startErr }},
			graceful.RunnerType{StartFunc: func(ctx context.Context) error { return startErr }},
		}
		g := graceful.Group{Runners: runners}
		g.Start(context.Background())
		for i := 0; i < len(runners); i++ {
			err := <-g.ErrCh()
			require.Error(t, err)
			require.Equal(t, startErr, err)
		}
	})

	t.Run("does not panic when runners are nil", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{Runners: []graceful.Runner{nil, nil, nil}}
		require.NotPanics(t, func() { g.Start(context.Background()) })
	})

	t.Run("does not panic when runners list is empty", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{Runners: []graceful.Runner{}}
		require.NotPanics(t, func() { g.Start(context.Background()) })
	})

	t.Run("does not panic when runners list is nil", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{Runners: nil}
		require.NotPanics(t, func() { g.Start(context.Background()) })
	})
}

func TestGroup_Wait(t *testing.T) {
	t.Parallel()

	t.Run("blocks until it receives an error", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		startErr := errors.New("start failed")
		g := graceful.Group{}
		g.CreateErrCh()
		go func() { g.ErrCh() <- startErr }()

		g.Wait(ctx) // will unblock when error is received before timeout.
		require.NotEqual(t, context.DeadlineExceeded, ctx.Err())
	})

	t.Run("blocks until its context is cancelled", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		g := graceful.Group{}
		g.CreateErrCh()

		g.Wait(ctx) // will block until timeout is cancelled.
		require.Equal(t, context.DeadlineExceeded, ctx.Err())
	})
}

func TestGroup_Stop(t *testing.T) {
	t.Parallel()

	t.Run("stops all runners", func(t *testing.T) {
		t.Parallel()

		aCalled, bCalled := new(bool), new(bool)
		runners := []graceful.Runner{
			graceful.RunnerType{StopFunc: func(ctx context.Context) error {
				*aCalled = true
				return nil
			}},
			graceful.RunnerType{StopFunc: func(ctx context.Context) error {
				*bCalled = true
				return nil
			}},
		}
		g := graceful.Group{Runners: runners}
		g.Stop(context.Background())
		require.True(t, *aCalled)
		require.True(t, *bCalled)
	})

	t.Run("sets context cause to graceful shutdown timeout error if it times out", func(t *testing.T) {
		t.Parallel()

		octx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		wasExpectedCause := new(bool)
		runners := []graceful.Runner{
			graceful.RunnerType{StopFunc: func(ctx context.Context) error {
				<-ctx.Done()
				*wasExpectedCause = errors.Is(context.Cause(ctx), graceful.ShutdownTimeoutError{})
				return nil
			}},
		}
		g := graceful.Group{
			Runners:     runners,
			StopTimeout: 10 * time.Millisecond,
		}
		g.Stop(octx)
		require.True(t, *wasExpectedCause)
	})

	t.Run("does not panic when runners are nil", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{Runners: []graceful.Runner{nil, nil, nil}}
		require.NotPanics(t, func() { g.Stop(context.Background()) })
	})

	t.Run("does not panic when runners list is empty", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{Runners: []graceful.Runner{}}
		require.NotPanics(t, func() { g.Stop(context.Background()) })
	})

	t.Run("does not panic when runners list is nil", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{Runners: nil}
		require.NotPanics(t, func() { g.Stop(context.Background()) })
	})
}

func TestGroup_Run(t *testing.T) {
	t.Parallel()

	t.Run("calls Start, Wait, and Stop in sequence", func(t *testing.T) {
		t.Parallel()

		startErr := errors.New("start failed")
		startCalled, stopCalled := new(bool), new(bool)
		runners := []graceful.Runner{
			graceful.RunnerType{StartFunc: func(ctx context.Context) error {
				*startCalled = true
				return startErr
			}},
			graceful.RunnerType{StopFunc: func(ctx context.Context) error {
				*stopCalled = true
				return nil
			}},
		}
		g := graceful.Group{Runners: runners}
		g.Run(context.Background())
		require.True(t, *startCalled)
		require.True(t, *stopCalled)
		require.True(t, g.Waited())
	})

}

func TestRunnerType_Start(t *testing.T) {
	t.Parallel()

	t.Run("calls StartFunc and returns its error", func(t *testing.T) {
		t.Parallel()

		startErr := errors.New("error")
		rt := graceful.RunnerType{
			StartFunc: func(ctx context.Context) error { return startErr },
			StopFunc:  func(ctx context.Context) error { return nil },
		}

		err := rt.Start(context.Background())
		require.Error(t, err)
		require.Equal(t, startErr, err)
	})

	t.Run("does not panic when StartFunc is nil", func(t *testing.T) {
		t.Parallel()

		rt := graceful.RunnerType{
			StartFunc: nil,
			StopFunc:  func(ctx context.Context) error { return nil },
		}

		require.NotPanics(t, func() { _ = rt.Start(context.Background()) })
	})
}

func TestRunnerType_Stop(t *testing.T) {
	t.Parallel()

	t.Run("calls StopFunc and returns its error", func(t *testing.T) {
		t.Parallel()

		stopErr := errors.New("error")
		rt := graceful.RunnerType{
			StartFunc: func(ctx context.Context) error { return nil },
			StopFunc:  func(ctx context.Context) error { return stopErr },
		}

		err := rt.Stop(context.Background())
		require.Error(t, err)
		require.Equal(t, stopErr, err)
	})

	t.Run("does not panic when StopFunc is nil", func(t *testing.T) {
		t.Parallel()

		rt := graceful.RunnerType{
			StartFunc: func(ctx context.Context) error { return nil },
			StopFunc:  nil,
		}

		require.NotPanics(t, func() { _ = rt.Stop(context.Background()) })
	})
}

func TestShutdownTimeoutError_Error(t *testing.T) {
	t.Parallel()

	t.Run("returns error message", func(t *testing.T) {
		t.Parallel()

		err := graceful.ShutdownTimeoutError{}
		require.Greater(t, len(err.Error()), 0)
	})
}

func TestShutdownTimeoutError_Timeout(t *testing.T) {
	t.Parallel()

	t.Run("returns true", func(t *testing.T) {
		t.Parallel()

		err := graceful.ShutdownTimeoutError{}
		require.True(t, err.Timeout())
	})
}
