package main

import (
	"log"
	"net/http"
	"os"

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
	r.Use(middleware.Logger)
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

	log.Fatal(http.ListenAndServeTLS(os.Getenv("SERVER_LISTEN_ADDRESS"), os.Getenv("SSL_SERVER_CERTIFICATE"), os.Getenv("SSL_SERVER_KEY"), r))
}
