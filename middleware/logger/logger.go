package logger

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/joeydtaylor/go-microservice/middleware/auth"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewLog() *zap.Logger {

	cfg := zap.NewProductionEncoderConfig()
	cfg.MessageKey = zapcore.OmitKey

	consoleDebugging := zapcore.Lock(os.Stdout)

	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   string(os.Getenv("LOG_FILE_PATH")),
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})
	core := zapcore.NewTee(zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg),
		w,
		zap.InfoLevel,
	), zapcore.NewCore(zapcore.NewJSONEncoder(cfg), consoleDebugging, zap.InfoLevel))

	return zap.New(core)

}

func Middleware(l *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var sessionCookie string

			cookie, err := r.Cookie(os.Getenv("SESSION_COOKIE_NAME"))
			if err == nil {
				sessionCookie = cookie.Value
			}

			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			defer l.Sync()

			l.Info("", zap.String("dateTime", time.Now().UTC().Format(time.RFC1123)),
				zap.String("requestId", middleware.GetReqID(r.Context())),
				zap.String("httpScheme", scheme),
				zap.Bool("isAuthenticated", auth.IsAuthenticated(r.Context())),
				zap.String("sessionCookie", sessionCookie),
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
