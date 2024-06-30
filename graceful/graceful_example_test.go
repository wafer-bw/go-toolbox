package graceful_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/wafer-bw/go-toolbox/graceful"
)

// RunnerServer is an example type that satisfies the graceful.Runner interface.
// In this case it is a simple wrapper around an http.Server but it could be
// a more complex type of your own design that happens to have Start and Stop
// methods.
type RunnerServer struct {
	http.Server
}

func (r *RunnerServer) Start(ctx context.Context) error {
	return r.ListenAndServe()
}

func (r *RunnerServer) Stop(ctx context.Context) error {
	return r.Shutdown(ctx)
}

// Demonstrates how to use [Group.Run] in a more realistic real-world scenario
// than the examples provided for [Group] as a whole.
//
// Remember to adjust the shutdownTimeout and exitSignals to suit your
// application's needs.
func ExampleGroup_Run() {
	ctx := context.Background()

	shutdownTimeout := 250 * time.Millisecond
	exitSignals := []os.Signal{syscall.SIGINT, syscall.SIGTERM}

	metricsServer := &RunnerServer{http.Server{Addr: ":8001"}}
	applicationServer := &RunnerServer{http.Server{Addr: ":8002"}}

	runners := graceful.Group{metricsServer, applicationServer}

	if err := runners.Run(ctx, shutdownTimeout, exitSignals...); err != nil {
		log.Println(err)
	}
}

