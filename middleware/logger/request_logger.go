package logger

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/joeydtaylor/go-microservice/middleware/auth"
	"github.com/joeydtaylor/go-microservice/utils"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Middleware struct{}

func ProvideLoggerMiddleware() Middleware {
	return Middleware{}
}

func ProvideLogger() *zap.Logger {
	return NewLog("system.log")
}

func NewLog(n string) *zap.Logger {

	cfg := zap.NewProductionEncoderConfig()
	cfg.MessageKey = zapcore.OmitKey
	consoleDebugging := zapcore.Lock(os.Stdout)
	var logPath string

	if runtime.GOOS == "windows" {
		logPath = fmt.Sprintf("%s\\%s", "log", n)
	} else {
		logPath = fmt.Sprintf("%s/%s", "log", n)
	}

	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    50,
		MaxBackups: 3,
		MaxAge:     7,
	})
	core := zapcore.NewTee(zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg),
		w,
		zap.InfoLevel,
	), zapcore.NewCore(zapcore.NewJSONEncoder(cfg), consoleDebugging, zap.InfoLevel))

	l := zap.New(core)

	return l

}

func (Middleware) Middleware(ca auth.Middleware) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := NewLog("http-access.log")

			ww := utils.NewWrapResponseWriter(w, r.ProtoMajor)
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
					zap.String("requestId", utils.GetReqID(r.Context())),
					zap.String("httpScheme", scheme),
					zap.Bool("isAuthenticated", ca.IsAuthenticated(r.Context())),
					zap.String("sessionCookie", sessionCookie),
					zap.String("username", ca.GetUser(r.Context()).Username),
					zap.String("role", ca.GetUser(r.Context()).Role.Name),
					zap.String("authenticationProvider", string(ca.GetUser(r.Context()).AuthenticationSource.Provider)),
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

			}()

			next.ServeHTTP(ww, r)

		})
	}
}

var Module = fx.Options(
	fx.Provide(ProvideLoggerMiddleware),
	fx.Provide(ProvideLogger),
)
