package auth

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Role struct {
	Name string `json:"name"`
}

type AuthenticationSource struct {
	Provider string `json:"provider"`
}

type Username string

type User struct {
	Username             Username             `json:"username"`
	AuthenticationSource AuthenticationSource `json:"authenticationSource"`
	Role                 Role                 `json:"role"`
}

var userCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

// Middleware decodes the share session cookie, gets userContext and packs the session into context
func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			cookie, err := r.Cookie(os.Getenv("SESSION_COOKIE_NAME"))
			if err != nil || cookie == nil {
				next.ServeHTTP(w, r)
				return
			}

			user, err := validateSession(cookie, r)
			if err != nil {
				log.Printf("%v", err)
				http.Error(w, "Unauthorized", http.StatusForbidden)
			}

			ctx := context.WithValue(r.Context(), userCtxKey, user)
			ctxWithLog := context.WithValue(ctx, contextKey{"LogEntry"}, user)
			r = r.WithContext(ctxWithLog)

			next.ServeHTTP(w, r)

		})
	}
}

func validateSession(c *http.Cookie, r *http.Request) (User, error) {

	user := User{}

	tr := &http.Transport{
		MaxIdleConns:       1,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}

	client := &http.Client{Transport: tr}
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(time.Millisecond*time.Duration(5000)))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", os.Getenv("SESSION_STATE_API"), nil)
	if err != nil {
		return User{}, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.AddCookie(c)

	res, err := client.Do(req)
	if err != nil {
		return User{}, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return User{}, err
	}
	jsonErr := json.Unmarshal(body, &user)
	if jsonErr != nil {
		return User{}, jsonErr
	}

	return user, nil

}

/* forContext finds the user from the context. REQUIRES Middleware to have run. Not exported, please see func GetUser(ctx context.Context) User  */
func forContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userCtxKey).(User)
	if !ok {
		return user, false
	} else {
		return user, true
	}
}

/* Get user from context */
func GetUser(ctx context.Context) User {
	user, ok := ctx.Value(userCtxKey).(User)
	if !ok {
		return User{}
	} else {
		return user
	}
}

/* Validate user is Role{ Name: "" } */
func IsRole(ctx context.Context, role Role) bool {
	if user, ok := forContext(ctx); ok && user.Role == role {
		return true
	} else {
		return false
	}
}

/* Validate user is admin */
func IsAdmin(ctx context.Context) bool {
	if user, ok := forContext(ctx); ok && user.Role == (Role{Name: os.Getenv("ADMIN_ROLE_NAME")}) {
		return true
	} else {
		return false
	}
}

/* Validate user is Username */
func IsUser(ctx context.Context, u Username) bool {
	if user, ok := forContext(ctx); ok && user.Username == u {
		return true
	} else {
		return false
	}
}

/* Validate user is authenticated */
func IsAuthenticated(ctx context.Context) bool {
	if _, ok := forContext(ctx); ok {
		return true
	} else {
		return false
	}
}
