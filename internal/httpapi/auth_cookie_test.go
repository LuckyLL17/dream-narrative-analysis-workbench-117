package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"dream117/internal/security"
	"dream117/internal/service"
	"dream117/internal/store"
)

// This test guards the two session-lifecycle bugs reported together:
//
//  1. Cookie lifetime on login was hardcoded to 86400s (24h) instead of the
//     TokenCodec TTL, so for a 2h TTL the browser kept sending tokens the
//     server had already rejected.
//  2. (covered by internal/security) boundary tokens whose exp == current
//     Unix second were accepted.
//
// Here we assert that BOTH login and register emit a Set-Cookie whose Max-Age
// and Expires line up with the codec TTL, so the browser's notion of session
// end matches the server's.

const cookieName = "dream_session"

func newAuthApp(t *testing.T, ttl time.Duration) (*App, *service.AuthService) {
	t.Helper()
	data, err := store.Open(t.TempDir() + "/dreams.json")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	codec := security.NewTokenCodec("test-secret", ttl)
	auth := service.NewAuthService(data, codec)
	return &App{auth: auth}, auth
}

func parseSetCookie(t *testing.T, recorder *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected exactly one Set-Cookie, got %d", len(cookies))
	}
	return cookies[0]
}

func doJSON(t *testing.T, method, path string, body string) (*httptest.ResponseRecorder, *http.Request) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	return rec, req
}

func TestRegisterCookieLifetimeMatchesTokenTTL(t *testing.T) {
	t.Parallel()
	ttl := 2 * time.Hour
	app, _ := newAuthApp(t, ttl)

	rec, req := doJSON(t, http.MethodPost, "/api/v1/auth/register", `{"email":"alice@example.com","name":"alice","password":"hunter222"}`)
	app.register(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", rec.Code, rec.Body.String())
	}

	cookie := parseSetCookie(t, rec)
	if cookie.MaxAge != int(ttl.Seconds()) {
		t.Errorf("register cookie MaxAge = %d, want %d (ttl seconds)", cookie.MaxAge, int(ttl.Seconds()))
	}
	wantExpires := time.Now().UTC().Add(ttl)
	if delta := cookie.Expires.Sub(wantExpires); delta > 2*time.Second || delta < -2*time.Second {
		t.Errorf("register cookie Expires = %v, want ~%v (now+ttl, within 2s)", cookie.Expires, wantExpires)
	}
	if cookie.HttpOnly != true {
		t.Error("register cookie should be HttpOnly")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Error("register cookie should be SameSite=Lax")
	}
}

func TestLoginCookieLifetimeMatchesTokenTTL(t *testing.T) {
	t.Parallel()
	ttl := 2 * time.Hour
	app, _ := newAuthApp(t, ttl)

	// Seed a user via register so login has a known account.
	seedRec, seedReq := doJSON(t, http.MethodPost, "/api/v1/auth/register", `{"email":"alice@example.com","name":"alice","password":"hunter222"}`)
	app.register(seedRec, seedReq)
	if seedRec.Code != http.StatusCreated {
		t.Fatalf("seed register failed: %d %s", seedRec.Code, seedRec.Body.String())
	}

	rec, req := doJSON(t, http.MethodPost, "/api/v1/auth/login", `{"email":"alice@example.com","password":"hunter222"}`)
	app.login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	cookie := parseSetCookie(t, rec)
	if cookie.MaxAge != int(ttl.Seconds()) {
		t.Errorf("login cookie MaxAge = %d, want %d (ttl seconds, was previously hardcoded 86400)", cookie.MaxAge, int(ttl.Seconds()))
	}
	wantExpires := time.Now().UTC().Add(ttl)
	if delta := cookie.Expires.Sub(wantExpires); delta > 2*time.Second || delta < -2*time.Second {
		t.Errorf("login cookie Expires = %v, want ~%v (now+ttl, within 2s)", cookie.Expires, wantExpires)
	}
}

func TestLoginAndRegisterCookieLifetimeEqual(t *testing.T) {
	t.Parallel()
	ttl := 90 * time.Minute
	app, _ := newAuthApp(t, ttl)

	regRec, regReq := doJSON(t, http.MethodPost, "/api/v1/auth/register", `{"email":"alice@example.com","name":"alice","password":"hunter222"}`)
	app.register(regRec, regReq)

	loginRec, loginReq := doJSON(t, http.MethodPost, "/api/v1/auth/login", `{"email":"alice@example.com","password":"hunter222"}`)
	app.login(loginRec, loginReq)

	regCookie := parseSetCookie(t, regRec)
	loginCookie := parseSetCookie(t, loginRec)
	if regCookie.MaxAge != loginCookie.MaxAge {
		t.Errorf("register MaxAge %d != login MaxAge %d; both must derive from the same TTL", regCookie.MaxAge, loginCookie.MaxAge)
	}
	if want := int(ttl.Seconds()); regCookie.MaxAge != want {
		t.Errorf("cookie MaxAge = %d, want %d", regCookie.MaxAge, want)
	}
}

func TestLogoutClearsCookie(t *testing.T) {
	t.Parallel()
	app, _ := newAuthApp(t, time.Hour)
	rec := httptest.NewRecorder()
	app.logout(rec, httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil))
	cookie := parseSetCookie(t, rec)
	if cookie.MaxAge != -1 {
		t.Errorf("logout cookie MaxAge = %d, want -1 (immediate expiry)", cookie.MaxAge)
	}
	if cookie.Value != "" {
		t.Errorf("logout cookie should clear value, got %q", cookie.Value)
	}
}

func TestSessionCookieMaxAgeReflectsTTL(t *testing.T) {
	t.Parallel()
	for _, ttl := range []time.Duration{2 * time.Hour, 30 * time.Minute, 24 * time.Hour} {
		app, auth := newAuthApp(t, ttl)
		_ = app
		if got := auth.SessionCookieMaxAge(); got != int(ttl.Seconds()) {
			t.Errorf("SessionCookieMaxAge() = %d, want %d for ttl %v", got, int(ttl.Seconds()), ttl)
		}
	}
}

func TestIssuedTokenCarriesTTLOnExpiry(t *testing.T) {
	t.Parallel()
	// End-to-end-ish: the token written into the login cookie must carry an exp
	// that matches now+ttl, so the server and the cookie agree on session end.
	ttl := 2 * time.Hour
	app, _ := newAuthApp(t, ttl)
	regRec, regReq := doJSON(t, http.MethodPost, "/api/v1/auth/register", `{"email":"alice@example.com","name":"alice","password":"hunter222"}`)
	app.register(regRec, regReq)
	if regRec.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", regRec.Code, regRec.Body.String())
	}
	var resp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(regRec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode register body: %v", err)
	}
	if resp.Data.Token == "" {
		t.Fatal("register did not return a token")
	}
	// Token is header.payload.sig; decode payload to read exp.
	parts := strings.Split(resp.Data.Token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 token parts, got %d", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	var claims security.Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatalf("unmarshal claims: %v", err)
	}
	if delta := time.Until(time.Unix(claims.ExpiresAt, 0)); delta < ttl-time.Second || delta > ttl+time.Second {
		t.Errorf("token exp delta = %v, want ~%v (within 1s)", delta, ttl)
	}
}
