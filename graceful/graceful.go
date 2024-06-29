// TODO: docstring
// TODO: examples
// TODO: handle nil errCh in Wait & Stop
// TODO: expose encountered errors to callers
// TODO: close errCh after all runners have stopped?
package graceful

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"
)

// DefaultShutdownTimeout is the default time to wait for all runners to
// gracefully stop after [Group.Stop] is called.
const DefaultShutdownTimeout = 30 * time.Second

// Runner is capable of starting and stopping itself
type Runner interface {
	Start(context.Context) error

	// Stop must complete within the lifetime of the context passed to it,
	// respecting the context's deadline.
	Stop(context.Context) error
}

// Group is used to run multiple [Runner]s concurrently and eventually
// gracefully stop them.
type Group struct {
	Runners []Runner

	// StopTimeout is the maximum time to wait for all runners to gracefully
	// stop after Group.Stop is called. If a runner does not stop within this
	// time, the context passed to Runner.Stop will be canceled.
	//
	// If StopTimeout is zero or less DefaultShutdownTimeout is used.
	StopTimeout time.Duration

	errCh  chan error
	waited bool // used for testing purposes only.
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
				g.errCh <- err // TODO: expose back to caller somehow.
			}
		}(runner)
	}
}

// Wait blocks until a runner encounters an error or a signal is received.
func (g *Group) Wait(ctx context.Context, signals ...os.Signal) {
	defer func() { g.waited = true }()
	ctx, stop := signal.NotifyContext(ctx, signals...)
	defer stop()

	select {
	case err := <-g.errCh:
		_ = err // TODO: expose back to caller somehow.
		return
	case <-ctx.Done():
		_ = ctx.Err() // TODO: expose back to caller somehow?
		return
	}
}

// Stop all runners concurrently and gracefully, blocking until all runners have
// stopped or Group.StopTimeout has elapsed.
//
// If a runner does not stop within the timeout the context passed to
// Runner.Stop will return [ShutdownTimeoutError] as the cause of the
// context cancellation.
func (g *Group) Stop(ctx context.Context) {
	timeout := g.StopTimeout
	if timeout <= 0 {
		timeout = DefaultShutdownTimeout
	}

	ctx, cancel := context.WithTimeoutCause(ctx, timeout, ShutdownTimeoutError{})
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
				_ = err // TODO: expose back to caller somehow.
			}
		}(runner)
	}

	wg.Wait()
}

// Run starts all runners concurrently, waits until one of the runners returns
// an error or a signal is received, then stops all runners gracefully.
//
// It is a convenience method that calls [Group.Start], [Group.Wait], and
// [Group.Stop] in sequence.
func (g *Group) Run(ctx context.Context, signals ...os.Signal) {
	g.Start(ctx)
	g.Wait(ctx, signals...)
	g.Stop(ctx)
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
