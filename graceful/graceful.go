// Package graceful provides mechanisms for starting and stopping groups of
// services, primarily used to accomplish a graceful shutdown.
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

// Group of [Runner] that should run concurrently together via a single call
// point and stop gracefully should one of them encounter an error or the
// application receive a signal.
type Group []Runner

// Run is a convenience method that calls [Group.Start] & [Group.Stop] in
// sequence returning the error (if any) from [Group.Start] and ignoring the
// error (if any) from [Group.Stop].
func (g Group) Run(ctx context.Context, timeout time.Duration, signals ...os.Signal) error {
	defer g.Stop(ctx, timeout) //nolint:errcheck // intentionally ignored.
	return g.Start(ctx, signals...)
}

// Start all [Runner] concurrently, blocking until either a Runner.Start call
// encounters an error, one of the provided signals is received via
// [signal.NotifyContext], or the context provided to it is canceled, then
// returns the first non-nil error (if any) or nil if a signal was received.
//
// An error returned from Start does not indicate that all runners have stopped,
// you must call [Group.Stop] to stop all runners.
func (g Group) Start(ctx context.Context, signals ...os.Signal) error {
	eg, errCtx := errgroup.WithContext(ctx)
	signalCtx, stop := signal.NotifyContext(ctx, signals...)
	defer stop()

	for _, r := range g {
		if r == nil {
			continue
		}
		eg.Go(func() error { return r.Start(ctx) })
	}

	select {
	case <-errCtx.Done():
		return context.Cause(errCtx)
	case <-signalCtx.Done():
		return ctx.Err()
	}
}

// Stop all [Runner] concurrently, blocking until all Runner.Stop calls have
// returned, then returns the first non-nil error (if any) from them.
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
		eg.Go(func() error { return r.Stop(ctx) })
	}

	return eg.Wait()
}

// RunnerType is an adapter type to allow the use of ordinary start and stop
// functions as a [Runner].
//   - A nil StartFunc will immediately return nil.
//   - A nil StopFunc will immediately return nil.
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
