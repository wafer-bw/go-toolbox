// Package probe provides types for use as health checks such as kubernetes'
// startup, liveness, & readiness probes.
package probe

import (
	"context"
	"encoding/json"
	"net/http"
)

// OkState is the string representation of a successful probe used in the
// response body of [Group.ServeHTTP].
const OkState string = "ok"

// Prober is capable of probing its state for problems.
//
// A prober is typically used as part of a kubernetes startup, liveness, or
// readiness probe. A returned error means the probe failed its check and is
// in an unhealthy state. A nil error means the probe succeeded and is in a
// healthy state.
type Prober interface {
	Probe(ctx context.Context) error
}

// ProberFunc is an adapter type to allow the use of ordinary functions as
// Probers. If f is a function with the appropriate signature, ProberFunc(f) is
// a [Prober] that calls f.
type ProberFunc func(ctx context.Context) error

// Probe calls f(ctx).
func (f ProberFunc) Probe(ctx context.Context) error {
	return f(ctx)
}

// Group of Probers to probe.
type Group map[string]Prober

// ProbeAll executes Probe() on each [Prober] returning a map of results and a
// bool representing whether or not all probes were successful.
func (g Group) ProbeAll(ctx context.Context) (map[string]error, bool) {
	ok := true
	results := make(map[string]error, len(g))
	for probeKey, probe := range g {
		err := probe.Probe(ctx)
		if err != nil {
			ok = false
		}
		results[probeKey] = err
	}

	return results, ok
}

// ServeHTTP probes the state of all [Prober] in the group.
//
// Response status codes:
//   - 200 OK                    (all probes were successful)
//   - 500 Internal Server Error (JSON encoding of the response failed)
//   - 503 Service Unavailable   (one or more probes failed)
//
// The response body is a JSON encoded map[string]string where the key is the
// probe name and the value is the probe's outcome. A probe state of [OkState]
// means the probe was successful, otherwise the value is the string
// representation of the error returned by the probe.
func (g Group) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := http.StatusOK

	results, ok := g.ProbeAll(ctx)
	if !ok {
		status = http.StatusServiceUnavailable
	}

	friendlyResults := make(map[string]string, len(results))
	for key, err := range results {
		if err != nil {
			friendlyResults[key] = err.Error()
			continue
		}
		friendlyResults[key] = OkState
	}

	reply, err := json.Marshal(friendlyResults)
	if err != nil {
		status = http.StatusInternalServerError
		reply = []byte(err.Error())
	}

	w.WriteHeader(status)
	_, _ = w.Write(reply)
}
