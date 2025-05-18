// Package graceful provides mechanisms for asynchronously starting &
// synchronously stopping trees of long running tasks enabling graceful
// shutdowns.
//
// This package handles the common case of starting several things in parallel
// then stopping them gracefully in series. It does not act as a full directed
// acyclic graph (https://en.wikipedia.org/wiki/Directed_acyclic_graph).
package graceful

import (
	"cmp"
	"context"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sync/errgroup"
)

// Runner is capable of starting and stopping itself.
type Runner interface {
	// Start must terminate when the passed context is canceled, [Runner.Stop]
	// is called, or it completes without error (whichever happens first). It
	// must not panic if [Runner.Stop] was called first.
	Start(context.Context) error

	// Stop must terminate when the passed context is canceled or it completes
	// without error (whichever happens first). It must not panic if
	// [Runner.Start] was not called.
	Stop(context.Context) error
}

type RunOption func(*RunConfig)

func WithStopTimeout(d time.Duration) RunOption {
	return func(cfg *RunConfig) {
		cfg.stopTimeout = d
	}
}

func WithStopSignals(signals ...os.Signal) RunOption {
	return func(cfg *RunConfig) {
		cfg.signals = signals
	}
}

func WithStoppingCh(ch chan<- struct{}) RunOption {
	return func(cfg *RunConfig) {
		cfg.stoppingCh = ch
	}
}

type RunConfig struct {
	stopTimeout time.Duration
	signals     []os.Signal
	stoppingCh  chan<- struct{}
}

// Group of [Runner] which can be started in parallel & stopped in series.
//
// Group satisfies [Runner] and thus it can be nested within itself to create
// a tree.
type Group []Runner

// Start all [Runner] in parallel. Blocks until all [Runner.Start] have returned
// normally, then returns the first non-nil errror (if any) from them.
func (g Group) Start(ctx context.Context) error {
	eg := new(errgroup.Group)
	for _, r := range g {
		if r == nil {
			continue
		}
		eg.Go(func() error { return r.Start(ctx) })
	}

	return eg.Wait()
}

// Stop all [Runner] in series. Blocks until all [Runner.Stop] have returned
// normally, then returns the first non-nil errror (if any) from them.
func (g Group) Stop(ctx context.Context) error {
	var firstErr error
	for _, r := range g {
		if r == nil {
			continue
		}
		if err := r.Stop(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// Run starts all [Runner] in parallel and stops them in series.
//
// Stopping is initiated when any of the following occurs:
//   - the passed context is canceled
//   - a signal passed via [WithStopSignals] is received
//   - a [Runner.Start] returns an error
//
// When stopping is initiated, the channel passed via [WithStoppingCh] will be
// closed. It will use the timeout passed via [WithStopTimeout] as the deadline
// for the [context.Context] passed to each [Runner.Stop].
//
// The first encountered error (either [Runner.Start] error,
// [context.Context.Err], or [Runner.Stop] error) will be returned. However, all
// [Runner.Stop] are guaranteed to be called.
func (g Group) Run(ctx context.Context, opts ...RunOption) error {
	cfg := &RunConfig{}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(cfg)
	}

	var startErr, runErr error
	startErrCh := make(chan error)
	go func() {
		if err := g.Start(ctx); err != nil {
			startErrCh <- err
		}
	}()

	signalCh := make(chan os.Signal, 1)
	if len(cfg.signals) != 0 {
		signal.Notify(signalCh, cfg.signals...)
	}
	select {
	case <-signalCh:
		// received signal
	case err := <-startErrCh:
		startErr = err
	case <-ctx.Done():
		runErr = ctx.Err()
	}
	signal.Stop(signalCh)

	if cfg.stoppingCh != nil {
		close(cfg.stoppingCh)
	}

	stopCtx, cancel := context.WithTimeout(ctx, cfg.stopTimeout)
	if cfg.stopTimeout == 0 {
		stopCtx = ctx
	}
	defer cancel()

	stopErr := g.Stop(stopCtx)

	return cmp.Or(startErr, stopErr, runErr)
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
