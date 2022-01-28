package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/joeydtaylor/go-microservice/session"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {

		userContext, err := session.User{}.Get(r)
		if err != nil {
			log.Printf("%v", err)
		}

		if (session.User{}.IsAuthorizedRole(r, []session.Role{{Name: os.Getenv("ADMIN_ROLE_NAME")}})) {
			log.Println(userContext.Username, userContext.Role.Name, userContext.AuthenticationSource.Provider)
		}

	})

	log.Fatal(http.ListenAndServeTLS(os.Getenv("SERVER_LISTEN_ADDRESS"), os.Getenv("SSL_SERVER_CERTIFICATE"), os.Getenv("SSL_SERVER_KEY"), r))
}
