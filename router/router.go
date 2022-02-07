package router

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/joeydtaylor/go-microservice/middleware/auth"
	"github.com/joeydtaylor/go-microservice/middleware/logger"
	"github.com/joeydtaylor/go-microservice/middleware/metrics"
	"go.uber.org/fx"
)

func ProvideRouter(l logger.Middleware, a auth.Middleware, m http.Handler) *chi.Mux {
	r := chi.NewRouter()
	if os.Getenv("DEFAULT_TIMEOUT_IN_SECONDS") != "" {
		if defaultTimeout, err := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_IN_SECONDS")); err == nil {
			r.Use(middleware.Timeout(time.Duration(defaultTimeout) * time.Second))
		}
	}

	r.Use(middleware.RequestID)
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(middleware.Recoverer)
	r.Use(a.Middleware())
	r.Use(metrics.Collect(a))
	r.Use(l.Middleware(a))
	r.Handle("/metrics", m)

	return r
}

var Module = fx.Options(
	fx.Provide(ProvideRouter),
)
