package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/joeydtaylor/go-microservice/auth"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	if os.Getenv("DEFAULT_TIMEOUT_IN_SECONDS") != "" {
		if defaultTimeout, err := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_IN_SECONDS")); err == nil {
			r.Use(middleware.Timeout(time.Duration(defaultTimeout) * time.Second))
		}
	}
	r.Use(auth.Middleware())

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if auth.IsAuthenticated(r.Context()) {
			log.Println("You are authenticated!!!")

			if auth.IsAdmin(r.Context()) {
				log.Println("You are an admin!!!")
			}

			log.Printf("You are the user %v!!!", auth.GetUser(r.Context()).Username)

		} else {
			log.Println("Whoops you are not authenticated!!!")
		}

	})

	if os.Getenv("SSL_SERVER_KEY") != "" && os.Getenv("SSL_SERVER_CERTIFICATE") != "" {
		log.Printf("Server listening at https://%v", os.Getenv("SERVER_LISTEN_ADDRESS"))
		log.Fatal(http.ListenAndServeTLS(os.Getenv("SERVER_LISTEN_ADDRESS"), os.Getenv("SSL_SERVER_CERTIFICATE"), os.Getenv("SSL_SERVER_KEY"), r))
	} else {
		log.Printf("Server listening at http://%v", os.Getenv("SERVER_LISTEN_ADDRESS"))
		log.Fatal(http.ListenAndServe(os.Getenv("SERVER_LISTEN_ADDRESS"), r))
	}

}
