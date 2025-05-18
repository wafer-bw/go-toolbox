package graceful_test

import (
	"context"
	"errors"
	"net/http"
	"syscall"
	"time"

	"github.com/wafer-bw/go-toolbox/graceful"
)

func ExampleGroup_Run() {
	ctx := context.TODO()

	stoppingCh := make(chan struct{})
	readinessProbe := func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-stoppingCh:
			return errors.New("stopping or stopped")
		default:
			return nil
		}
	}

	s := http.Server{
		Addr: ":0",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := readinessProbe(r.Context()); err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
		}),
	}

	g := graceful.Group{
		graceful.RunnerType{
			StartFunc: func(ctx context.Context) error {
				return s.ListenAndServe()
			},
			StopFunc: func(ctx context.Context) error {
				if err := s.Shutdown(ctx); err != nil {
					_ = s.Close()
					return err
				}
				return nil
			},
		},
		// TODO: populate with more runners.
	}

	if err := g.Run(ctx,
		graceful.WithStopSignals(syscall.SIGTERM, syscall.SIGINT),
		graceful.WithStopTimeout(1*time.Minute),
		graceful.WithStoppingCh(stoppingCh),
	); err != nil {
		panic(err) // TODO: handle error appropriately.
	}
}
