package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	authDto "golang-dining-ordering/services/auth/dto"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

//nolint:gochecknoglobals
var (
	testUserID = "67676767-6767-4676-8767-676767676767"
	testToken  = "my-token"
)

func TestHandleAuthError_FailFalse(t *testing.T) {
	t.Parallel()

	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	nextCalled := false

	next := func(_ echo.Context) error {
		nextCalled = true

		return nil
	}

	err := handleAuthError(
		c,
		false, // fail
		http.StatusUnauthorized,
		"ignored",
		nil,
		next,
	)

	require.NoError(t, err)
	require.True(t, nextCalled)

	user := c.Get(ContextKeyAuthUser)
	require.NotNil(t, user)
	_, ok := user.(*authDto.TokenClaimsDto)
	require.True(t, ok)
}

func TestHandleAuthError_FailTrue_WithBody(t *testing.T) {
	t.Parallel()

	e := echo.New()

	body := []byte(`{"error":"token invalid"}`)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(_ echo.Context) error {
		t.Fatal("next handler should not be called")

		return nil
	}

	err := handleAuthError(
		c,
		true, // fail
		http.StatusUnauthorized,
		"ignored",
		body,
		next,
	)

	require.Error(t, err)
	require.ErrorIs(t, err, errUnauthorized)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.JSONEq(t, `{"error":"token invalid"}`, rec.Body.String())
}

func TestHandleAuthError_FailTrue_NoBody(t *testing.T) {
	t.Parallel()

	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(_ echo.Context) error {
		t.Fatal("next handler should not be called")

		return nil
	}

	err := handleAuthError(
		c,
		true, // fail
		http.StatusUnauthorized,
		"missing or invalid token",
		nil,
		next,
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "sending response to client")

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.JSONEq(t, `{"error":"missing or invalid token"}`, rec.Body.String())
}

func TestParseAndStoreAuthResponse_Success(t *testing.T) {
	t.Parallel()

	e := echo.New()
	c := e.NewContext(
		httptest.NewRequest(http.MethodGet, "/", nil),
		httptest.NewRecorder(),
	)

	expectedUser := authDto.TokenClaimsDto{
		UserID: uuid.MustParse(testUserID),
	}

	respBody, err := json.Marshal(AuthResponse{Data: expectedUser})
	require.NoError(t, err)

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(respBody)),
	}

	err = parseAndStoreAuthResponse(c, resp)
	require.NoError(t, err)

	stored := c.Get(ContextKeyAuthUser)
	require.NotNil(t, stored)

	user, ok := stored.(*authDto.TokenClaimsDto)
	require.True(t, ok)
	require.Equal(t, expectedUser, *user)
}

func TestParseAndStoreAuthResponse_ReadBodyError(t *testing.T) {
	t.Parallel()

	e := echo.New()
	c := e.NewContext(
		httptest.NewRequest(http.MethodGet, "/", nil),
		httptest.NewRecorder(),
	)

	brokenBody := io.NopCloser(errReader{})

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       brokenBody,
	}

	err := parseAndStoreAuthResponse(c, resp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read auth response body")
}

type errReader struct{}

var errRead = errors.New("read error")

func (errReader) Read([]byte) (int, error) {
	return 0, errRead
}

func TestParseAndStoreAuthResponse_UnmarshalError(t *testing.T) {
	t.Parallel()

	e := echo.New()
	c := e.NewContext(
		httptest.NewRequest(http.MethodGet, "/", nil),
		httptest.NewRecorder(),
	)

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString("{invalid json")),
	}

	err := parseAndStoreAuthResponse(c, resp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to unmarshal auth response")
}

