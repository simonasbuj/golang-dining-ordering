package handlers

import (
	authDto "golang-dining-ordering/services/auth/dto"
	"golang-dining-ordering/services/management/middleware"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestGetUUUIDFromParams(t *testing.T) {
	t.Parallel()

	e := echo.New()

	validUUID := uuid.New()
	tests := []struct {
		name          string
		paramValue    string
		wantUUID      uuid.UUID
		wantErrSubstr string
	}{
		{
			name:          "valid UUID",
			paramValue:    validUUID.String(),
			wantUUID:      validUUID,
			wantErrSubstr: "",
		},
		{
			name:          "invalid UUID",
			paramValue:    "not-a-uuid",
			wantUUID:      uuid.Nil,
			wantErrSubstr: "parsing uuid from params for tableId",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.SetParamNames("tableId")
			c.SetParamValues(tt.paramValue)

			id, err := GetUUUIDFromParams(c, "tableId")

			if tt.wantErrSubstr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrSubstr)
				require.Equal(t, uuid.Nil, id)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantUUID, id)
			}
		})
	}
}

func TestGetUserFromContext(t *testing.T) { //nolint:funlen
	t.Parallel()

	e := echo.New()

	testUserID := uuid.New()
	emptyUser := &authDto.TokenClaimsDto{}

	tests := []struct {
		name          string
		setupContext  func(c echo.Context)
		failOnMissing bool
		wantUserID    uuid.UUID
		wantErr       bool
	}{
		{
			name: "user present",
			setupContext: func(c echo.Context) {
				c.Set(middleware.ContextKeyAuthUser, &authDto.TokenClaimsDto{
					UserID: testUserID,
				})
			},
			failOnMissing: true,
			wantUserID:    testUserID,
			wantErr:       false,
		},
		{
			name:          "no user, fail true",
			setupContext:  func(_ echo.Context) {},
			failOnMissing: true,
			wantErr:       true,
		},
		{
			name:          "no user, fail false",
			setupContext:  func(_ echo.Context) {},
			failOnMissing: false,
			wantUserID:    uuid.Nil,
			wantErr:       false,
		},
		{
			name: "user with Nil UUID, fail true",
			setupContext: func(c echo.Context) {
				c.Set(middleware.ContextKeyAuthUser, emptyUser)
			},
			failOnMissing: true,
			wantErr:       true,
		},
		{
			name: "user with Nil UUID, fail false",
			setupContext: func(c echo.Context) {
				c.Set(middleware.ContextKeyAuthUser, emptyUser)
			},
			failOnMissing: false,
			wantUserID:    uuid.Nil,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			tt.setupContext(c)

			user, err := GetUserFromContext(c, tt.failOnMissing)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				require.Equal(t, tt.wantUserID, user.UserID)
			}
		})
	}
}
