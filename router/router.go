package router

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/joeydtaylor/go-microservice/controllers"
	"github.com/joeydtaylor/go-microservice/middleware/auth"
	"github.com/joeydtaylor/go-microservice/middleware/logger"
	"github.com/joeydtaylor/go-microservice/middleware/metrics"
	"go.uber.org/fx"
)

func ProtectedRoute(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if auth.ProvideAuthentication().IsAuthenticated(r.Context()) {
			h(w, r)
		} else {
			w.Write([]byte("Unauthorized"))
		}
	}
}

func RoleProtectedRoute(h http.HandlerFunc, roleName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if auth.ProvideAuthentication().IsRole(r.Context(), auth.Role{Name: roleName}) || auth.ProvideAuthentication().IsAdmin(r.Context()) {
			h(w, r)
		} else {
			w.Write([]byte("Forbidden"))
		}
	}
}

func AdminProtectedRoute(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if auth.ProvideAuthentication().IsRole(r.Context(), auth.Role{Name: os.Getenv("ADMIN_ROLE_NAME")}) {
			h(w, r)
		} else {
			w.Write([]byte("Forbidden"))
		}
	}
}

func UserProtectedRoute(h http.HandlerFunc, username string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if auth.ProvideAuthentication().IsUser(r.Context(), username) || auth.ProvideAuthentication().IsAdmin(r.Context()) {
			h(w, r)
		} else {
			w.Write([]byte("Forbidden"))
		}
	}
}

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
	r.Get("/", controllers.Index)
	r.Get("/protectedPage", ProtectedRoute(controllers.ProtectedPage))
	r.Get("/roleProtectedPage", RoleProtectedRoute(controllers.RoleProtectedPage, os.Getenv("DEVELOPER_ROLE_NAME")))
	r.Get("/adminProtectedPage", AdminProtectedRoute(controllers.AdminProtectedPage))

	return r
}

var Module = fx.Options(
	fx.Provide(ProvideRouter),
)
