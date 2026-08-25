package security

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// fixedClock returns a clock pinned to an exact instant so the second-granularity
// expiry boundary can be tested without the wall clock slipping a tick between
// Issue and Parse. The whole point of the bug is the behavior *at* the boundary
// second, so a real clock would make these tests flaky.
func fixedClock(now time.Time) func() time.Time {
	return func() time.Time { return now }
}

// boundary is a Unix instant chosen so its second is stable during the test.
var boundary = time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC)

// issueWithExpiry mints a token stamped with an explicit exp instead of now+ttl,
// so tests can target the exact expiry second (current, past, future).
func (c TokenCodec) issueWithExpiry(userID, name string, exp int64) (string, error) {
	header := b64([]byte(`{"alg":"HS256","typ":"DREAM"}`))
	claims, err := json.Marshal(Claims{UserID: userID, Name: name, ExpiresAt: exp})
	if err != nil {
		return "", err
	}
	payload := b64(claims)
	return header + "." + payload + "." + c.sign(header+"."+payload), nil
}

func TestParseRejectsTokenExpiredExactlyAtCurrentSecond(t *testing.T) {
	t.Parallel()
	codec := NewTokenCodec("test-secret", time.Hour).withClock(fixedClock(boundary))

	// Token stamped with exp == current Unix second: the boundary the bug
	// reports. Under the old code this used `<` and the token was honored for
	// the rest of the tick; it must now be rejected immediately.
	token, err := codec.issueWithExpiry("user-1", "alice", boundary.Unix())
	if err != nil {
		t.Fatalf("issue boundary token: %v", err)
	}
	if _, err := codec.Parse(token); err == nil {
		t.Fatal("expected token with exp == current second to be rejected, but Parse accepted it")
	}
}

func TestParseAcceptsTokenExpiringNextSecond(t *testing.T) {
	t.Parallel()
	codec := NewTokenCodec("test-secret", time.Hour).withClock(fixedClock(boundary))

	// exp one second into the future is still valid at the current second.
	token, err := codec.issueWithExpiry("user-1", "alice", boundary.Unix()+1)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	claims, err := codec.Parse(token)
	if err != nil {
		t.Fatalf("expected token expiring next second to be valid, got %v", err)
	}
	if claims.UserID != "user-1" {
		t.Fatalf("unexpected claims %+v", claims)
	}
}

func TestParseRejectsTokenOneSecondAfterExpiry(t *testing.T) {
	t.Parallel()
	codec := NewTokenCodec("test-secret", time.Hour).withClock(fixedClock(boundary))

	token, err := codec.issueWithExpiry("user-1", "alice", boundary.Unix()-1)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	if _, err := codec.Parse(token); err == nil {
		t.Fatal("expected token one second past exp to be rejected")
	}
}

func TestIssueStampsExpiryAsNowPlusTTL(t *testing.T) {
	t.Parallel()
	ttl := 2 * time.Hour
	codec := NewTokenCodec("test-secret", ttl).withClock(fixedClock(boundary))

	token, err := codec.Issue("user-1", "alice")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	claims, err := codec.Parse(token)
	if err != nil {
		t.Fatalf("parse freshly issued token: %v", err)
	}
	if want := boundary.Add(ttl).Unix(); claims.ExpiresAt != want {
		t.Fatalf("ExpiresAt = %d, want %d (now+ttl)", claims.ExpiresAt, want)
	}
}

func TestParseRejectsTamperedSignature(t *testing.T) {
	t.Parallel()
	codec := NewTokenCodec("test-secret", time.Hour).withClock(fixedClock(boundary))

	header := b64([]byte(`{"alg":"HS256","typ":"DREAM"}`))
	payload := b64(jsonMustMarshal(Claims{UserID: "user-1", Name: "alice", ExpiresAt: boundary.Unix() + 60}))
	forged := header + "." + payload + "." + codec.sign(header+"."+payload+"tampered")
	if _, err := codec.Parse(forged); err == nil {
		t.Fatal("expected signature mismatch to be rejected")
	}
}

func TestParseRejectsEmptyUserID(t *testing.T) {
	t.Parallel()
	codec := NewTokenCodec("test-secret", time.Hour).withClock(fixedClock(boundary))

	token, err := codec.issueWithExpiry("", "alice", boundary.Unix()+60)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	if _, err := codec.Parse(token); err == nil {
		t.Fatal("expected token with empty uid to be rejected")
	}
}

func TestTTLReflectsConfiguredValue(t *testing.T) {
	t.Parallel()
	for _, ttl := range []time.Duration{2 * time.Hour, 30 * time.Minute, 24 * time.Hour} {
		if got := NewTokenCodec("s", ttl).TTL(); got != ttl {
			t.Errorf("TTL() = %v, want %v", got, ttl)
		}
	}
}

// jsonMustMarshal is a tiny test helper; tests construct well-formed claims.
func jsonMustMarshal(c Claims) []byte {
	out, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return out
}

func TestIssueProducesThreePartToken(t *testing.T) {
	t.Parallel()
	codec := NewTokenCodec("test-secret", time.Hour).withClock(fixedClock(boundary))
	token, err := codec.Issue("user-1", "alice")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	if got := strings.Count(token, "."); got != 2 {
		t.Fatalf("expected two dots in token, got %d", got)
	}
}
