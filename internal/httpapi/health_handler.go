package httpapi

import "net/http"

func (
	a *App,
) healthCheck(
	writer http.ResponseWriter,
	request *http.Request,
) {
	writeJSON(writer, http.StatusOK, a.health.Status())
}