func TestRoleMiddleware(t *testing.T) { //nolint:funlen
	t.Parallel()

	e := echo.New()

	tests := []struct {
		name         string
		setupContext func(c echo.Context)
		allowedRoles []authDto.Role
		wantErrCode  int
		nextCalled   bool
	}{
		{
			name:         "missing claims",
			setupContext: func(_ echo.Context) {},
			allowedRoles: []authDto.Role{authDto.RoleManager},
			wantErrCode:  http.StatusBadRequest,
			nextCalled:   false,
		},
		{
			name: "wrong claims type",
			setupContext: func(c echo.Context) {
				c.Set(ContextKeyAuthUser, "not a TokenClaimsDto")
			},
			allowedRoles: []authDto.Role{authDto.RoleManager},
			wantErrCode:  http.StatusBadRequest,
			nextCalled:   false,
		},
		{
			name: "role allowed",
			setupContext: func(c echo.Context) {
				c.Set(ContextKeyAuthUser, &authDto.TokenClaimsDto{
					Role: authDto.RoleManager,
				})
			},
			allowedRoles: []authDto.Role{authDto.RoleWaiter, authDto.RoleManager},
			wantErrCode:  0, // means next() should run with no error
			nextCalled:   true,
		},
		{
			name: "role forbidden",
			setupContext: func(c echo.Context) {
				c.Set(ContextKeyAuthUser, &authDto.TokenClaimsDto{
					Role: authDto.RoleWaiter,
				})
			},
			allowedRoles: []authDto.Role{authDto.RoleManager},
			wantErrCode:  http.StatusForbidden,
			nextCalled:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			tt.setupContext(c)

			called := false
			next := func(_ echo.Context) error {
				called = true

				return nil
			}

			m := RoleMiddleware(tt.allowedRoles...)
			err := m(next)(c)

			if tt.wantErrCode != 0 {
				require.Error(t, err)

				httpErr := &echo.HTTPError{}
				ok := errors.As(err, &httpErr)
				require.True(t, ok)
				require.Equal(t, tt.wantErrCode, httpErr.Code)
				require.False(t, called)
			} else {
				require.NoError(t, err)
				require.True(t, called)
			}
		})
	}
}

func TestCallAuthService_Success(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	resp, err := callAuthService(context.Background(), ts.URL, testToken)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = resp.Body.Close()
	require.NoError(t, err)
	require.JSONEq(t, `{"status":"ok"}`, string(body))
}

func TestCallAuthService_RequestCreationError(t *testing.T) {
	t.Parallel()

	_, err := callAuthService(context.Background(), ":", "token") //nolint:bodyclose
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create auth request")
}

var errNetwork = errors.New("network error")

func TestCallAuthService_DoError(t *testing.T) {
	t.Parallel()

	oldClient := http.DefaultClient

	defer func() { http.DefaultClient = oldClient }()

	http.DefaultClient = &http.Client{
		Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return nil, errNetwork
		}),
	}

	//nolint:bodyclose
	_, err := callAuthService(context.Background(), "http://fake-website.eu", "token")
	require.Error(t, err)
	require.Contains(t, err.Error(), "calling auth service")
}

// helper type to mock Transport.
type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestAuthMiddleware(t *testing.T) { //nolint:funlen
	t.Parallel()

	e := echo.New()

	tests := []struct {
		name          string
		authHeader    string
		serverHandler http.HandlerFunc
		nextCalled    bool
	}{
		{
			name:       "missing Authorization header",
			authHeader: "",
			nextCalled: false,
		},
		{
			name:       "auth service returns non-200",
			authHeader: testToken,
			serverHandler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"invalid"}`))
			},
			nextCalled: false,
		},
		{
			name:       "parseAndStoreAuthResponse fails",
			authHeader: testToken,
			serverHandler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`invalid json`))
			},
			nextCalled: false,
		},
		{
			name:       "success path",
			authHeader: testToken,
			serverHandler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)

				respBody, _ := json.Marshal(AuthResponse{ //nolint:errchkjson
					Data: authDto.TokenClaimsDto{UserID: uuid.MustParse(testUserID)},
				})

				_, _ = w.Write(respBody)
			},
			nextCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var authURL string

			if tt.serverHandler != nil {
				ts := httptest.NewServer(tt.serverHandler)
				defer ts.Close()

				authURL = ts.URL
			} else {
				authURL = "http://example.invalid"
			}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			nextCalled := false
			next := func(_ echo.Context) error {
				nextCalled = true

				return nil
			}

			mw := AuthMiddleware(authURL, true) // failOnMissingUser = true
			err := mw(next)(c)

			if !tt.nextCalled {
				require.Error(t, err)
				require.False(t, nextCalled)
			} else {
				require.NoError(t, err)
				require.True(t, nextCalled)

				user := c.Get(ContextKeyAuthUser)
				require.NotNil(t, user)
				claims, ok := user.(*authDto.TokenClaimsDto)
				require.True(t, ok)
				require.Equal(t, uuid.MustParse(testUserID), claims.UserID)
			}
		})
	}
}

func TestAuthMiddleware_AuthServiceUnreachable(t *testing.T) {
	t.Parallel()

	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", testToken)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	nextCalled := false
	next := func(_ echo.Context) error {
		nextCalled = true

		return nil
	}

	mw := AuthMiddleware("http://invalid-unreachable-url.eu", true)

	err := mw(next)(c)

	require.Error(t, err)
	require.False(t, nextCalled)
}
