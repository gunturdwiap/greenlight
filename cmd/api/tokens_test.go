package main

import (
	"net/http"
	"testing"

	"github.com/gunturdwiap/greenlight/internal/assert"
)

func TestCreateAuthenticationToken(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	user := app.createTestUser(t)

	code, _, body := ts.postJSON(t, "/v1/tokens/authentication", map[string]any{
		"email":    user.Email,
		"password": "pa55word",
	})

	assert.Equal(t, code, http.StatusCreated)
	assert.StringContains(t, body, "authentication_token")
}

func TestCreateAuthenticationTokenInvalidCredentials(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	user := app.createTestUser(t)

	code, _, _ := ts.postJSON(t, "/v1/tokens/authentication", map[string]string{
		"email":    user.Email,
		"password": "wrongpassword",
	})
	assert.Equal(t, http.StatusUnauthorized, code)

	code, _, _ = ts.postJSON(t, "/v1/tokens/authentication", map[string]string{
		"email":    "wrongemail@example.com",
		"password": "pa55word",
	})
	assert.Equal(t, http.StatusUnauthorized, code)
}

func TestCreateAuthenticationTokenValidation(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	tests := []struct {
		name         string
		payload      map[string]string
		expectedCode int
	}{
		{
			name: "invalid email format",
			payload: map[string]string{
				"email":    "not-an-email",
				"password": "pa55word",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "missing email",
			payload: map[string]string{
				"password": "pa55word",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "missing password",
			payload: map[string]string{
				"email": "test@example.com",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "blank password",
			payload: map[string]string{
				"email":    "test@example.com",
				"password": "",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, _ := ts.postJSON(t, "/v1/tokens/authentication", tt.payload)
			assert.Equal(t, tt.expectedCode, code)
		})
	}
}
