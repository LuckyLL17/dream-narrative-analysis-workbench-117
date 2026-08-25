package httpapi

import "net/http"

func (
	a *App,
) elementsList(
	writer http.ResponseWriter,
	request *http.Request,
) {
	user, _ := currentUser(
		request)
	writeJSON(writer, http.StatusOK, a.elements.List(user.ID))
}
