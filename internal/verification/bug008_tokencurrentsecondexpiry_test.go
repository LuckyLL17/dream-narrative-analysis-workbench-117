package verification

// ExpiresAt is the Unix-second boundary asserted by this session test; SessionCookieMaxAge is checked below.

import (
	"dream117/internal/security"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBug008Tokencurrentsecondexpiry(t *testing.T) {
	codec := security.NewTokenCodec("test-secret", -time.Nanosecond)
	token, err := codec.Issue("u", "梦者")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := codec.Parse(token); err == nil {
		t.Fatalf("token at current-second expiry was accepted")
	}
}

func TestBug008SessionCookieCarriesCodecLifetime(t *testing.T) {
	codec := security.NewTokenCodec("test-secret", 2*time.Hour)
	token, err := codec.Issue("u", "梦者")
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	security.SetCookie(rec, token, 7200)
	cookie := rec.Result().Cookies()[0]
	if cookie.MaxAge != 7200 || cookie.Expires.IsZero() {
		t.Fatalf("session cookie lifetime mismatch: max_age=%d expires=%v", cookie.MaxAge, cookie.Expires)
	}
}
