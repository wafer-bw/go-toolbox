package graceful_test

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/wafer-bw/go-toolbox/graceful"
)

func TestRunnerType_Start(t *testing.T) {
	t.Parallel()

	t.Run("does not panic when StartFunc is nil", func(t *testing.T) {
		t.Parallel()

		r := graceful.RunnerType{}
		require.NotPanics(t, func() {
			err := r.Start(t.Context())
			require.NoError(t, err)
		})
	})

	t.Run("returns StartFunc error", func(t *testing.T) {
		t.Parallel()

		startErr := errors.New("start failed")
		r := graceful.RunnerType{StartFunc: func(ctx context.Context) error { return startErr }}
		err := r.Start(t.Context())
		require.ErrorIs(t, startErr, err)
	})
}

func TestRunnerType_Stop(t *testing.T) {
	t.Parallel()

	t.Run("does not panic when StopFunc is nil", func(t *testing.T) {
		t.Parallel()

		r := graceful.RunnerType{}
		require.NotPanics(t, func() {
			err := r.Stop(t.Context())
			require.NoError(t, err)
		})
	})

	t.Run("returns StopFunc error", func(t *testing.T) {
		t.Parallel()

		stopErr := errors.New("stop failed")
		r := graceful.RunnerType{StopFunc: func(ctx context.Context) error { return stopErr }}
		err := r.Stop(t.Context())
		require.ErrorIs(t, stopErr, err)
	})
}

func TestGroup_Start(t *testing.T) {
	t.Parallel()

	t.Run("starts all runners", func(t *testing.T) {
		t.Parallel()

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

		err := g.Start(t.Context())
		require.NoError(t, err)
		_, aOpen := <-aCh
		require.False(t, aOpen)
		_, bOpen := <-bCh
		require.False(t, bOpen)
	})

	t.Run("returns first runner start error encountered", func(t *testing.T) {
		t.Parallel()

		startErr := errors.New("start failed")
		g := graceful.Group{
			graceful.RunnerType{StartFunc: func(ctx context.Context) error { return nil }},
			graceful.RunnerType{StartFunc: func(ctx context.Context) error { return startErr }},
			graceful.RunnerType{StartFunc: func(ctx context.Context) error { return startErr }},
			graceful.RunnerType{StartFunc: func(ctx context.Context) error { return nil }},
		}

		err := g.Start(t.Context())
		require.Error(t, err)
		require.Equal(t, startErr, err)
	})

	t.Run("does not panic when runners are nil", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{nil, nil, nil}
		require.NotPanics(t, func() {
			err := g.Start(t.Context())
			require.NoError(t, err)
		})
	})

	t.Run("blocks indefinitely if all starts block", func(t *testing.T) {
		t.Parallel()

		timeout := 25 * time.Millisecond
		timeoutCause := errors.New("parent context timeout")

		g := graceful.Group{
			graceful.RunnerType{StartFunc: func(ctx context.Context) error {
				<-ctx.Done()
				return ctx.Err()
			}},
			graceful.RunnerType{StartFunc: func(ctx context.Context) error {
				<-ctx.Done()
				return ctx.Err()
			}},
		}

		ctx, cancel := context.WithTimeoutCause(t.Context(), timeout, timeoutCause)
		defer cancel()

		err := g.Start(ctx)
		require.Error(t, err)
		cause := context.Cause(ctx)
		require.ErrorIs(t, cause, timeoutCause)

	})

	t.Run("does not panic when slice is empty", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{}
		require.NotPanics(t, func() {
			err := g.Start(t.Context())
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

		err := g.Stop(t.Context())
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
			graceful.RunnerType{StopFunc: func(ctx context.Context) error { return nil }},
			graceful.RunnerType{StopFunc: func(ctx context.Context) error { return stopErr }},
			graceful.RunnerType{StopFunc: func(ctx context.Context) error { return stopErr }},
			graceful.RunnerType{StopFunc: func(ctx context.Context) error { return nil }},
		}
		err := g.Stop(t.Context())
		require.Error(t, err)
		require.Equal(t, stopErr, err)
	})

	t.Run("does not panic when runners are nil", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{nil, nil, nil}
		require.NotPanics(t, func() {
			err := g.Stop(t.Context())
			require.NoError(t, err)
		})
	})

	t.Run("does not panic when slice is empty", func(t *testing.T) {
		t.Parallel()

		g := graceful.Group{}
		require.NotPanics(t, func() {
			err := g.Stop(t.Context())
			require.NoError(t, err)
		})
	})
}

