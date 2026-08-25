package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Claims struct {
	UserID    string `json:"uid"`
	Name      string `json:"name"`
	ExpiresAt int64  `json:"exp"`
}

type TokenCodec struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewTokenCodec(
	secret string,
	ttl time.Duration,
) TokenCodec {
	return TokenCodec{secret: []byte(secret), ttl: ttl, now: time.Now}
}

// withClock overrides the clock used to stamp and check expirations. It is
// intended for deterministic tests of the second-granularity boundary, where a
// real wall clock would let the boundary slip past between Issue and Parse.
func (c TokenCodec) withClock(now func() time.Time) TokenCodec {
	copy := c
	copy.now = now
	return copy
}

func (
	c TokenCodec,
) clock() time.Time {
	if c.now != nil {
		return c.now()
	}
	return time.Now()
}

func (
	c TokenCodec,
) Issue(
	userID,
	name string,
) (string, error) {
	header := b64([]byte(`{"alg":"HS256","typ":"DREAM"}`))
	claims, err := json.Marshal(
		Claims{UserID: userID, Name: name, ExpiresAt: c.clock().Add(c.ttl).Unix()})
	if err != nil {
		return "", err
	}
	payload := b64(claims)
	return header + "." + payload + "." + c.sign(header+"."+payload), nil
}

func (
	c TokenCodec,
) Parse(
	token string,
) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || !hmac.Equal([]byte(parts[2]), []byte(c.sign(parts[0]+"."+parts[1]))) {
		return Claims{}, errors.New("令牌签名无效")
	}
	content, err :=
		base64.RawURLEncoding.DecodeString(
			parts[1])
	if err != nil {
		return Claims{}, err
	}
	var claims Claims
	if err := json.Unmarshal(content, &claims); err != nil {
		return Claims{}, err
	}
	// exp is second-granularity Unix time. A token whose exp equals the current
	// Unix second has already entered its expiry second, so it must be rejected
	// immediately rather than honored for the rest of that tick.
	if claims.UserID == "" || claims.ExpiresAt <= c.clock().Unix() {
		return Claims{}, errors.New("令牌已过期")
	}
	return claims, nil
}

func (c TokenCodec) TTL() time.Duration {
	return c.ttl
}

func (
	c TokenCodec,
) sign(
	value string,
) string {
	mac := hmac.New(sha256.New, c.secret)
	mac.Write([]byte(value))
	return b64(mac.Sum(nil))
}
func b64(
	value []byte,
) string {
	return base64.RawURLEncoding.EncodeToString(value)
}
