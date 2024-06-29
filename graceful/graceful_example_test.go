package graceful_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"syscall"
	"time"

	"github.com/wafer-bw/go-toolbox/graceful"
)

func ExampleGroup() {
	ctx := context.Background()

	s1 := http.Server{Addr: ":1234"}
	s2 := http.Server{Addr: ":1235"}

	g := graceful.Group{
		Runners: []graceful.Runner{
			&graceful.RunnerType{
				StartFunc: func(_ context.Context) error { return s1.ListenAndServe() },
				StopFunc:  func(ctx context.Context) error { return s1.Shutdown(ctx) },
			},
			&graceful.RunnerType{
				StartFunc: func(_ context.Context) error { return s2.ListenAndServe() },
				StopFunc:  func(ctx context.Context) error { return s2.Shutdown(ctx) },
			},
		},
	}

	g.Start(ctx)

	go func() { // simulate a signal being sent to the process.
		time.Sleep(250 * time.Millisecond)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	waitErr := g.Wait(ctx, syscall.SIGINT, syscall.SIGTERM)
	stopErr := g.Stop(ctx, 5*time.Second)

	fmt.Println(waitErr) // wait returns nil when a signal is received.
	fmt.Println(stopErr)
	// Output:
	// <nil>
	// <nil>
}

func ExampleGroup_runnerStartError() {
	ctx := context.Background()

	s := http.Server{Addr: ":1235"}

	g := graceful.Group{
		Runners: []graceful.Runner{
			&graceful.RunnerType{
				StartFunc: func(_ context.Context) error { return errors.New("failed to start") },
				StopFunc:  func(ctx context.Context) error { return nil },
			},
			&graceful.RunnerType{
				StartFunc: func(_ context.Context) error { return s.ListenAndServe() },
				StopFunc:  func(ctx context.Context) error { return s.Shutdown(ctx) },
			},
		},
	}

	g.Start(ctx)
	waitErr := g.Wait(ctx, syscall.SIGINT, syscall.SIGTERM)
	stopErr := g.Stop(ctx, 1*time.Second)

	fmt.Println(waitErr) // wait returns the first error encountered by a runner.
	fmt.Println(stopErr)
	// Output:
	// failed to start
	// <nil>
}

func ExampleGroup_waitContextCancelled() {
	ctx, cancel := context.WithCancel(context.Background())

	s1 := http.Server{Addr: ":1234"}
	s2 := http.Server{Addr: ":1235"}

	g := graceful.Group{
		Runners: []graceful.Runner{
			&graceful.RunnerType{
				StartFunc: func(_ context.Context) error { return s1.ListenAndServe() },
				StopFunc:  func(ctx context.Context) error { return s1.Shutdown(ctx) },
			},
			&graceful.RunnerType{
				StartFunc: func(_ context.Context) error { return s2.ListenAndServe() },
				StopFunc:  func(ctx context.Context) error { return s2.Shutdown(ctx) },
			},
		},
	}

	g.Start(ctx)
	cancel() // cancel context wait uses.
	waitErr := g.Wait(ctx, syscall.SIGINT, syscall.SIGTERM)
	stopErr := g.Stop(ctx, 5*time.Second)

	fmt.Println(waitErr) // wait returns the context error if context is closed.
	fmt.Println(stopErr)
	// Output:
	// context canceled
	// <nil>
}

func ExampleGroup_runnerStopError() {
	ctx := context.Background()

	s := http.Server{Addr: ":1234"}

	g := graceful.Group{
		Runners: []graceful.Runner{
			&graceful.RunnerType{
				StartFunc: func(_ context.Context) error { return s.ListenAndServe() },
				StopFunc:  func(ctx context.Context) error { return s.Shutdown(ctx) },
			},
			&graceful.RunnerType{
				StartFunc: func(_ context.Context) error { return nil },
				StopFunc:  func(ctx context.Context) error { return errors.New("failed to stop") },
			},
		},
	}

	g.Start(ctx)

	go func() { // simulate a signal being sent to the process.
		time.Sleep(250 * time.Millisecond)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	waitErr := g.Wait(ctx, syscall.SIGINT, syscall.SIGTERM)
	stopErr := g.Stop(ctx, 1*time.Second)

	fmt.Println(waitErr)
	fmt.Println(stopErr)
	// Output:
	// <nil>
	// failed to stop
}
