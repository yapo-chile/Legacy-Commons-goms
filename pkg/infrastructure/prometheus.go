package infrastructure

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusHandler provides both, a way to instrument http.HandlerFunc with
// Prometheus, and a Prometheus http.Handler that can receive scrape requests
// from a central server
type PrometheusHandler struct {
	counter      *prometheus.CounterVec
	duration     prometheus.ObserverVec
	inFlight     prometheus.Gauge
	requestSize  prometheus.ObserverVec
	responseSize prometheus.ObserverVec
}

// MakePrometheusHandler Builds a fresh PrometheusHandler, initializing its
// metrics
func MakePrometheusHandler() PrometheusHandler {
	h := PrometheusHandler{
		counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_requests_total",
				Help: "A counter for requests to the wrapped handler.",
			},
			[]string{"code", "method"},
		),
		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "request_duration_seconds",
				Help:    "A histogram of latencies for requests.",
				Buckets: []float64{.25, .5, 1, 2.5, 5, 10},
			},
			[]string{"handler", "method"},
		),
		inFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "in_flight_requests",
			Help: "A gauge of requests currently being served by the wrapped handler.",
		}),
		requestSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "request_size_bytes",
				Help:    "A histogram of request sizes for requests.",
				Buckets: []float64{50, 100, 200, 500, 1000, 1500},
			},
			[]string{},
		),
		responseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "response_size_bytes",
				Help:    "A histogram of response sizes for requests.",
				Buckets: []float64{200, 500, 900, 1500},
			},
			[]string{},
		),
	}

	// Register all of the metrics in the standard registry.
	prometheus.MustRegister(h.counter, h.duration, h.inFlight, h.requestSize, h.responseSize)
	return h
}

// TrackHandlerFunc instruments handler with Prometheus, adding every
// configured metric
func (h *PrometheusHandler) TrackHandlerFunc(pattern string, handler http.HandlerFunc) http.HandlerFunc {
	return promhttp.InstrumentHandlerInFlight(h.inFlight,
		promhttp.InstrumentHandlerCounter(h.counter,
			promhttp.InstrumentHandlerDuration(h.duration.MustCurryWith(prometheus.Labels{"handler": pattern}),
				promhttp.InstrumentHandlerRequestSize(h.requestSize,
					promhttp.InstrumentHandlerResponseSize(h.responseSize, handler),
				),
			),
		),
	).ServeHTTP
}

// Handler returns an http.Handler suitable to handle prometheus scraping
// requests. Usually it's run under /metrics URL
func (h *PrometheusHandler) Handler() http.Handler {
	return promhttp.Handler()
}
