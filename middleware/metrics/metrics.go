package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/joeydtaylor/go-microservice/middleware/auth"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"
)

var (
	responseTime = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "response_time",
			Help:    "http response time.",
			Buckets: []float64{0.5, 1, 5, 10, 30, 60},
		},
	)

	totalHttpRequestsFromRole = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "total_http_requests_from_role",
		Help: "http requests from role",
	},
		[]string{"role"})

	totalHttpRequestsToUri = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "total_http_requests_to_uri",
		Help: "http requests to uri",
	},
		[]string{"code", "uri", "method"})

	totalHttpRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "total_http_requests",
		Help: "http requests by code, and method",
	},
		[]string{"code", "method"})
)

func Collect(ca auth.Middleware) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			startTime := time.Now()

			defer func() {

				endTime := time.Since(startTime)

				if r.RequestURI != "/metrics" {
					defer func() {
						totalHttpRequestsFromRole.With(prometheus.Labels{"role": ca.GetUser(r.Context()).Role.Name}).Inc()
						totalHttpRequestsToUri.With(prometheus.Labels{"code": strconv.Itoa(ww.Status()), "uri": fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI), "method": r.Method}).Inc()
						totalHttpRequests.With(prometheus.Labels{"code": strconv.Itoa(ww.Status()), "method": r.Method}).Inc()
						responseTime.Observe(endTime.Seconds())
					}()
				}

			}()

			next.ServeHTTP(ww, r)

		})
	}
}

func NewPromHttpHandler() http.Handler {
	return promhttp.Handler()
}

func ProvideMetrics() http.Handler {
	return NewPromHttpHandler()
}

func init() {
	prometheus.MustRegister(
		responseTime,
		totalHttpRequestsFromRole,
		totalHttpRequestsToUri,
		totalHttpRequests,
	)
}

var Module = fx.Options(
	fx.Provide(ProvideMetrics),
)
