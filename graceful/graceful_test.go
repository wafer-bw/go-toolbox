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

		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		defer cancel()

		aCh, bCh := make(chan struct{}), make(chan struct{})
		g := graceful.Group{
			graceful.RunnerType{StartFunc: func(ctx context.Context) error {
				close(aCh)
				return nil
			}},
			graceful.RunnerType{StartFunc: func(ctx context.Context) error {
				close(bCh)
				return nil
			}},
		}

		err := g.Start(ctx, syscall.SIGTERM)
		require.NoError(t, err)
		_, aOpen := <-aCh
		require.False(t, aOpen)
		_, bOpen := <-bCh
		require.False(t, bOpen)
	})

	t.Run("returns first runner start error encountered", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		defer cancel()

		startErr := errors.New("start failed")
		g := graceful.Group{
			graceful.RunnerType{StartFunc: func(ctx context.Context) error { return startErr }},
			graceful.RunnerType{StartFunc: func(ctx context.Context) error { return startErr }},
		}

		err := g.Start(ctx, syscall.SIGTERM)
		require.Error(t, err)
		require.Equal(t, startErr, err)
	})

	t.Run("returns nil when a signal is received", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		defer cancel()
		g := graceful.Group{}
		go func() {
			time.Sleep(50 * time.Millisecond)
			err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
			require.NoError(t, err)
		}()

		err := g.Start(ctx, syscall.SIGTERM)
		require.NoError(t, err)
		require.NoError(t, ctx.Err())
	})

	t.Run("returns context canceled error if encountered", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		g := graceful.Group{}

		err := g.Start(ctx, syscall.SIGTERM)
		require.Error(t, err)
		require.Equal(t, context.Canceled, err)
	})

	t.Run("returns context deadline error if encountered", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		g := graceful.Group{}

		err := g.Start(ctx, syscall.SIGTERM)
		require.Error(t, err)
		require.Equal(t, context.DeadlineExceeded, err)
	})

	t.Run("does not panic when runners are nil", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{nil, nil, nil}
		require.NotPanics(t, func() {
			err := g.Start(context.Background())
			require.NoError(t, err)
		})
	})

	t.Run("does not panic when slice is empty", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{}
		require.NotPanics(t, func() {
			err := g.Start(context.Background())
			require.NoError(t, err)
		})
	})
}

func TestGroup_Stop(t *testing.T) {
	t.Parallel()

	t.Run("stops all runners", func(t *testing.T) {
		t.Parallel()

		aCh, bCh := make(chan struct{}), make(chan struct{})
		g := graceful.Group{
			graceful.RunnerType{StopFunc: func(ctx context.Context) error {
				close(aCh)
				return nil
			}},
			graceful.RunnerType{StopFunc: func(ctx context.Context) error {
				close(bCh)
				return nil
			}},
		}
		err := g.Stop(context.Background(), 25*time.Millisecond)
		require.NoError(t, err)
		_, aOpen := <-aCh
		require.False(t, aOpen)
		_, bOpen := <-bCh
		require.False(t, bOpen)
	})

	t.Run("returns first runner stop error encountered", func(t *testing.T) {
		t.Parallel()

		stopErr := errors.New("stop failed")
		g := graceful.Group{
			graceful.RunnerType{StopFunc: func(ctx context.Context) error { return stopErr }},
			graceful.RunnerType{StopFunc: func(ctx context.Context) error { return stopErr }},
		}
		err := g.Stop(context.Background(), 25*time.Millisecond)
		require.Error(t, err)
		require.Equal(t, stopErr, err)
	})

	t.Run("sets context cause to shutdown timeout error if it times out", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		defer cancel()

		g := graceful.Group{
			graceful.RunnerType{
				StopFunc: func(ctx context.Context) error {
					<-ctx.Done()
					return context.Cause(ctx)
				},
			},
		}
		err := g.Stop(ctx, 25*time.Millisecond)
		require.Error(t, err)
		require.ErrorIs(t, err, graceful.ShutdownTimeoutError{})
	})

	t.Run("does not panic when runners are nil", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{nil, nil, nil}
		require.NotPanics(t, func() {
			err := g.Stop(context.Background(), 25*time.Millisecond)
			require.NoError(t, err)
		})
	})

	t.Run("does not panic when slice is empty", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{}
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
