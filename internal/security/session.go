package security

import (
	"net/http"
	"strings"
	"time"
)

const CookieName = "dream_session"

func SetCookie(
	writer http.ResponseWriter,
	token string,
	maxAge int,
) {
	cookie := &http.Cookie{Name: CookieName, Value: token, Path: "/", MaxAge: maxAge, HttpOnly: true, SameSite: http.SameSiteLaxMode}
	if maxAge > 0 {
		cookie.Expires = time.Now().UTC().Add(time.Duration(maxAge) * time.Second)
	}
	http.SetCookie(writer, cookie)
}

func ClearCookie(
	writer http.ResponseWriter,
) {
	SetCookie(writer, "", -1)
}

func ReadToken(
	request *http.Request,
) string {
	if cookie, err := request.Cookie(CookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	header := request.Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	}
	return ""
}
