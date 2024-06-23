package probe_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wafer-bw/go-toolbox/probe"
)

func TestGroup_ProbeAll(t *testing.T) {
	t.Parallel()

	t.Run("returns a map & true value when all startup probes are successful", func(t *testing.T) {
		t.Parallel()

		probes := probe.Group{
			"p1": probe.ProberFunc(func(ctx context.Context) error { return nil }),
			"p2": probe.ProberFunc(func(ctx context.Context) error { return nil }),
		}

		results, ok := probes.ProbeAll(context.Background())
		require.True(t, ok)
		require.Equal(t, map[string]error{"p1": nil, "p2": nil}, results)
	})

	t.Run("returns map indicating failed probes & false when one or more startup probes fail", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")
		startupProbes := probe.Group{
			"p1": probe.ProberFunc(func(ctx context.Context) error { return err }),
			"p2": probe.ProberFunc(func(ctx context.Context) error { return nil }),
			"p3": probe.ProberFunc(func(ctx context.Context) error { return err }),
		}

		results, ok := startupProbes.ProbeAll(context.Background())
		require.False(t, ok)
		require.Equal(t, map[string]error{"p1": err, "p2": nil, "p3": err}, results)
	})
}

func TestGroup_ServeHTTP(t *testing.T) {
	t.Parallel()

	t.Run("returns an ok status when all startup probes are successful", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		h := probe.Group{
			"p1": probe.ProberFunc(func(ctx context.Context) error { return nil }),
			"p2": probe.ProberFunc(func(ctx context.Context) error { return nil }),
		}
		expectedStatus := http.StatusOK
		expectedResponse := fmt.Sprintf(`{"p1":"%s","p2":"%s"}`, probe.OkState, probe.OkState)

		h.ServeHTTP(w, r)
		body := w.Body.String()
		require.Equal(t, expectedStatus, w.Code, body)
		require.Equal(t, expectedResponse, strings.TrimSpace(body))
	})

	t.Run("returns a service unavailable status when one or more startup probes fail", func(t *testing.T) {
		t.Parallel()

		err := errors.New("failed")
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		h := probe.Group{
			"p1": probe.ProberFunc(func(ctx context.Context) error { return err }),
			"p2": probe.ProberFunc(func(ctx context.Context) error { return nil }),
			"p3": probe.ProberFunc(func(ctx context.Context) error { return err }),
		}
		expectedStatus := http.StatusServiceUnavailable
		expectedResponse := fmt.Sprintf(`{"p1":"%s","p2":"%s","p3":"%s"}`, err, probe.OkState, err)

		h.ServeHTTP(w, r)
		body := w.Body.String()
		require.Equal(t, expectedStatus, w.Code, body)
		require.Equal(t, expectedResponse, strings.TrimSpace(body))
	})
}