func TestGroup_Run(t *testing.T) {
	// this test uses signals so it cannot be run using t.Parallel().

	t.Run("successfully stops via signal within timeout", func(t *testing.T) {
		ctx := t.Context()
		aStartCh, bStartCh, cStartCh := make(chan struct{}), make(chan struct{}), make(chan struct{})
		aStopCh, bStopCh, cStopCh := make(chan struct{}), make(chan struct{}), make(chan struct{})
		g := graceful.Group{
			graceful.RunnerType{
				StartFunc: func(ctx context.Context) error {
					close(aStartCh)
					<-ctx.Done()
					return nil
				},
				StopFunc: func(ctx context.Context) error {
					close(aStopCh)
					return nil
				},
			},
			graceful.RunnerType{
				StartFunc: func(ctx context.Context) error {
					close(bStartCh)
					<-ctx.Done()
					return nil
				},
				StopFunc: func(ctx context.Context) error {
					close(bStopCh)
					return nil
				},
			},
			graceful.Group{
				graceful.RunnerType{
					StartFunc: func(ctx context.Context) error {
						close(cStartCh)
						return nil
					},
					StopFunc: func(ctx context.Context) error {
						close(cStopCh)
						return nil
					},
				},
			},
		}

		go func() {
			time.Sleep(50 * time.Millisecond)
			p, err := os.FindProcess(syscall.Getpid())
			require.NoError(t, err)
			require.NoError(t, p.Signal(syscall.SIGHUP))
		}()

		err := g.Run(ctx,
			graceful.WithStopSignals(syscall.SIGHUP),
			graceful.WithStopTimeout(100*time.Millisecond),
		)
		require.NoError(t, err)
		_, aStartOpen := <-aStartCh
		require.False(t, aStartOpen)
		_, aStopOpen := <-aStopCh
		require.False(t, aStopOpen)
		_, bStartOpen := <-bStartCh
		require.False(t, bStartOpen)
		_, bStopOpen := <-bStopCh
		require.False(t, bStopOpen)
		_, cStartOpen := <-cStartCh
		require.False(t, cStartOpen)
		_, cStopOpen := <-cStopCh
		require.False(t, cStopOpen)
	})

	t.Run("calls all stops when parent context is done", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		aStartCh, bStartCh := make(chan struct{}), make(chan struct{})
		aStopCh, bStopCh := make(chan struct{}), make(chan struct{})
		g := graceful.Group{
			graceful.RunnerType{
				StartFunc: func(ctx context.Context) error {
					close(aStartCh)
					return nil
				},
				StopFunc: func(ctx context.Context) error {
					close(aStopCh)
					return nil
				},
			},
			graceful.RunnerType{
				StartFunc: func(ctx context.Context) error {
					close(bStartCh)
					return nil
				},
				StopFunc: func(ctx context.Context) error {
					close(bStopCh)
					return nil
				},
			},
		}

		err := g.Run(ctx)
		require.Equal(t, ctx.Err(), err)
		_, aStartOpen := <-aStartCh
		require.False(t, aStartOpen)
		_, aStopOpen := <-aStopCh
		require.False(t, aStopOpen)
		_, bStartOpen := <-bStartCh
		require.False(t, bStartOpen)
		_, bStopOpen := <-bStopCh
		require.False(t, bStopOpen)
	})

	t.Run("calls all stops when a start fails", func(t *testing.T) {
		ctx := t.Context()
		fail := errors.New("oh no")
		aStartCh, bStartCh := make(chan struct{}), make(chan struct{})
		aStopCh, bStopCh := make(chan struct{}), make(chan struct{})
		g := graceful.Group{
			graceful.RunnerType{
				StartFunc: func(ctx context.Context) error {
					close(aStartCh)
					return fail
				},
				StopFunc: func(ctx context.Context) error {
					close(aStopCh)
					return nil
				},
			},
			graceful.RunnerType{
				StartFunc: func(ctx context.Context) error {
					close(bStartCh)
					return nil
				},
				StopFunc: func(ctx context.Context) error {
					close(bStopCh)
					return nil
				},
			},
		}

		err := g.Run(ctx)
		require.Equal(t, fail, err)
		_, aStartOpen := <-aStartCh
		require.False(t, aStartOpen)
		_, aStopOpen := <-aStopCh
		require.False(t, aStopOpen)
		_, bStartOpen := <-bStartCh
		require.False(t, bStartOpen)
		_, bStopOpen := <-bStopCh
		require.False(t, bStopOpen)
	})

	t.Run("calls all stops when a stop fails", func(t *testing.T) {
		ctx := t.Context()
		fail := errors.New("oh no")
		aStartCh, bStartCh := make(chan struct{}), make(chan struct{})
		aStopCh, bStopCh := make(chan struct{}), make(chan struct{})
		g := graceful.Group{
			graceful.RunnerType{
				StartFunc: func(ctx context.Context) error {
					close(aStartCh)
					return nil
				},
				StopFunc: func(ctx context.Context) error {
					close(aStopCh)
					return nil
				},
			},
			graceful.RunnerType{
				StartFunc: func(ctx context.Context) error {
					close(bStartCh)
					return fail
				},
				StopFunc: func(ctx context.Context) error {
					close(bStopCh)
					return nil
				},
			},
		}

		err := g.Run(ctx)
		require.Equal(t, fail, err)
		_, aStartOpen := <-aStartCh
		require.False(t, aStartOpen)
		_, aStopOpen := <-aStopCh
		require.False(t, aStopOpen)
		_, bStartOpen := <-bStartCh
		require.False(t, bStartOpen)
		_, bStopOpen := <-bStopCh
		require.False(t, bStopOpen)
	})

	t.Run("waits indefinitely for stops when no timeout is provided", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		g := graceful.Group{
			graceful.RunnerType{
				StartFunc: func(ctx context.Context) error {
					<-ctx.Done()
					return nil
				},
				StopFunc: func(ctx context.Context) error {
					<-ctx.Done()
					return nil
				},
			},
			graceful.RunnerType{
				StartFunc: func(ctx context.Context) error {
					<-ctx.Done()
					return nil
				},
				StopFunc: func(ctx context.Context) error {
					<-ctx.Done()
					return nil
				},
			},
		}

		go func() {
			time.Sleep(50 * time.Millisecond)
			p, err := os.FindProcess(syscall.Getpid())
			require.NoError(t, err)
			require.NoError(t, p.Signal(syscall.SIGHUP))
		}()

		err := g.Run(ctx, graceful.WithStopSignals(syscall.SIGHUP))
		require.NoError(t, err)
	})

	t.Run("waits for parent context to stop when no signals are provided", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		g := graceful.Group{
			graceful.RunnerType{
				StartFunc: func(ctx context.Context) error {
					<-ctx.Done()
					return nil
				},
				StopFunc: func(ctx context.Context) error {
					<-ctx.Done()
					return nil
				},
			},
			graceful.RunnerType{
				StartFunc: func(ctx context.Context) error {
					<-ctx.Done()
					return nil
				},
				StopFunc: func(ctx context.Context) error {
					<-ctx.Done()
					return nil
				},
			},
		}

		err := g.Run(ctx, graceful.WithStopTimeout(50*time.Millisecond))
		require.Equal(t, ctx.Err(), err)
	})

	t.Run("closes stopping channel", func(t *testing.T) {
		ctx := t.Context()
		stoppingCh := make(chan struct{})
		g := graceful.Group{
			graceful.RunnerType{
				StartFunc: func(ctx context.Context) error { return nil },
				StopFunc:  func(ctx context.Context) error { return nil },
			},
		}

		go func() {
			time.Sleep(50 * time.Millisecond)
			p, err := os.FindProcess(syscall.Getpid())
			require.NoError(t, err)
			require.NoError(t, p.Signal(syscall.SIGHUP))
		}()

		go func() {
			_ = g.Run(ctx,
				graceful.WithStopSignals(syscall.SIGHUP),
				graceful.WithStopTimeout(100*time.Millisecond),
				graceful.WithStoppingCh(stoppingCh),
			)
		}()

		_, open := <-stoppingCh
		require.False(t, open)
	})

	t.Run("does not panic when provided nil signal", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		g := graceful.Group{
			graceful.RunnerType{
				StartFunc: func(ctx context.Context) error { return nil },
				StopFunc:  func(ctx context.Context) error { return nil },
			},
			graceful.RunnerType{
				StartFunc: func(ctx context.Context) error { return nil },
				StopFunc:  func(ctx context.Context) error { return nil },
			},
		}

		go func() {
			time.Sleep(50 * time.Millisecond)
			p, err := os.FindProcess(syscall.Getpid())
			require.NoError(t, err)
			require.NoError(t, p.Signal(syscall.SIGHUP))
		}()

		require.NotPanics(t, func() {
			err := g.Run(ctx, graceful.WithStopSignals(nil, syscall.SIGHUP))
			require.NoError(t, err)
		})
	})

	t.Run("does not panic when provided no signals", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		g := graceful.Group{
			graceful.RunnerType{
				StartFunc: func(ctx context.Context) error { return nil },
				StopFunc:  func(ctx context.Context) error { return nil },
			},
			graceful.RunnerType{
				StartFunc: func(ctx context.Context) error { return nil },
				StopFunc:  func(ctx context.Context) error { return nil },
			},
		}

		require.NotPanics(t, func() {
			err := g.Run(ctx, graceful.WithStopSignals())
			require.Equal(t, ctx.Err(), err)
		})
	})

	t.Run("does not panic when provided nil options", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
		defer cancel()

		g := graceful.Group{
			graceful.RunnerType{
				StartFunc: func(ctx context.Context) error { return nil },
				StopFunc:  func(ctx context.Context) error { return nil },
			},
		}

		require.NotPanics(t, func() {
			err := g.Run(ctx, nil)
			require.Equal(t, ctx.Err(), err)
		})
	})

	t.Run("returns start error when there is a stop error", func(t *testing.T) {
		ctx := t.Context()
		startErr := errors.New("start failed")
		stopErr := errors.New("stop failed")
		g := graceful.Group{
			graceful.RunnerType{
				StartFunc: func(ctx context.Context) error { return startErr },
				StopFunc:  func(ctx context.Context) error { return stopErr },
			},
		}

		err := g.Run(ctx)
		require.Error(t, err)
		require.Equal(t, startErr, err)
	})

	t.Run("returns stop error when there is a run error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		stopErr := errors.New("stop failed")
		g := graceful.Group{
			graceful.RunnerType{
				StartFunc: func(ctx context.Context) error { return nil },
				StopFunc:  func(ctx context.Context) error { return stopErr },
			},
		}

		err := g.Run(ctx)
		require.Error(t, err)
		require.Equal(t, stopErr, err)
	})
}
