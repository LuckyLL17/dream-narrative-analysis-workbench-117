package httpapi

import (
	"net/http"

	"dream117/internal/domain"
	"dream117/internal/security"
)

type authRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (
	a *App,
) register(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var input authRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, err)
		return
	}
	user, token, err := a.auth.Register(input.Email, input.Name, input.Password)
	if err != nil {
		writeError(writer, err)
		return
	}
	security.SetCookie(writer, token, 86400)
	writeJSON(writer, http.StatusCreated, map[string]interface{}{"user": publicUser(user), "token": token})
}
func (
	a *App,
) login(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var input authRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, err)
		return
	}
	user, token, err := a.auth.Login(input.Email, input.Password)
	if err != nil {
		writeError(writer, err)
		return
	}
	security.SetCookie(writer, token, 86400)
	writeJSON(writer, http.StatusOK, map[string]interface{}{"user": publicUser(user), "token": token})
}
func (
	a *App,
) session(
	writer http.ResponseWriter,
	request *http.Request,
) {
	user, ok := currentUser(
		request)
	if !ok {
		writeError(writer, securityError())
		return
	}
	writeJSON(writer, http.StatusOK, publicUser(user))
}
func (
	a *App,
) logout(
	writer http.ResponseWriter,
	request *http.Request,
) {
	security.ClearCookie(writer)
	writeMessage(writer, http.StatusOK, "已退出")
}
func publicUser(
	user domain.User,
) map[string]interface{} {
	return map[string]interface{}{"id": user.ID, "email": user.Email, "display_name": user.DisplayName, "created_at": user.CreatedAt}
}
func securityError() error { return errUnauthorized{} }

type errUnauthorized struct{}

func (errUnauthorized) Error() string { return "未授权" }
