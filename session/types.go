package session

import (
	"net/http"
)

type Role struct {
	Name string `json:"name"`
}

type AuthenticationSource struct {
	Provider string `json:"provider"`
}

type UserContext struct {
	Username             string               `json:"username"`
	AuthenticationSource AuthenticationSource `json:"authenticationSource"`
	Role                 Role                 `json:"role"`
}

type User struct{}

type UserContextGetter interface {
	Get(r *http.Request) (UserContext, error)
}

type Authenticator interface {
	IsAuthenticated(r *http.Request) bool
}

type Authorizer interface {
	IsAuthorizedRole(r *http.Request, roles []Role) bool
	IsAuthorizedUser(r *http.Request, usernames []string) bool
}
