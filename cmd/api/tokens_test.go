package main

import (
	"net/http"
	"testing"

	"github.com/gunturdwiap/greenlight/internal/assert"
)

func TestCreateAuthenticationToken(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	user := app.createTestUser(t, false)

	code, _, body := ts.postJSON(t, "/v1/tokens/authentication", map[string]any{
		"email":    user.Email,
		"password": "pa55word",
	}, nil)

	assert.Equal(t, code, http.StatusCreated)
	assert.StringContains(t, body, "authentication_token")
}

func TestCreateAuthenticationTokenInvalidCredentials(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	user := app.createTestUser(t, false)

	code, _, _ := ts.postJSON(t, "/v1/tokens/authentication", map[string]string{
		"email":    user.Email,
		"password": "wrongpassword",
	}, nil)
	assert.Equal(t, http.StatusUnauthorized, code)

	code, _, _ = ts.postJSON(t, "/v1/tokens/authentication", map[string]string{
		"email":    "wrongemail@example.com",
		"password": "pa55word",
	}, nil)
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
			name: "invalid email",
			payload: map[string]string{
				"email":    "invalid",
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
			code, _, _ := ts.postJSON(t, "/v1/tokens/authentication", tt.payload, nil)
			assert.Equal(t, tt.expectedCode, code)
		})
	}
}

func TestCreatePasswordResetToken(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	user := app.createTestUser(t, true)

	code, _, _ := ts.postJSON(t, "/v1/tokens/password-reset", map[string]string{
		"email": user.Email,
	}, nil)

	assert.Equal(t, code, http.StatusAccepted)
	assert.Equal(t, app.mailer.recipient, user.Email)
	assert.Equal(t, app.mailer.templateFile, "token_password_reset.tmpl")
}

func TestCreatePasswordResetTokenInactiveUser(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	user := app.createTestUser(t, false)

	code, _, _ := ts.postJSON(t, "/v1/tokens/password-reset", map[string]string{
		"email": user.Email,
	}, nil)

	assert.Equal(t, code, http.StatusUnprocessableEntity)
}

func TestCreatePasswordResetTokenNonExistentEmail(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	code, _, _ := ts.postJSON(t, "/v1/tokens/password-reset", map[string]string{
		"email": "nonexistent@example.com",
	}, nil)

	assert.Equal(t, http.StatusUnprocessableEntity, code)
}

func TestCreatePasswordResetTokenValidation(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	tests := []struct {
		name         string
		payload      map[string]string
		expectedCode int
	}{
		{
			name: "invalid email",
			payload: map[string]string{
				"email": "invalid",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:    "missing email",
			payload: map[string]string{
				// no email
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, _ := ts.postJSON(t, "/v1/tokens/password-reset", tt.payload, nil)
			assert.Equal(t, tt.expectedCode, code)
		})
	}
}
