// Package graceful v2 provides mechanisms for starting and stopping groups of
// services, primarily used to accomplish a graceful shutdown.
//
// TODO: Update all docstrings indicating differences between v1 and v2.
// TODO: Add example & testing for recursive use of groups.
// TODO: Document that v1 is not superceded by v2, it's just a diff use case.
// TODO: Update Start tests to verify concurrent vs sequential behavior.
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
type Group struct {
	Runners []Runner

	// StartupSequentially will start all runners in sequence, blocking at each
	// each Runner's Start method before moving to the next.
	StartupSequentially bool

	// ShutdownSignals are the signals that will trigger a graceful shutdown.
	ShutdownSignals []os.Signal

	// ShutdownTimeout is the maximum time allowed for all runners to stop.
	ShutdownTimeout time.Duration

	// ShutdownSequentially will stop all runners in sequence, blocking at each
	// Runner's Stop method before moving to the next.
	ShutdownSequentially bool

	// ShutdownReversed can be applied when ShutdownSequentially is true. It
	// will cause Stop to traverse the Runners slice in reverse while shutting
	// down.
	ShutdownReversed bool
}

// Run is a convenience method that calls [Group.Start] & [Group.Stop] in
// sequence returning the error (if any) from [Group.Start] and ignoring the
// error (if any) from [Group.Stop].
func (g Group) Run(ctx context.Context) error {
	defer g.Stop(ctx) //nolint:errcheck // intentionally ignored.
	return g.Start(ctx)
}

// Start all [Runner] concurrently, blocking until either a Runner.Start call
// encounters an error, one of the provided signals is received via
// [signal.NotifyContext], or the context provided to it is canceled, then
// returns the first non-nil error (if any) or nil if a signal was received.
//
// An error returned from Start does not indicate that all runners have stopped,
// you must call [Group.Stop] to stop all runners.
func (g Group) Start(ctx context.Context) error {
	eg, errCtx := errgroup.WithContext(ctx)
	signalCtx, stop := signal.NotifyContext(ctx, g.ShutdownSignals...)
	defer stop()

	if g.StartupSequentially {
		eg.Go(func() error { return g.sequentialStart(ctx) })
	} else {
		g.concurrentStart(errCtx, eg)
	}

	select {
	case <-errCtx.Done():
		return context.Cause(errCtx)
	case <-signalCtx.Done():
		return ctx.Err()
	}
}

func (g Group) concurrentStart(ctx context.Context, eg *errgroup.Group) {
	for _, r := range g.Runners {
		if r == nil {
			continue
		}
		r := r
		eg.Go(func() error { return r.Start(ctx) })
	}
}

func (g Group) sequentialStart(ctx context.Context) error {
	var firstErr error
	for _, r := range g.Runners {
		if r == nil {
			continue
		}

		if err := r.Start(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// Stop all [Runner], blocking until all Runner.Stop calls have returned, then
// returns the first non-nil error (if any) from them.
//
// If a Runner.Stop does not complete before timeout the context passed to
// it will cancel with [ShutdownTimeoutError] as the [context.Cause].
func (g Group) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeoutCause(ctx, g.ShutdownTimeout, ShutdownTimeoutError{})
	defer cancel()

	if g.ShutdownSequentially {
		return g.sequentialStop(ctx)
	}
	return g.concurrentStop(ctx)
}

func (g Group) concurrentStop(ctx context.Context) error {
	eg := new(errgroup.Group)
	for _, r := range g.Runners {
		if r == nil {
			continue
		}
		r := r
		eg.Go(func() error { return r.Stop(ctx) })
	}
	return eg.Wait()
}

func (g Group) sequentialStop(ctx context.Context) error {
	var firstErr error
	for i := 0; i < len(g.Runners); i++ {
		var r Runner
		if g.ShutdownReversed {
			r = g.Runners[len(g.Runners)-1-i]
		} else {
			r = g.Runners[i]
		}

		if r == nil {
			continue
		}

		if err := r.Stop(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
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
