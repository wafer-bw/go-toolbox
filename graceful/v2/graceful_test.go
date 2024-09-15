package graceful_test

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/wafer-bw/go-toolbox/graceful/v2"
)

func TestGroup_Run(t *testing.T) {
	t.Parallel()

	t.Run("calls start and stop in sequence returning the error from start", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		defer cancel()

		startErr := errors.New("start failed")
		aCh, bCh := make(chan struct{}), make(chan struct{})
		g := graceful.Group{
			Runners: []graceful.Runner{
				graceful.RunnerType{
					StartFunc: func(ctx context.Context) error {
						close(aCh)
						return startErr
					},
					StopFunc: func(ctx context.Context) error {
						close(bCh)
						return nil
					},
				},
			},
			ShutdownTimeout: 25 * time.Millisecond,
		}

		err := g.Run(ctx)
		require.Error(t, err)
		require.Equal(t, startErr, err)
		_, aOpen := <-aCh
		require.False(t, aOpen)
		_, bOpen := <-bCh
		require.False(t, bOpen)
	})
}

func TestGroup_Start(t *testing.T) {
	t.Parallel()

	t.Run("starts all runners", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		defer cancel()

		aCh, bCh := make(chan struct{}), make(chan struct{})
		g := graceful.Group{
			Runners: []graceful.Runner{
				graceful.RunnerType{StartFunc: func(ctx context.Context) error {
					close(aCh)
					return nil
				}},
				graceful.RunnerType{StartFunc: func(ctx context.Context) error {
					close(bCh)
					return nil
				}},
			},
		}

		err := g.Start(ctx)
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
			Runners: []graceful.Runner{
				graceful.RunnerType{StartFunc: func(ctx context.Context) error { return startErr }},
				graceful.RunnerType{StartFunc: func(ctx context.Context) error { return startErr }},
			},
		}

		err := g.Start(ctx)
		require.Error(t, err)
		require.Equal(t, startErr, err)
	})

	t.Run("returns nil when a signal is received", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		defer cancel()
		g := graceful.Group{ShutdownSignals: []os.Signal{syscall.SIGTERM}}
		go func() {
			time.Sleep(50 * time.Millisecond)
			err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
			require.NoError(t, err)
		}()

		err := g.Start(ctx)
		require.NoError(t, err)
		require.NoError(t, ctx.Err())
	})

	t.Run("returns context canceled error if encountered", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		g := graceful.Group{}

		err := g.Start(ctx)
		require.Error(t, err)
		require.Equal(t, context.Canceled, err)
	})

	t.Run("returns context deadline error if encountered", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		g := graceful.Group{}

		err := g.Start(ctx)
		require.Error(t, err)
		require.Equal(t, context.DeadlineExceeded, err)
	})

	t.Run("does not panic when runners are nil", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{Runners: []graceful.Runner{nil, nil, nil}}
		require.NotPanics(t, func() {
			err := g.Start(context.Background())
			require.NoError(t, err)
		})
	})

	t.Run("does not panic when slice is empty", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{Runners: []graceful.Runner{}}
		require.NotPanics(t, func() {
			err := g.Start(context.Background())
			require.NoError(t, err)
		})
	})

	t.Run("does not panic when slice is nil", func(t *testing.T) {
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

	t.Run("stops all runners concurrently", func(t *testing.T) {
		t.Parallel()

		aCh, bCh := make(chan struct{}), make(chan struct{})
		g := graceful.Group{
			Runners: []graceful.Runner{
				graceful.RunnerType{StopFunc: func(ctx context.Context) error {
					close(aCh)
					return nil
				}},
				graceful.RunnerType{StopFunc: func(ctx context.Context) error {
					close(bCh)
					return nil
				}},
			},
			ShutdownTimeout: 25 * time.Millisecond,
		}
		err := g.Stop(context.Background())
		require.NoError(t, err)
		_, aOpen := <-aCh
		require.False(t, aOpen)
		_, bOpen := <-bCh
		require.False(t, bOpen)
	})

	t.Run("stops all runners sequentially", func(t *testing.T) {
		t.Parallel()

		sqCh := make(chan string, 3)
		aCh, bCh, cCh := make(chan struct{}), make(chan struct{}), make(chan struct{})
		defer close(sqCh)

		g := graceful.Group{
			Runners: []graceful.Runner{
				graceful.RunnerType{StopFunc: func(ctx context.Context) error {
					close(aCh)
					sqCh <- "a"
					return nil
				}},
				graceful.RunnerType{StopFunc: func(ctx context.Context) error {
					close(bCh)
					sqCh <- "b"
					return nil
				}},
				graceful.RunnerType{StopFunc: func(ctx context.Context) error {
					close(cCh)
					sqCh <- "c"
					return nil
				}},
			},
			ShutdownTimeout:  25 * time.Millisecond,
			SequentiallyStop: true,
		}
		err := g.Stop(context.Background())
		require.NoError(t, err)
		_, aOpen := <-aCh
		require.False(t, aOpen)
		_, bOpen := <-bCh
		require.False(t, bOpen)
		require.Equal(t, "a", <-sqCh)
		require.Equal(t, "b", <-sqCh)
		require.Equal(t, "c", <-sqCh)
	})

	t.Run("stops all runners sequentially in reverse", func(t *testing.T) {
		t.Parallel()

		sqCh := make(chan string, 3)
		aCh, bCh, cCh := make(chan struct{}), make(chan struct{}), make(chan struct{})
		defer close(sqCh)

		g := graceful.Group{
			Runners: []graceful.Runner{
				graceful.RunnerType{StopFunc: func(ctx context.Context) error {
					close(aCh)
					sqCh <- "a"
					return nil
				}},
				graceful.RunnerType{StopFunc: func(ctx context.Context) error {
					close(bCh)
					sqCh <- "b"
					return nil
				}},
				graceful.RunnerType{StopFunc: func(ctx context.Context) error {
					close(cCh)
					sqCh <- "c"
					return nil
				}},
			},
			ShutdownTimeout:  25 * time.Millisecond,
			SequentiallyStop: true,
			ReverseStop:      true,
		}
		err := g.Stop(context.Background())
		require.NoError(t, err)
		_, aOpen := <-aCh
		require.False(t, aOpen)
		_, bOpen := <-bCh
		require.False(t, bOpen)
		require.Equal(t, "c", <-sqCh)
		require.Equal(t, "b", <-sqCh)
		require.Equal(t, "a", <-sqCh)
	})

	t.Run("returns first concurrent runner stop error encountered", func(t *testing.T) {
		t.Parallel()

		stopErr := errors.New("stop failed")
		g := graceful.Group{
			Runners: []graceful.Runner{
				graceful.RunnerType{StopFunc: func(ctx context.Context) error { return stopErr }},
				graceful.RunnerType{StopFunc: func(ctx context.Context) error { return stopErr }},
			},
			ShutdownTimeout: 25 * time.Millisecond,
		}
		err := g.Stop(context.Background())
		require.Error(t, err)
		require.Equal(t, stopErr, err)
	})

	t.Run("returns first sequential runner stop error encountered", func(t *testing.T) {
		t.Parallel()

		stopErr1 := errors.New("stop failed1")
		stopErr2 := errors.New("stop failed2")
		g := graceful.Group{
			Runners: []graceful.Runner{
				graceful.RunnerType{StopFunc: func(ctx context.Context) error { return stopErr1 }},
				graceful.RunnerType{StopFunc: func(ctx context.Context) error { return stopErr2 }},
			},
			ShutdownTimeout:  25 * time.Millisecond,
			SequentiallyStop: true,
		}
		err := g.Stop(context.Background())
		require.Error(t, err)
		require.Equal(t, stopErr1, err)
	})

	t.Run("sets context cause to shutdown timeout error if it times out", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		defer cancel()

		g := graceful.Group{
			Runners: []graceful.Runner{
				graceful.RunnerType{
					StopFunc: func(ctx context.Context) error {
						<-ctx.Done()
						return context.Cause(ctx)
					},
				},
			},
			ShutdownTimeout: 25 * time.Millisecond,
		}
		err := g.Stop(ctx)
		require.Error(t, err)
		require.ErrorIs(t, err, graceful.ShutdownTimeoutError{})
	})

	t.Run("does not panic when concurrent runners are nil", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{
			Runners:         []graceful.Runner{nil, nil, nil},
			ShutdownTimeout: 25 * time.Millisecond,
		}
		require.NotPanics(t, func() {
			err := g.Stop(context.Background())
			require.NoError(t, err)
		})
	})

	t.Run("does not panic when sequential runners are nil", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{
			Runners:          []graceful.Runner{nil, nil, nil},
			ShutdownTimeout:  25 * time.Millisecond,
			SequentiallyStop: true,
		}
		require.NotPanics(t, func() {
			err := g.Stop(context.Background())
			require.NoError(t, err)
		})
	})

	t.Run("does not panic when concurrent slice is empty", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{
			Runners:         []graceful.Runner{},
			ShutdownTimeout: 25 * time.Millisecond,
		}
		require.NotPanics(t, func() {
			err := g.Stop(context.Background())
			require.NoError(t, err)
		})
	})

	t.Run("does not panic when sequential slice is empty", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{
			Runners:          []graceful.Runner{},
			ShutdownTimeout:  25 * time.Millisecond,
			SequentiallyStop: true,
		}
		require.NotPanics(t, func() {
			err := g.Stop(context.Background())
			require.NoError(t, err)
		})
	})

	t.Run("does not panic when concurrent slice is nil", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{ShutdownTimeout: 25 * time.Millisecond}
		require.NotPanics(t, func() {
			err := g.Stop(context.Background())
			require.NoError(t, err)
		})
	})

	t.Run("does not panic when sequential slice is nil", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{
			ShutdownTimeout:  25 * time.Millisecond,
			SequentiallyStop: true,
		}
		require.NotPanics(t, func() {
			err := g.Stop(context.Background())
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
