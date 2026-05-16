package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Session struct {
	RefreshToken   string    `json:"refresh_token"`
	UserID         string    `json:"user_id"`
	Email          string    `json:"email,omitempty"`
	OrganizationID string    `json:"organization_id,omitempty"`
	Issuer         string    `json:"issuer"`
	ObtainedAt     time.Time `json:"obtained_at"`

	mu          sync.Mutex
	accessToken string
	accessExp   time.Time
}

type RefreshOptions struct {
	ClientID string
	Issuer   string
}

// AccessToken returns a valid access token, refreshing via WorkOS when the
// cached one is missing or within 60s of expiry. The session is mutated in
// place — callers should persist it after use because WorkOS rotates the
// refresh token on every exchange.
func AccessToken(ctx context.Context, sess *Session, opts RefreshOptions) (string, error) {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	if sess.accessToken != "" && time.Until(sess.accessExp) > 60*time.Second {
		return sess.accessToken, nil
	}
	if sess.RefreshToken == "" {
		return "", errors.New("no refresh token; run `geoagenda login`")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", opts.ClientID)
	form.Set("refresh_token", sess.RefreshToken)

	tr, err := postForm(ctx, strings.TrimRight(opts.Issuer, "/")+"/authenticate", form)
	if err != nil {
		return "", fmt.Errorf("refresh: %w", err)
	}
	sess.accessToken = tr.AccessToken
	sess.accessExp, _ = parseExpiry(tr.AccessToken)
	if tr.RefreshToken != "" {
		sess.RefreshToken = tr.RefreshToken
	}
	return sess.accessToken, nil
}

func parseExpiry(jwt string) (time.Time, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return time.Time{}, errors.New("not a jwt")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		payload, err = base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return time.Time{}, err
		}
	}
	var c struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &c); err != nil {
		return time.Time{}, err
	}
	if c.Exp == 0 {
		return time.Time{}, errors.New("no exp claim")
	}
	return time.Unix(c.Exp, 0), nil
}
