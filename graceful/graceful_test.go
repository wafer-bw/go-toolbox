package graceful_test

import (
	"context"
	"errors"
	"syscall"
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

	// TODO: decide what this test needs to do and update it.
	// t.Run("puts the first error encountered by a start call in the error channel and ignores the rest", func(t *testing.T) {
	// 	t.Parallel()

	// 	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	// 	defer cancel()

	// 	startErr := errors.New("start failed")
	// 	runners := []graceful.Runner{
	// 		graceful.RunnerType{StartFunc: func(ctx context.Context) error { return startErr }},
	// 		graceful.RunnerType{StartFunc: func(ctx context.Context) error { return startErr }},
	// 		graceful.RunnerType{StartFunc: func(ctx context.Context) error { return startErr }},
	// 	}
	// 	g := graceful.Group{Runners: runners}

	// 	g.Start(context.Background())
	// 	err := <-g.ErrCh()
	// 	require.Error(t, err)
	// 	require.Equal(t, startErr, err)
	// 	select {
	// 	case <-g.ErrCh():
	// 		t.Fatal("unexpected error in error channel")
	// 	case <-ctx.Done():
	// 		// pass
	// 	}
	// })

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

	t.Run("blocks until it receives an error then returns it", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()

		startErr := errors.New("start failed")
		g := graceful.Group{}
		g.CreateErrCh(1)
		g.ErrCh() <- startErr

		err := g.Wait(ctx, syscall.SIGTERM)
		require.Error(t, err)
		require.Equal(t, startErr, err)
	})

	t.Run("blocks until signal is received then returns nil", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		g := graceful.Group{}
		go func() {
			time.Sleep(50 * time.Millisecond)
			err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
			require.NoError(t, err)
		}()

		err := g.Wait(ctx, syscall.SIGTERM)
		require.NoError(t, err)
		require.NoError(t, ctx.Err())
	})

	t.Run("returns context canceled error if encountered error", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		g := graceful.Group{}

		err := g.Wait(ctx, syscall.SIGTERM)
		require.Error(t, err)
		require.Equal(t, context.Canceled, err)
	})

	t.Run("returns context deadline error if encountered", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		g := graceful.Group{}

		err := g.Wait(ctx, syscall.SIGTERM)
		require.Error(t, err)
		require.Equal(t, context.DeadlineExceeded, err)
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
		err := g.Stop(context.Background(), 25*time.Millisecond)
		require.NoError(t, err)
		require.True(t, *aCalled)
		require.True(t, *bCalled)
	})

	t.Run("sets context cause to shutdown timeout error if it times out", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		runners := []graceful.Runner{
			graceful.RunnerType{StopFunc: func(ctx context.Context) error {
				<-ctx.Done()
				return context.Cause(ctx)
			}},
		}
		g := graceful.Group{Runners: runners}
		err := g.Stop(ctx, 25*time.Millisecond)
		require.Error(t, err)
		require.ErrorIs(t, err, graceful.ShutdownTimeoutError{})
	})

	t.Run("does not panic when runners are nil", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{Runners: []graceful.Runner{nil, nil, nil}}
		require.NotPanics(t, func() {
			err := g.Stop(context.Background(), 25*time.Millisecond)
			require.NoError(t, err)
		})
	})

	t.Run("does not panic when runners list is empty", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{Runners: []graceful.Runner{}}
		require.NotPanics(t, func() {
			err := g.Stop(context.Background(), 25*time.Millisecond)
			require.NoError(t, err)
		})
	})

	t.Run("does not panic when runners list is nil", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{Runners: nil}
		require.NotPanics(t, func() {
			err := g.Stop(context.Background(), 25*time.Millisecond)
			require.NoError(t, err)
		})
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
