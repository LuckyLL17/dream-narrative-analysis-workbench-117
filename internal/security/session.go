package security

import (
	"net/http"
	"strings"
)

const CookieName = "dream_session"

func SetCookie(
	writer http.ResponseWriter,
	token string,
	maxAge int,
) {
	http.SetCookie(writer, &http.Cookie{Name: CookieName, Value: token, Path: "/", MaxAge: maxAge, HttpOnly: true, SameSite: http.SameSiteLaxMode})
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
