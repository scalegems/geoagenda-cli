package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/browser"
)

type LoginOptions struct {
	ClientID string
	Issuer   string // e.g. https://api.workos.com/user_management
	Port     int    // local callback port
}

const callbackPath = "/callback"

func Login(ctx context.Context, opts LoginOptions) (*Session, error) {
	verifier, challenge, err := generatePKCE()
	if err != nil {
		return nil, fmt.Errorf("pkce: %w", err)
	}
	state, err := randomState()
	if err != nil {
		return nil, fmt.Errorf("state: %w", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", opts.Port))
	if err != nil {
		return nil, fmt.Errorf("listen on 127.0.0.1:%d: %w", opts.Port, err)
	}
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d%s", opts.Port, callbackPath)
	authURL := authorizeURL(opts, redirectURI, state, challenge)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if errMsg := q.Get("error"); errMsg != "" {
			http.Error(w, errMsg, http.StatusBadRequest)
			errCh <- fmt.Errorf("authorization error: %s — %s", errMsg, q.Get("error_description"))
			return
		}
		if q.Get("state") != state {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			errCh <- errors.New("state mismatch on callback")
			return
		}
		code := q.Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			errCh <- errors.New("callback missing authorization code")
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(callbackHTML))
		codeCh <- code
	})

	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(listener) }()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	fmt.Println("Opening browser to authenticate...")
	fmt.Println("If it doesn't open automatically, visit:")
	fmt.Println("  " + authURL)
	_ = browser.OpenURL(authURL)

	var code string
	select {
	case code = <-codeCh:
	case err = <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Minute):
		return nil, errors.New("timed out waiting for authorization callback")
	}

	return exchangeCode(ctx, opts, code, verifier)
}

func authorizeURL(opts LoginOptions, redirectURI, state, challenge string) string {
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", opts.ClientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("provider", "authkit")
	q.Set("state", state)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	return strings.TrimRight(opts.Issuer, "/") + "/authorize?" + q.Encode()
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
	OrganizationID string `json:"organization_id"`
}

func exchangeCode(ctx context.Context, opts LoginOptions, code, verifier string) (*Session, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", opts.ClientID)
	form.Set("code", code)
	form.Set("code_verifier", verifier)

	tr, err := postForm(ctx, strings.TrimRight(opts.Issuer, "/")+"/authenticate", form)
	if err != nil {
		return nil, err
	}
	exp, _ := parseExpiry(tr.AccessToken)
	return &Session{
		RefreshToken:   tr.RefreshToken,
		UserID:         tr.User.ID,
		Email:          tr.User.Email,
		OrganizationID: tr.OrganizationID,
		Issuer:         opts.Issuer,
		ObtainedAt:     time.Now().UTC(),
		accessToken:    tr.AccessToken,
		accessExp:      exp,
	}, nil
}

func postForm(ctx context.Context, endpoint string, form url.Values) (*tokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("token endpoint %d: %s", resp.StatusCode, body)
	}
	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, fmt.Errorf("parse token response: %w (body=%s)", err, body)
	}
	if tr.AccessToken == "" {
		return nil, fmt.Errorf("empty access token in response: %s", body)
	}
	return &tr, nil
}

const callbackHTML = `<!doctype html>
<html><head><meta charset="utf-8"><title>geoagenda — signed in</title>
<style>body{font-family:system-ui;display:flex;align-items:center;justify-content:center;height:100vh;margin:0;background:#0f172a;color:#e2e8f0}
.card{padding:32px 40px;border-radius:16px;background:#1e293b;text-align:center;box-shadow:0 10px 40px rgba(0,0,0,.4)}
h1{margin:0 0 8px;font-size:22px}p{margin:0;color:#94a3b8}</style></head>
<body><div class="card"><h1>You're signed in</h1><p>You can close this tab and return to the terminal.</p></div>
<script>setTimeout(()=>window.close(),800)</script></body></html>`
