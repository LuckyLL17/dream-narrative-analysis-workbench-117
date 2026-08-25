package httpapi

import (
	"net/http"
	"time"
)

func (
	a *App,
) weeklyReport(
	writer http.ResponseWriter,
	request *http.Request,
) {
	user, _ := currentUser(
		request)
	report, err :=
		a.reports.Current(
			user.ID)
	if err != nil {
		writeError(writer, err)
		return
	}
	writeOK(writer, report)
}
func (
	a *App,
) reportHistory(
	writer http.ResponseWriter,
	request *http.Request,
) {
	user, _ := currentUser(
		request)
	writeOK(writer, a.reports.History(user.ID))
}

func (
	a *App,
) reportDigests(
	writer http.ResponseWriter,
	request *http.Request,
) {
	user, _ := currentUser(
		request)
	writeOK(writer, a.reports.HistoryDigests(user.ID))
}
func (
	a *App,
) refreshReport(
	writer http.ResponseWriter,
	request *http.Request,
) {
	user, _ := currentUser(
		request)
	report, err :=
		a.reports.Refresh(
			user.ID, time.Now().UTC().Round(time.Second))
	if err != nil {
		writeError(writer, err)
		return
	}
	writeOK(writer, report)
}
