package httpapi

import "net/http"

func (
	a *App,
) exportMarkdown(
	writer http.ResponseWriter,
	request *http.Request,
) {
	user, _ := currentUser(
		request)
	writer.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	writer.Header().Set("Content-Disposition", "attachment; filename=dreams.md")
	_, _ = writer.Write([]byte(a.exports.Markdown(user.ID)))
}
func (
	a *App,
) exportJSON(
	writer http.ResponseWriter,
	request *http.Request,
) {
	user, _ := currentUser(
		request)
	content, err :=
		a.exports.JSON(
			user.ID)
	if err != nil {
		writeError(writer, err)
		return
	}
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Header().Set("Content-Disposition", "attachment; filename=dreams.json")
	_, _ = writer.Write(content)
}
