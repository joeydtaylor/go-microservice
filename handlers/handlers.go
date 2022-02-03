package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/joeydtaylor/go-microservice/middleware/auth"
)

func GetIndex(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Second * 2)
	if auth.IsAdmin(r.Context()) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(fmt.Sprintf("Hello %s", auth.GetUser(r.Context()).Username))
	} else {
		w.WriteHeader(401)
	}
}
