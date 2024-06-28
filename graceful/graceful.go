package graceful

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"
)

// GracefulShutdownTimeoutError is returned when a [Runner] is unable to
// complete [Runner.Stop] before [Group.StopTimeout] during execution of
// [Group.Stop].
type GracefulShutdownTimeoutError struct{}

func (GracefulShutdownTimeoutError) Error() string {
	return "graceful shutdown timeout"
}

// Timeout returns true if the error is a timeout error, this allows callers
// to identify the nature of the error without needing to match the error
// based on equality.
func (GracefulShutdownTimeoutError) Timeout() bool {
	return true
}

// Runner is capable of starting and stopping itself
type Runner interface {
	Start(context.Context) error

	// Stop must complete within the lifetime of the context passed to it.
	Stop(context.Context) error
}

// Group is used to run multiple [Runner]s concurrently.
type Group struct {
	Runners []Runner

	// StopTimeout is the maximum time to wait for all runners to gracefully
	// stop after [Group.Stop] is called. If a runner does not stop within this
	// time, the context passed to [Runner.Stop] will be canceled.
	StopTimeout time.Duration

	errCh chan error // TODO: maybe replace with error group?

	// TODO: maybe use a channel for msgs so they can use their own logger that listens to the channel?
}

// Start all runners concurrently
func (g Group) Start(ctx context.Context) error {
	g.errCh = make(chan error)

	for _, runner := range g.Runners {
		if runner == nil {
			continue
		}

		go func(r Runner) {
			if err := r.Start(ctx); err != nil {
				// TODO: log?
				g.errCh <- err
			}
		}(runner)
	}

	return nil
}

// Wait blocks until a runner encounters an error or a signal is received.
func (g Group) Wait(ctx context.Context, signals ...os.Signal) {
	ctx, stop := signal.NotifyContext(ctx, signals...)
	defer stop()

	select {
	case _ = <-g.errCh:
		// TODO: log?
		return
	case <-ctx.Done():
		// TODO: log?
		_ = ctx.Err()
		return
	}
}

// Stop all runners concurrently and gracefully, blocking until all runners have
// stopped or [Group.StopTimeout] has elapsed.
//
// If a runner does not stop within the timeout, the context passed to
// [Runner.Stop] will return [GracefulShutdownTimeoutError] as the cause of the
// context cancellation.
func (g Group) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeoutCause(ctx, g.StopTimeout, GracefulShutdownTimeoutError{})
	defer cancel()

	var wg sync.WaitGroup
	for _, runner := range g.Runners {
		if runner == nil {
			continue
		}

		wg.Add(1)
		go func(r Runner) {
			defer wg.Done()
			if err := r.Stop(ctx); err != nil {
				_ = err // TODO: log?
			}
		}(runner)
	}

	wg.Wait()

	return nil
}
