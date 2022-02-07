package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joeydtaylor/go-microservice/bundlefx"
	"github.com/joeydtaylor/go-microservice/router"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func registerHooks(
	lifecycle fx.Lifecycle,
	logger *zap.Logger,
	r *chi.Mux,
) {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				err := godotenv.Load()
				if err != nil {
					log.Fatalf("Error loading .env file: %v", err)
				}
				if os.Getenv("SSL_SERVER_KEY") != "" && os.Getenv("SSL_SERVER_CERTIFICATE") != "" {
					l := logger.With(zap.Field(zap.String("event", ("Server listening at https://" + os.Getenv("SERVER_LISTEN_ADDRESS")))))
					l.Info("")
					log.Fatal(http.ListenAndServeTLS(os.Getenv("SERVER_LISTEN_ADDRESS"), os.Getenv("SSL_SERVER_CERTIFICATE"), os.Getenv("SSL_SERVER_KEY"), r))
				} else {
					l := logger.With(zap.Field(zap.String("event", ("Server listening at http://" + os.Getenv("SERVER_LISTEN_ADDRESS")))))
					l.Info("")
					log.Fatal(http.ListenAndServe(os.Getenv("SERVER_LISTEN_ADDRESS"), r))
				}
				return nil
			},
			OnStop: func(ctx context.Context) error {
				l := logger.With(zap.Field(zap.String("event", "Server stopped")))
				l.Info("")
				return nil
			},
		},
	)
}

func main() {
	fx.New(
		bundlefx.Module,
		router.Module,
		fx.Invoke(registerHooks),
	).Start(context.Background())
}
