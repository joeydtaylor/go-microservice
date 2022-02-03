package router

import (
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/joeydtaylor/go-microservice/handlers"
	"github.com/joeydtaylor/go-microservice/middleware/auth"
	"github.com/joeydtaylor/go-microservice/middleware/logger"
)

func NewRouter() *chi.Mux {
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
	r.Use(auth.Middleware())
	r.Use(logger.Request())

	r.Get("/", handlers.GetIndex)
	r.Handle("/metrics", logger.NewPromHttpHandler())

	return r
}
