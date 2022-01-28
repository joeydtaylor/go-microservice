package session

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"
)

func (u User) Get(r *http.Request) (UserContext, error) {

	uCtx := UserContext{}

	cookie, err := r.Cookie(os.Getenv("SESSION_COOKIE_NAME"))
	if err != nil {
		return UserContext{}, err
	}

	tr := &http.Transport{
		MaxIdleConns:       1,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}
	client := &http.Client{Transport: tr}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(5000)))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", os.Getenv("SESSION_STATE_API"), nil)
	req = req.WithContext(ctx)
	if err != nil {
		return UserContext{}, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.AddCookie(cookie)

	res, err := client.Do(req)
	if err != nil {
		return UserContext{}, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return UserContext{}, err
	}

	jsonErr := json.Unmarshal(body, &uCtx)
	if jsonErr != nil {
		return UserContext{}, jsonErr
	}

	return uCtx, nil

}

func (u User) IsAuthenticated(r *http.Request) bool {
	_, err := User{}.Get(r)
	return err == nil
}

func (u User) IsAuthorizedRole(r *http.Request, roles []Role) bool {
	userContext, err := User{}.Get(r)
	if err == nil && userContext.Role.Name == os.Getenv("ADMIN_ROLE_NAME") {
		return true
	}
	if err == nil {
		for _, i := range roles {
			if userContext.Role.Name == i.Name {
				return true
			}
		}
	}
	return false
}

func (u User) IsAuthorizedUser(r *http.Request, usernames []string) bool {
	userContext, err := User{}.Get(r)
	if err == nil && userContext.Username == os.Getenv("ADMIN_ROLE_NAME") {
		return true
	}
	if err == nil {
		for _, i := range usernames {
			if userContext.Username == i {
				return true
			}
		}
	}
	return false
}
