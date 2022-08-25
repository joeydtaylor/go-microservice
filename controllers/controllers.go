package controllers

import (
	"log"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, from the Index Controller!"))
}

func ProtectedPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, from the ProtectedPage Controller!"))
}

func RoleProtectedPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, from the RoleProtectedPage Controller!"))
}

func AdminProtectedPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, from the AdminProtectedPage Controller!"))
	userContext := r.Context().Value("username")
	log.Println(userContext)
}
