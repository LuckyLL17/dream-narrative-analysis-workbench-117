package telemetry

import (
	"net/http"
	"sync/atomic"
)

type Metrics struct {
	requests atomic.Uint64
	errors   atomic.Uint64
}

func (
	m *Metrics,
) Observe(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		m.requests.Add(1)
		next.ServeHTTP(writer, request)
	})
}
func (m *Metrics) Snapshot() map[string]uint64 {
	return map[string]uint64{"requests": m.requests.Load(), "errors": m.errors.Load()}
}
func (m *Metrics) Error() { m.errors.Add(1) }
