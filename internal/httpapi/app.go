package httpapi

import (
	"net/http"

	"dream117/internal/service"
)

type App struct {
	auth       *service.AuthService
	sessionSvc *service.SessionService
	dreams     *service.DreamService
	elements   *service.ElementService
	analysis   *service.AnalysisService
	reports    *service.ReportService
	exports    *service.ExportService
	health     *service.HealthService
}

func NewApp(
	auth *service.AuthService,
	session *service.SessionService,
	dreams *service.DreamService,
	elements *service.ElementService,
	analysis *service.AnalysisService,
	reports *service.ReportService,
	exports *service.ExportService,
	health *service.HealthService,
) *App {
	return &App{auth: auth, sessionSvc: session, dreams: dreams, elements: elements, analysis: analysis, reports: reports, exports: exports, health: health}
}

func (
	a *App,
) Handler(
	frontend http.Handler,
) http.Handler {
	return a.routes(frontend)
}
