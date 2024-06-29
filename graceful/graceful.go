// Package graceful provides a way to run a group of goroutines and gracefully
// stop them via context cancellation or signals.
package graceful

import (
	"context"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sync/errgroup"
)

// Runner is capable of starting and stopping itself.
type Runner interface {
	// Start must either complete within the lifetime of the context passed to
	// it respecting the context's deadline or terminate when Stop is called.
	Start(context.Context) error

	// Stop must complete within the lifetime of the context passed to it
	// respecting the context's deadline.
	Stop(context.Context) error
}

// Group is used to run multiple [Runner]s concurrently and eventually
// gracefully stop them.
type Group []Runner

// Start all runners concurrently, then blocks until either Runner.Start call
// encounters an error, one of the provided signals is received via
// [signal.NotifyContext], or the context provided to it is canceled returning
// the first encountered error or nil if a signal was received.
func (g Group) Start(ctx context.Context, signals ...os.Signal) error {
	eg, errCtx := errgroup.WithContext(ctx)
	signalCtx, stop := signal.NotifyContext(ctx, signals...)
	defer stop()

	for _, r := range g {
		if r == nil {
			continue
		}
		r := r
		eg.Go(func() error { return r.Start(ctx) })
	}

	select {
	case <-errCtx.Done():
		return context.Cause(errCtx)
	case <-signalCtx.Done():
		return ctx.Err()
	}
}

// Stop blocks until all Runner.Stop calls have returned, then returns the first
// non-nil (if any) from them.
//
// If a Runner.Stop does not complete before timeout the context passed to
// it will cancel with [ShutdownTimeoutError] as the [context.Cause].
func (g Group) Stop(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeoutCause(ctx, timeout, ShutdownTimeoutError{})
	defer cancel()

	eg := new(errgroup.Group)
	for _, r := range g {
		if r == nil {
			continue
		}
		r := r
		eg.Go(func() error { return r.Stop(ctx) })
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
