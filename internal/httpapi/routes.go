package httpapi

import (
	"net/http"
	"strings"
)

func (
	a *App,
) routes(
	frontend http.Handler,
) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", a.healthCheck)
	mux.Handle("/api/v1/auth/register", method(http.MethodPost, a.register))
	mux.Handle("/api/v1/auth/login", method(http.MethodPost, a.login))
	a.mountProtected(mux, a.session, "/api/v1/session")
	mux.HandleFunc("/api/v1/auth/logout", a.logout)
	a.mountProtected(mux, a.dreamCollection, "/api/v1/dreams")
	a.mountProtected(mux, a.searchDreams, "/api/v1/dreams/search")
	a.mountProtected(mux, a.elementsList, "/api/v1/elements")
	a.mountProtected(mux, a.suggestions, "/api/v1/elements/suggestions")
	for _, path := range []string{"overview", "signals", "patterns", "facets", "review", "themes", "emotions", "wordcloud", "mood", "sleep-contrast", "theme-pairs", "theme-timeline", "calendar", "phrases", "narratives"} {
		a.mountProtected(mux, a.windowAnalysis, "/api/v1/analysis/"+path)
	}
	a.mountProtected(mux, a.lexicon, "/api/v1/lexicon")
	a.mountProtected(mux, a.weeklyReport, "/api/v1/reports/weekly")
	a.mountProtected(mux, a.reportHistory, "/api/v1/reports/history")
	a.mountProtected(mux, a.reportDigests, "/api/v1/reports/history/digests")
	a.mountProtected(mux, a.refreshReport, "/api/v1/reports/refresh")
	a.mountProtected(mux, a.exportMarkdown, "/api/v1/export/markdown")
	a.mountProtected(mux, a.exportJSON, "/api/v1/export/json")
	root := http.NewServeMux()
	root.Handle("/api/", a.routeWithDreamID(mux))
	root.Handle("/", frontend)
	return withLogging(withCORS(withBodyLimit(root, 2<<20)))
}

func method(
	want string,
	handler http.HandlerFunc,
) http.Handler {
	return wrap(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != want {
			methodNotAllowed(writer)
			return
		}
		handler(writer, request)
	})
}

func (a *App) mountProtected(
	mux *http.ServeMux,
	handler http.HandlerFunc,
	paths ...string,
) {
	protected := requireUser(a.sessionSvc, wrap(handler))
	for index := range paths {
		path := paths[index]
		mux.Handle(path, protected)
	}
}

func (a *App) protectedMethod(
	want string,
	handler http.HandlerFunc,
) http.Handler {
	return requireUser(a.sessionSvc, method(want, handler))
}
func (
	a *App,
) dreamCollection(
	writer http.ResponseWriter,
	request *http.Request,
) {
	switch request.Method {
	case http.MethodGet:
		a.listDreams(writer, request)
	case http.MethodPost:
		a.createDream(writer, request)
	default:
		methodNotAllowed(writer)
	}
}

func (
	a *App,
) routeWithDreamID(
	next http.Handler,
) http.Handler {
	return wrap(func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.URL.Path, "/api/v1/dreams/") {
			id := strings.TrimSpace(strings.TrimPrefix(request.URL.Path, "/api/v1/dreams/"))
			if id == "" || id == "search" {
				next.ServeHTTP(writer, request)
				return
			}
			if strings.HasSuffix(id, "/similar") {
				// similarity labels are resolved by the analysis layer so the API keeps boundary semantics
				dreamID := strings.TrimSpace(strings.TrimSuffix(id, "/similar"))
				a.protectedMethod(http.MethodGet, func(writer http.ResponseWriter, request *http.Request) {
					a.similarDreams(writer, request, dreamID)
				}).ServeHTTP(writer, request)
				return
			}
			if strings.HasSuffix(id, "/reflection") {
				dreamID := strings.TrimSuffix(id, "/reflection")
				a.protectedMethod(http.MethodGet, func(writer http.ResponseWriter, request *http.Request) {
					a.reflection(writer, request, dreamID)
				}).ServeHTTP(writer, request)
				return
			}
			requireUser(a.sessionSvc, wrap(a.dreamByID(id))).ServeHTTP(writer, request)
			return
		}
		next.ServeHTTP(writer, request)
	})
}

func (a *App) dreamByID(id string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			a.getDream(writer, request, id)
		case http.MethodPut:
			a.updateDream(writer, request, id)
		case http.MethodDelete:
			a.deleteDream(writer, request, id)
		default:
			methodNotAllowed(writer)
		}
	}
}
