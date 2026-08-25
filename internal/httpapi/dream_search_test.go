package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"dream117/internal/domain"
	"dream117/internal/security"
	"dream117/internal/service"
	"dream117/internal/store"
	"dream117/web"
)

// newTestApp wires the real HTTP layer (routes, middleware, jsonutil response
// shaping) to an in-memory store so the pagination/date contract is exercised
// end to end: HTTP params -> AnalysisService.Search -> DreamFilter.Normalized ->
// store.SearchPage.
func newTestApp(t *testing.T) (*App, *store.Store) {
	t.Helper()
	data, err := store.Open(filepath.Join(t.TempDir(), "dreams.json"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	codec := security.NewTokenCodec("test-secret", time.Hour)
	auth := service.NewAuthService(data, codec)
	session := service.NewSessionService(auth)
	elements := service.NewElementService(data)
	dreams := service.NewDreamService(data, elements)
	analysisService := service.NewAnalysisService(data)
	reports := service.NewReportService(data)
	exports := service.NewExportService(data)
	health := service.NewHealthService(data)
	return NewApp(auth, session, dreams, elements, analysisService, reports, exports, health), data
}

func registerForTest(t *testing.T, app *App, email string) (domain.User, string) {
	t.Helper()
	user, token, err := app.auth.Register(email, "分页用户", "password1234")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	return user, token
}

func createDreamForTest(t *testing.T, app *App, userID string, title string, date time.Time) {
	t.Helper()
	input := service.DreamInput{
		Title:      title,
		Content:    "这是一段关于分页与日期边界的梦境描述",
		DreamDate:  date,
		WakeTime:   date.Add(time.Hour),
		SleepHours: 7.5,
		Clarity:    6,
		Emotion:    string(domain.EmotionCalm),
	}
	// Seed directly through the service so the test focuses on the search
	// contract (HTTP params -> service -> filter -> store) rather than auth.
	if _, err := app.dreams.Create(userID, input); err != nil {
		t.Fatalf("create dream %q: %v", title, err)
	}
}

// TestSearchDreamsHTTPContract is the end-to-end regression for the reported
// defect: searching "分页" over 2026-08-01..08-02 with page_size=2 page=2 must
// return total=3, total_pages=2, and the single oldest match "分页一" — and the
// record at the very end of the end day must participate.
func TestSearchDreamsHTTPContract(t *testing.T) {
	app, _ := newTestApp(t)
	user, token := registerForTest(t, app, "pager@example.com")

	day := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	createDreamForTest(t, app, user.ID, "分页一", day.Add(10*time.Hour))
	createDreamForTest(t, app, user.ID, "分页二", day.Add(11*time.Hour))
	createDreamForTest(t, app, user.ID, "分页三", clockDayEndUTC(day))

	handler := app.Handler(web.Handler())
	url := "/api/v1/dreams/search?q=" + url.QueryEscape("分页") +
		"&from=2026-08-01&to=2026-08-02&page=2&page_size=2"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var envelope struct {
		Data struct {
			Items []struct {
				Title     string `json:"title"`
				DreamDate string `json:"dream_date"`
			} `json:"items"`
			Page        int  `json:"page"`
			PageSize    int  `json:"page_size"`
			Total       int  `json:"total"`
			TotalPages  int  `json:"total_pages"`
			HasNext     bool `json:"has_next"`
			HasPrevious bool `json:"has_previous"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rec.Body.String())
	}
	page := envelope.Data
	if page.Total != 3 {
		t.Errorf("expected total=3, got total=%d", page.Total)
	}
	if page.TotalPages != 2 {
		t.Errorf("expected total_pages=2, got total_pages=%d", page.TotalPages)
	}
	if page.Page != 2 {
		t.Errorf("expected echoed page=2, got page=%d", page.Page)
	}
	if page.PageSize != 2 {
		t.Errorf("expected page_size=2, got page_size=%d", page.PageSize)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected 1 item on page 2, got %d (%+v)", len(page.Items), page.Items)
	}
	if got, want := page.Items[0].Title, "分页一"; got != want {
		t.Errorf("expected page 2 item %q (oldest after descending sort), got %q", want, got)
	}
	if page.HasNext {
		t.Errorf("expected has_next=false on the last page")
	}
	if !page.HasPrevious {
		t.Errorf("expected has_previous=true on page 2")
	}
}

// TestSearchDreamsHTTPFirstPageFull guards the per-page slice count: page 1
// with page_size=2 over three matches must return two items, not one.
func TestSearchDreamsHTTPFirstPageFull(t *testing.T) {
	app, _ := newTestApp(t)
	user, token := registerForTest(t, app, "pager2@example.com")

	day := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	createDreamForTest(t, app, user.ID, "分页一", day.Add(10*time.Hour))
	createDreamForTest(t, app, user.ID, "分页二", day.Add(11*time.Hour))
	createDreamForTest(t, app, user.ID, "分页三", clockDayEndUTC(day))

	handler := app.Handler(web.Handler())
	url := "/api/v1/dreams/search?q=" + url.QueryEscape("分页") +
		"&from=2026-08-01&to=2026-08-02&page=1&page_size=2"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var envelope struct {
		Data struct {
			Items []struct {
				Title string `json:"title"`
			} `json:"items"`
			Total      int  `json:"total"`
			TotalPages int  `json:"total_pages"`
			HasNext    bool `json:"has_next"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rec.Body.String())
	}
	if len(envelope.Data.Items) != 2 {
		t.Fatalf("expected 2 items on page 1, got %d", len(envelope.Data.Items))
	}
	if envelope.Data.Total != 3 || envelope.Data.TotalPages != 2 {
		t.Fatalf("expected total=3 total_pages=2, got total=%d total_pages=%d", envelope.Data.Total, envelope.Data.TotalPages)
	}
	if !envelope.Data.HasNext {
		t.Fatalf("expected has_next=true on page 1")
	}
	if envelope.Data.Items[0].Title != "分页三" || envelope.Data.Items[1].Title != "分页二" {
		t.Fatalf("expected descending [分页三, 分页二], got %q then %q", envelope.Data.Items[0].Title, envelope.Data.Items[1].Title)
	}
}

// TestSearchDreamsHTTPEndDayInclusive asserts the end-day record participates:
// searching for the title that only exists at the last nanosecond of the end
// day returns total=1.
func TestSearchDreamsHTTPEndDayInclusive(t *testing.T) {
	app, _ := newTestApp(t)
	user, token := registerForTest(t, app, "pager3@example.com")

	day := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	createDreamForTest(t, app, user.ID, "分页三", clockDayEndUTC(day))

	handler := app.Handler(web.Handler())
	url := "/api/v1/dreams/search?q=" + url.QueryEscape("分页三") +
		"&from=2026-08-02&to=2026-08-02&page=1&page_size=10"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var envelope struct {
		Data struct {
			Total      int `json:"total"`
			TotalPages int `json:"total_pages"`
			Items      []struct {
				Title string `json:"title"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rec.Body.String())
	}
	if envelope.Data.Total != 1 {
		t.Fatalf("expected the end-of-day record to participate (total=1), got total=%d", envelope.Data.Total)
	}
	if len(envelope.Data.Items) != 1 || envelope.Data.Items[0].Title != "分页三" {
		t.Fatalf("expected 分页三, got %+v", envelope.Data.Items)
	}
}

// TestSearchDreamsHTTPEmptyResult asserts an empty result returns consistent
// metadata (total=0, total_pages=1) rather than "total=0 total_pages=0".
func TestSearchDreamsHTTPEmptyResult(t *testing.T) {
	app, _ := newTestApp(t)
	_, token := registerForTest(t, app, "pager4@example.com")

	handler := app.Handler(web.Handler())
	url := "/api/v1/dreams/search?q=" + url.QueryEscape("不存在的关键词") + "&page=1&page_size=2"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var envelope struct {
		Data struct {
			Total       int  `json:"total"`
			TotalPages  int  `json:"total_pages"`
			HasNext     bool `json:"has_next"`
			HasPrevious bool `json:"has_previous"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rec.Body.String())
	}
	if envelope.Data.Total != 0 {
		t.Fatalf("expected total=0, got total=%d", envelope.Data.Total)
	}
	if envelope.Data.TotalPages != 1 {
		t.Fatalf("expected total_pages=1 for empty result, got total_pages=%d", envelope.Data.TotalPages)
	}
	if envelope.Data.HasNext || envelope.Data.HasPrevious {
		t.Fatalf("expected no next/previous on empty result, got has_next=%v has_previous=%v", envelope.Data.HasNext, envelope.Data.HasPrevious)
	}
}

func clockDayEndUTC(value time.Time) time.Time {
	return value.Add(24*time.Hour - time.Nanosecond)
}
