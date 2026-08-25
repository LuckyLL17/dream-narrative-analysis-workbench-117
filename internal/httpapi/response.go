package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"dream117/internal/domain"
	"dream117/pkg/jsonutil"
)

func writeJSON(
	writer http.ResponseWriter,
	status int,
	value interface{},
) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	content, _ := jsonutil.Marshal(map[string]interface{}{"data": value})
	_, _ = writer.Write(append(content, '\n'))
}

func writeOK(writer http.ResponseWriter, value interface{}) {
	writeJSON(writer, http.StatusOK, value)
}

func writeCreated(writer http.ResponseWriter, value interface{}) {
	writeJSON(writer, http.StatusCreated, value)
}
func writeError(
	writer http.ResponseWriter,
	err error,
) {
	status := errorStatus(err)
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	content, _ := jsonutil.Marshal(map[string]string{"error": err.Error()})
	_, _ = writer.Write(append(content, '\n'))
}

func errorStatus(err error) int {
	for candidate, status := range map[error]int{
		domain.ErrUnauthorized: http.StatusUnauthorized,
		domain.ErrNotFound:     http.StatusNotFound,
		domain.ErrConflict:     http.StatusConflict,
	} {
		if errors.Is(err, candidate) {
			return status
		}
	}
	return http.StatusBadRequest
}
func writeMessage(
	writer http.ResponseWriter,
	status int,
	message string,
) {
	writeJSON(writer, status, map[string]string{"message": message})
}
func isJSON(
	request *http.Request,
) bool {
	return strings.Contains(request.Header.Get("Content-Type"), "application/json")
}
