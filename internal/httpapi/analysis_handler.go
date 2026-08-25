package httpapi

import (
	"net/http"
	"strings"

	"dream117/internal/analysis"
)

func (
	a *App,
) window(
	request *http.Request,
) analysis.Window {
	days := queryInt(request, "days", 30)
	window, err :=
		analysis.ParseWindow(
			request.URL.Query().Get("from"), request.URL.Query().Get("to"), days)
	if err != nil {
		return analysis.DefaultWindow(days)
	}
	return window
}

func (
	a *App,
) windowAnalysis(
	writer http.ResponseWriter,
	request *http.Request,
) {
	user, _ := currentUser(
		request)
	path := strings.TrimPrefix(request.URL.Path, "/api/v1/analysis/")
	writeJSON(writer, http.StatusOK, a.analysis.Analyze(user.ID, path, a.window(request)))
}
func (
	a *App,
) suggestions(
	writer http.ResponseWriter,
	request *http.Request,
) {
	user, _ := currentUser(
		request)
	title, content :=
		request.URL.Query().Get("title"), request.URL.Query().Get("content")
	writeJSON(writer, http.StatusOK, a.elements.Suggestions(user.ID, title, content))
}

func (
	a *App,
) lexicon(
	writer http.ResponseWriter,
	request *http.Request,
) {
	writeJSON(writer, http.StatusOK, a.analysis.Lexicon(request.URL.Query().Get("q")))
}
