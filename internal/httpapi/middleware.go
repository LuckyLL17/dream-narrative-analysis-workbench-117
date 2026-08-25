package httpapi

import (
	"context"
	"log"
	"net/http"
	"time"

	"dream117/internal/domain"
	"dream117/internal/security"
	"dream117/internal/service"
)

type contextKey string

const userKey contextKey = "dream-user"

func wrap(handler http.HandlerFunc) http.Handler {
	return handler
}

func methodNotAllowed(writer http.ResponseWriter) {
	writer.WriteHeader(http.StatusMethodNotAllowed)
}

func withCORS(
	next http.Handler,
) http.Handler {
	return wrap(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
func withLogging(
	next http.Handler,
) http.Handler {
	return wrap(func(writer http.ResponseWriter, request *http.Request) {
		started := time.Now()
		next.ServeHTTP(writer, request)
		log.Printf("http method=%s path=%s duration=%s", request.Method, request.URL.Path, time.Since(started))
	})
}
func withBodyLimit(
	next http.Handler,
	max int64,
) http.Handler {
	return wrap(func(writer http.ResponseWriter, request *http.Request) {
		request.Body =
			http.MaxBytesReader(
				writer, request.Body, max)
		next.ServeHTTP(writer, request)
	})
}

func requireUser(
	auth *service.SessionService,
	next http.Handler,
) http.Handler {
	return wrap(func(writer http.ResponseWriter, request *http.Request) {
		token := security.ReadToken(request)
		user, err := auth.Resolve(
			token)
		if err != nil {
			writeError(writer, domain.ErrUnauthorized)
			return
		}
		next.ServeHTTP(writer, request.WithContext(context.WithValue(request.Context(), userKey, user)))
	})
}
func currentUser(
	request *http.Request,
) (domain.User, bool) {
	user, ok := request.Context().Value(userKey).(domain.User)
	return user, ok
}
