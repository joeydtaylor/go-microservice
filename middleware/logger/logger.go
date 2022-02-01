package logger

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/joeydtaylor/go-microservice/middleware/auth"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
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

func init() {
	prometheus.MustRegister(
		responseTime,
		totalHttpRequestsFromRole,
		totalHttpRequestsToUri,
		totalHttpRequests,
	)
}

func NewPromHttpHandler() http.Handler {
	return promhttp.Handler()
}

func NewLog() *zap.Logger {

	logFileMaxSizeInMb, err := strconv.Atoi(os.Getenv("LOG_FILE_MAX_SIZE_IN_MB"))
	if err != nil {
		log.Panic(err)
	}
	logFileMaxBackups, err := strconv.Atoi(os.Getenv("LOG_FILE_MAX_BACKUPS"))
	if err != nil {
		log.Panic(err)
	}
	logFileMaxAgeInDays, err := strconv.Atoi(os.Getenv("LOG_FILE_MAX_AGE_IN_DAYS"))
	if err != nil {
		log.Panic(err)
	}

	cfg := zap.NewProductionEncoderConfig()
	cfg.MessageKey = zapcore.OmitKey
	consoleDebugging := zapcore.Lock(os.Stdout)

	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   fmt.Sprintf("%s\\%s", os.Getenv("LOG_DIRECTORY"), os.Getenv("LOG_FILE_NAME")),
		MaxSize:    logFileMaxSizeInMb,
		MaxBackups: logFileMaxBackups,
		MaxAge:     logFileMaxAgeInDays,
	})
	core := zapcore.NewTee(zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg),
		w,
		zap.InfoLevel,
	), zapcore.NewCore(zapcore.NewJSONEncoder(cfg), consoleDebugging, zap.InfoLevel))

	l := zap.New(core)

	return l

}

func Middleware(l *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			var sessionCookie string
			cookie, err := r.Cookie(os.Getenv("SESSION_COOKIE_NAME"))
			if err == nil {
				sessionCookie = cookie.Value
			}
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			startTime := time.Now()
			defer r.Body.Close()
			body, bodyReadErr := io.ReadAll(r.Body)
			defer func() {

				endTime := time.Since(startTime)

				log := l.With(
					zap.String("dateTime", startTime.UTC().Format(time.RFC1123)),
					zap.String("requestId", middleware.GetReqID(r.Context())),
					zap.String("httpScheme", scheme),
					zap.Bool("isAuthenticated", auth.IsAuthenticated(r.Context())),
					zap.String("sessionCookie", sessionCookie),
					zap.String("username", auth.GetUser(r.Context()).Username),
					zap.String("role", auth.GetUser(r.Context()).Role.Name),
					zap.String("authenticationProvider", string(auth.GetUser(r.Context()).AuthenticationSource.Provider)),
					zap.String("httpProto", r.Proto),
					zap.String("httpMethod", r.Method),
					zap.String("remoteAddr", r.RemoteAddr),
					zap.String("uri", fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)),
					zap.Duration("lat", endTime),
					zap.Int("responseSize", ww.BytesWritten()),
					zap.Int("status", ww.Status()))

				if r.Method == http.MethodPost {
					log.Info("", zap.ByteString("requestData", body))
				}
				if bodyReadErr != nil {
					log.Error("", zap.NamedError("Error", bodyReadErr))
					ww.WriteHeader(500)
				}
				log.Info("")

				if r.RequestURI != "/metrics" {
					defer func() {
						totalHttpRequestsFromRole.With(prometheus.Labels{"role": auth.GetUser(r.Context()).Role.Name}).Inc()
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