// Demonstrates how to use the graceful package in a simple real-world scenario.
func ExampleGroup() {
	ctx := context.Background()

	s1 := http.Server{Addr: ":8001"}
	s2 := http.Server{Addr: ":8002"}

	g := graceful.Group{
		&graceful.RunnerType{
			StartFunc: func(_ context.Context) error { return s1.ListenAndServe() },
			StopFunc:  func(ctx context.Context) error { return s1.Shutdown(ctx) },
		},
		&graceful.RunnerType{
			StartFunc: func(_ context.Context) error { return s2.ListenAndServe() },
			StopFunc:  func(ctx context.Context) error { return s2.Shutdown(ctx) },
		},
	}

	go func() { // simulate a signal being sent to the process.
		time.Sleep(250 * time.Millisecond)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	defer func() {
		if err := g.Stop(ctx, 5*time.Second); err != nil {
			log.Println(err)
		}
	}()

	if err := g.Start(ctx, syscall.SIGINT, syscall.SIGTERM); err != nil {
		log.Println(err)
	}

	// Output:
}

// VisibleStages prints out the different stages of a group runner's lifecycle.
func ExampleGroup_visibleStages() {
	ctx := context.Background()

	s := http.Server{Addr: ":8000"}

	g := graceful.Group{
		&graceful.RunnerType{
			StartFunc: func(_ context.Context) error {
				fmt.Println("starting server")
				// this doesn't get captured by startErr because it won't happen
				// until Stop is called at which point the group is no longer
				// capturing start errors.
				err := s.ListenAndServe()
				fmt.Println("server has stopped listening:", err)
				return err
			},
			StopFunc: func(ctx context.Context) error {
				fmt.Println("gracefully stopping server")
				err := s.Shutdown(ctx)
				fmt.Println("server gracefully stopped")
				return err
			},
		},
	}

	go func() { // simulate a signal being sent to the process.
		time.Sleep(250 * time.Millisecond)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	startErr := g.Start(ctx, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("start error:", startErr) // nil here indicates a signal was received.

	stopErr := g.Stop(ctx, 2*time.Second)
	fmt.Println("stop error:", stopErr) // nil here indicates everything shutdown gracefully.
	// Output:
	// starting server
	// start error: <nil>
	// gracefully stopping server
	// server has stopped listening: http: Server closed
	// server gracefully stopped
	// stop error: <nil>
}

// RunnerStartError demonstrates the behavior of a group when at least one
// [Runner] fails to start.
func ExampleGroup_runnerStartError() {
	ctx := context.Background()

	s1 := http.Server{Addr: "-1:8001"}
	s2 := http.Server{Addr: ":18002"}
	s3 := http.Server{Addr: "-1:8003"}

	g := graceful.Group{
		&graceful.RunnerType{
			StartFunc: func(_ context.Context) error { return s1.ListenAndServe() },
			StopFunc:  func(ctx context.Context) error { return s1.Shutdown(ctx) },
		},
		&graceful.RunnerType{
			StartFunc: func(_ context.Context) error { return s2.ListenAndServe() },
			StopFunc:  func(ctx context.Context) error { return s2.Shutdown(ctx) },
		},
		&graceful.RunnerType{
			StartFunc: func(_ context.Context) error { return s3.ListenAndServe() },
			StopFunc:  func(ctx context.Context) error { return s3.Shutdown(ctx) },
		},
	}

	startErr := g.Start(ctx, syscall.SIGINT, syscall.SIGTERM)
	stopErr := g.Stop(ctx, 1*time.Second)

	fmt.Println(startErr) // wait returns the first error encountered by a runner.
	fmt.Println(stopErr)
	// Output:
	// listen tcp: lookup -1: no such host
	// <nil>
}

// RunnerStopError demonstrates the behavior of a group when at least one
// [Runner] fails to stop.
func ExampleGroup_runnerStopError() {
	ctx := context.Background()

	s1 := http.Server{Addr: ":8001"}
	s2 := http.Server{Addr: ":18002"}
	s3 := http.Server{Addr: ":8003"}

	g := graceful.Group{
		&graceful.RunnerType{
			StartFunc: func(_ context.Context) error { return s1.ListenAndServe() },
			StopFunc: func(ctx context.Context) error {
				_ = s1.Shutdown(ctx)
				return fmt.Errorf("failed to stop")
			},
		},
		&graceful.RunnerType{
			StartFunc: func(_ context.Context) error { return s2.ListenAndServe() },
			StopFunc:  func(ctx context.Context) error { return s2.Shutdown(ctx) },
		},
		&graceful.RunnerType{
			StartFunc: func(_ context.Context) error { return s3.ListenAndServe() },
			StopFunc: func(ctx context.Context) error {
				_ = s3.Shutdown(ctx)
				return fmt.Errorf("failed to stop")
			},
		},
	}

	go func() { // simulate a signal being sent to the process.
		time.Sleep(250 * time.Millisecond)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	startErr := g.Start(ctx, syscall.SIGINT, syscall.SIGTERM)
	stopErr := g.Stop(ctx, 1*time.Second)

	fmt.Println(startErr)
	fmt.Println(stopErr)
	// Output:
	// <nil>
	// failed to stop
}

// StartContextCancelled demonstrates the behavior of a group when the context
// passed to [Group.Start] is cancelled.
func ExampleGroup_startContextCancelled() {
	ctx, cancel := context.WithCancel(context.Background())

	s1 := http.Server{Addr: ":8001"}
	s2 := http.Server{Addr: ":8002"}

	g := graceful.Group{
		&graceful.RunnerType{
			StartFunc: func(_ context.Context) error { return s1.ListenAndServe() },
			StopFunc:  func(ctx context.Context) error { return s1.Shutdown(ctx) },
		},
		&graceful.RunnerType{
			StartFunc: func(_ context.Context) error { return s2.ListenAndServe() },
			StopFunc:  func(ctx context.Context) error { return s2.Shutdown(ctx) },
		},
	}

	cancel() // cancel context wait uses.
	startErr := g.Start(ctx, syscall.SIGINT, syscall.SIGTERM)
	stopErr := g.Stop(ctx, 5*time.Second)

	fmt.Println(startErr) // wait returns the context error if context is closed.
	fmt.Println(stopErr)
	// Output:
	// context canceled
	// <nil>
}
