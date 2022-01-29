package logger

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/joeydtaylor/go-microservice/middleware/auth"
	"go.uber.org/zap"
)

func Middleware(l *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			defer l.Sync()

			l.Info("received", zap.String("dateTime", time.Now().UTC().Format(time.RFC1123)),
				zap.String("requestId", middleware.GetReqID(r.Context())),
				zap.String("httpScheme", scheme),
				zap.Bool("isAuthenticated", auth.IsAuthenticated(r.Context())),
				zap.String("username", string(auth.GetUser(r.Context()).Username)),
				zap.String("role", string(auth.GetUser(r.Context()).Role.Name)),
				zap.String("authenticationProvider", string(auth.GetUser(r.Context()).AuthenticationSource.Provider)),
				zap.String("httpProto", r.Proto),
				zap.String("httpMethod", r.Method),
				zap.String("remoteAddr", r.RemoteAddr),
				zap.String("uri", fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)))

			next.ServeHTTP(w, r)

		})
	}
}
