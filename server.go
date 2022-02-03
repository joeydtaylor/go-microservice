package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joeydtaylor/go-microservice/router"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	r := router.NewRouter()

	if os.Getenv("SSL_SERVER_KEY") != "" && os.Getenv("SSL_SERVER_CERTIFICATE") != "" {
		log.Printf("Server listening at https://%v", os.Getenv("SERVER_LISTEN_ADDRESS"))
		log.Fatal(http.ListenAndServeTLS(os.Getenv("SERVER_LISTEN_ADDRESS"), os.Getenv("SSL_SERVER_CERTIFICATE"), os.Getenv("SSL_SERVER_KEY"), r))
	} else {
		log.Printf("Server listening at http://%v", os.Getenv("SERVER_LISTEN_ADDRESS"))
		log.Fatal(http.ListenAndServe(os.Getenv("SERVER_LISTEN_ADDRESS"), r))
	}

}
