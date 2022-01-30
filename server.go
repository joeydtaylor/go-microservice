package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/joeydtaylor/go-microservice/middleware/auth"
	"github.com/joeydtaylor/go-microservice/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	l := logger.NewLog()
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(middleware.Recoverer)
	r.Use(auth.Middleware())
	r.Use(logger.Middleware(l))

	if os.Getenv("DEFAULT_TIMEOUT_IN_SECONDS") != "" {
		if defaultTimeout, err := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_IN_SECONDS")); err == nil {
			r.Use(middleware.Timeout(time.Duration(defaultTimeout) * time.Second))
		}
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second * 2)
		if auth.IsAdmin(r.Context()) {
			w.WriteHeader(200)
			json.NewEncoder(w).Encode(fmt.Sprintf("Hello %s", auth.GetUser(r.Context()).Username))
		} else {
			w.WriteHeader(401)
			json.NewEncoder(w).Encode("Unauthorized")
		}
	})
	r.Handle("/metrics", logger.NewPromHttpHandler())

	if os.Getenv("SSL_SERVER_KEY") != "" && os.Getenv("SSL_SERVER_CERTIFICATE") != "" {
		log.Printf("Server listening at https://%v", os.Getenv("SERVER_LISTEN_ADDRESS"))
		log.Fatal(http.ListenAndServeTLS(os.Getenv("SERVER_LISTEN_ADDRESS"), os.Getenv("SSL_SERVER_CERTIFICATE"), os.Getenv("SSL_SERVER_KEY"), r))
	} else {
		log.Printf("Server listening at http://%v", os.Getenv("SERVER_LISTEN_ADDRESS"))
		log.Fatal(http.ListenAndServe(os.Getenv("SERVER_LISTEN_ADDRESS"), r))
	}

}
