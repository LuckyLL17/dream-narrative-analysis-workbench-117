package httpapi

import (
	"net/http"
	"strings"
	"time"

	"dream117/internal/domain"
	"dream117/internal/service"
)

func (
	a *App,
) listDreams(
	writer http.ResponseWriter,
	request *http.Request,
) {
	user, _ := currentUser(
		request)
	to := queryTime(request, "to", time.Now().UTC())
	from := queryTime(request, "from", to.AddDate(0, 0, -90))
	query := strings.TrimSpace(request.URL.Query().Get("q"))
	writeJSON(writer, http.StatusOK, a.dreams.List(user.ID, from, to, query))
}

func (
	a *App,
) searchDreams(
	writer http.ResponseWriter,
	request *http.Request,
) {
	user, _ := currentUser(
		request)
	page, pageSize := queryInt(request, "page", 1), queryInt(request, "page_size", 24)
	filter := domain.DreamFilter{
		From:           queryTime(request, "from", time.Time{}),
		To:             queryTime(request, "to", time.Time{}),
		Query:          strings.TrimSpace(request.URL.Query().Get("q")),
		Emotion:        domain.Emotion(strings.TrimSpace(request.URL.Query().Get("emotion"))),
		Theme:          strings.TrimSpace(request.URL.Query().Get("theme")),
		MinimumClarity: queryInt(request, "min_clarity", 0),
		MaximumSleep:   queryFloat(request, "max_sleep", 0),
		RememberedOnly: queryBool(request, "remembered_only"),
		Page:           page,
		PageSize:       pageSize,
	}
	writeJSON(writer, http.StatusOK, a.analysis.Search(user.ID, filter))
}
func (
	a *App,
) createDream(
	writer http.ResponseWriter,
	request *http.Request,
) {
	user, _ := currentUser(
		request)
	var input service.DreamInput
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, err)
		return
	}
	d, err := a.dreams.Create(
		user.ID, input)
	if err != nil {
		writeError(writer, err)
		return
	}
	writeCreated(writer, d)
}
func (
	a *App,
) updateDream(
	writer http.ResponseWriter,
	request *http.Request,
	id string,
) {
	user, _ := currentUser(
		request)
	var input service.DreamInput
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, err)
		return
	}
	d, err := a.dreams.Update(
		user.ID, id, input)
	if err != nil {
		writeError(writer, err)
		return
	}
	writeOK(writer, d)
}
func (
	a *App,
) getDream(
	writer http.ResponseWriter,
	request *http.Request,
	id string,
) {
	user, _ := currentUser(
		request)
	d, err := a.dreams.Get(
		user.ID, id)
	if err != nil {
		writeError(writer, err)
		return
	}
	writeOK(writer, d)
}

func (
	a *App,
) similarDreams(
	writer http.ResponseWriter,
	request *http.Request,
	id string,
) {
	user, _ := currentUser(
		request)
	items, err :=
		a.analysis.Similar(
			user.ID, id, queryInt(request, "limit", 8))
	if err != nil {
		writeError(writer, err)
		return
	}
	writeOK(writer, items)
}

func (
	a *App,
) reflection(
	writer http.ResponseWriter,
	request *http.Request,
	id string,
) {
	user, _ := currentUser(
		request)
	report, err :=
		a.analysis.Reflection(
			user.ID, id)
	if err != nil {
		writeError(writer, err)
		return
	}
	writeOK(writer, report)
}
func (
	a *App,
) deleteDream(
	writer http.ResponseWriter,
	request *http.Request,
	id string,
) {
	user, _ := currentUser(
		request)
	if err := a.dreams.Delete(user.ID, id); err != nil {
		writeError(writer, err)
		return
	}
	writeMessage(writer, http.StatusOK, "梦境已删除")
}
