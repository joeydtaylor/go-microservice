package auth

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"go.uber.org/fx"
)

type Role struct {
	Name string `json:"name"`
}

type AuthenticationSource struct {
	Provider string `json:"provider"`
}

type User struct {
	Username             string               `json:"username"`
	AuthenticationSource AuthenticationSource `json:"authenticationSource"`
	Role                 Role                 `json:"role"`
}

var userCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

type Middleware struct{}

func ProvideAuthentication() Middleware {
	return Middleware{}
}

// Middleware decodes the share session cookie, gets userContext and packs the session into context
func (Middleware) Middleware() func(http.Handler) http.Handler {
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
			r = r.WithContext(ctx)

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

/* Get user from context */
func (Middleware) GetUser(ctx context.Context) User {
	user, ok := ctx.Value(userCtxKey).(User)
	if !ok {
		return User{}
	} else {
		return user
	}
}

/* Validate user is Role{ Name: "" } */
func (Middleware) IsRole(ctx context.Context, role Role) bool {
	if user, ok := ctx.Value(userCtxKey).(User); ok && user.Role == role {
		return true
	} else {
		return false
	}
}

/* Validate user is admin */
func (Middleware) IsAdmin(ctx context.Context) bool {
	if user, ok := ctx.Value(userCtxKey).(User); ok && user.Role == (Role{Name: os.Getenv("ADMIN_ROLE_NAME")}) {
		return true
	} else {
		return false
	}
}

/* Validate user is Username */
func (Middleware) IsUser(ctx context.Context, u string) bool {
	if user, ok := ctx.Value(userCtxKey).(User); ok && user.Username == u {
		return true
	} else {
		return false
	}
}

/* Validate user is authenticated */
func (Middleware) IsAuthenticated(ctx context.Context) bool {
	if _, ok := ctx.Value(userCtxKey).(User); ok {
		return true
	} else {
		return false
	}
}

var Module = fx.Options(
	fx.Provide(ProvideAuthentication),
)
