// Package graceful provides a way to run a group of goroutines and gracefully
// stop them via context cancellation or signals.
//
// TODO: examples
// TODO: close errCh after all runners have stopped?
package graceful

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sync/errgroup"
)

// Runner is capable of starting and stopping itself.
type Runner interface {
	// Start must complete within the lifetime of the context passed to it,
	// respecting the context's deadline.
	Start(context.Context) error

	// Stop must complete within the lifetime of the context passed to it,
	// respecting the context's deadline.
	Stop(context.Context) error
}

// Group is used to run multiple [Runner]s concurrently and eventually
// gracefully stop them.
type Group struct {
	Runners []Runner
	errCh   chan error
}

// Start all runners concurrently.
func (g *Group) Start(ctx context.Context) {
	g.errCh = make(chan error)

	for _, runner := range g.Runners {
		if runner == nil {
			continue
		}

		go func(r Runner) {
			if err := r.Start(ctx); err != nil {
				select { // TODO: use errgroup.Group instead?
				case g.errCh <- err:
					return
				default:
					return
				}
			}
		}(runner)
	}
}

// Wait blocks until a runner encounters an error or one of the provided signals
// is received, then returns the first non-nil error encountered by
// a runner or nil.
func (g *Group) Wait(ctx context.Context, signals ...os.Signal) error {
	ctx, stop := signal.NotifyContext(ctx, signals...)
	defer stop()

	select {
	case err := <-g.errCh:
		return err
	case <-ctx.Done():
		fmt.Println("wait ending bc context done")
		fmt.Println(ctx.Err())
		fmt.Println(context.Cause(ctx))
		return nil // TODO: return the ctx err?
	}
}

// Stop blocks until all Runner.Stop calls have returned, then returns the first
// non-nil (if any) from them.
//
// If a Runner.Stop does not complete before timeout the context passed to
// it will cancel with [ShutdownTimeoutError] as the [context.Cause].
func (g *Group) Stop(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeoutCause(ctx, timeout, ShutdownTimeoutError{})
	defer cancel()

	eg := new(errgroup.Group)
	for _, runner := range g.Runners {
		if runner == nil {
			continue
		}

		runner := runner
		eg.Go(func() error { return runner.Stop(ctx) })
	}

	return eg.Wait()
}

// RunnerType is an adapter type to allow the use of ordinary start and stop
// functions as a [Runner].
//
// A nil StartFunc will be a no-op start function.
// A nil StopFunc will be a no-op stop function.
type RunnerType struct {
	StartFunc func(context.Context) error
	StopFunc  func(context.Context) error
}

func (r RunnerType) Start(ctx context.Context) error {
	if r.StartFunc == nil {
		return nil
	}
	return r.StartFunc(ctx)
}

func (r RunnerType) Stop(ctx context.Context) error {
	if r.StopFunc == nil {
		return nil
	}
	return r.StopFunc(ctx)
}

// ShutdownTimeoutError is set as the context cause when a [Runner] is
// unable to complete Runner.Stop before Group.StopTimeout during execution
// of [Group.Stop].
type ShutdownTimeoutError struct{}

func (ShutdownTimeoutError) Error() string {
	return "graceful shutdown timed out"
}

// Timeout returns true if the error is a timeout error, this allows callers
// to identify the nature of the error without needing to match the error
// based on equality.
func (ShutdownTimeoutError) Timeout() bool {
	return true
}
