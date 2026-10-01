package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/gunturdwiap/greenlight/internal/assert"
	"github.com/gunturdwiap/greenlight/internal/data"
)

func TestRegisterUser(t *testing.T) {
	app := newTestApplication(t)

	ts := newTestServer(t, app.routes())

	code, _, _ := ts.postJSON(t, "/v1/users", map[string]string{
		"name":     "test",
		"email":    "test@example.com",
		"password": "pa55word",
	})

	assert.Equal(t, app.mailer.recipient, "test@example.com")
	assert.Equal(t, app.mailer.templateFile, "user_welcome.tmpl")
	assert.Equal(t, http.StatusAccepted, code)
}

func TestRegisterUserDuplicateEmail(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	// first registration
	code, _, _ := ts.postJSON(t, "/v1/users", map[string]string{
		"name":     "test",
		"email":    "test@example.com",
		"password": "pa55word",
	})
	assert.Equal(t, http.StatusAccepted, code)

	// second registration
	code, _, _ = ts.postJSON(t, "/v1/users", map[string]string{
		"name":     "another one",
		"email":    "test@example.com",
		"password": "pa55word",
	})
	assert.Equal(t, http.StatusUnprocessableEntity, code)
}

func TestRegisterUserValidation(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	tests := []struct {
		name         string
		payload      map[string]string
		expectedCode int
	}{
		{
			name: "short password",
			payload: map[string]string{
				"name":     "test",
				"email":    "test2@example.com",
				"password": "short",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "missing password",
			payload: map[string]string{
				"name":  "test",
				"email": "test3@example.com",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "missing email",
			payload: map[string]string{
				"name":     "test",
				"password": "pa55word",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "missing name",
			payload: map[string]string{
				"email":    "test4@example.com",
				"password": "pa55word",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid email",
			payload: map[string]string{
				"name":     "test",
				"email":    "invalid-email",
				"password": "pa55word",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, _ := ts.postJSON(t, "/v1/users", tt.payload)
			assert.Equal(t, tt.expectedCode, code)
		})
	}
}

func TestActivateUser(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	user := app.createTestUser(t)
	activationToken := app.createTestToken(t, user.ID, time.Hour, data.ScopeActivation)

	code, _, body := ts.putJSON(t, "/v1/users/activated", map[string]any{
		"token": activationToken.Plaintext,
	})

	assert.Equal(t, code, http.StatusOK)
	assert.StringContains(t, body, `"activated": true`)
}

func TestActivateUserInvalidToken(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	code, _, _ := ts.putJSON(t, "/v1/users/activated", map[string]any{
		"token": "invalid-token",
	})

	assert.Equal(t, code, http.StatusUnprocessableEntity)
}

func TestActivateUserExpiredToken(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	user := app.createTestUser(t)
	activationToken := app.createTestToken(t, user.ID, -time.Hour, data.ScopeActivation)

	code, _, _ := ts.putJSON(t, "/v1/users/activated", map[string]any{
		"token": activationToken.Plaintext,
	})

	assert.Equal(t, code, http.StatusUnprocessableEntity)
}

func TestActivateUserMissingToken(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	code, _, _ := ts.putJSON(t, "/v1/users/activated", map[string]any{
		// no token
	})

	assert.Equal(t, code, http.StatusUnprocessableEntity)
}

func TestUpdateUserPassword(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	user := app.createTestUser(t, withActivated(true))
	passwordResetToken := app.createTestToken(t, user.ID, time.Hour, data.ScopePasswordReset)

	code, _, _ := ts.putJSON(t, "/v1/users/password", map[string]any{
		"token":    passwordResetToken.Plaintext,
		"password": "updatedPa55word",
	})

	assert.Equal(t, code, http.StatusOK)

	updatedUser, err := app.models.Users.GetByEmail(user.Email)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := updatedUser.Password.Matches("updatedPa55word")
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, ok)
}

func TestUpdateUserPasswordInvalidToken(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	code, _, _ := ts.putJSON(t, "/v1/users/password", map[string]any{
		"token":    "invalid",
		"password": "updatedPa55word",
	})

	assert.Equal(t, code, http.StatusUnprocessableEntity)
}

func TestUpdateUserPasswordExpiredToken(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	user := app.createTestUser(t, withActivated(true))
	passwordResetToken := app.createTestToken(t, user.ID, -time.Hour, data.ScopePasswordReset)

	code, _, _ := ts.putJSON(t, "/v1/users/password", map[string]any{
		"token":    passwordResetToken.Plaintext,
		"password": "updatedPa55word",
	})

	assert.Equal(t, code, http.StatusUnprocessableEntity)
}

func TestUpdateUserPasswordMissingToken(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	code, _, _ := ts.putJSON(t, "/v1/users/password", map[string]any{
		// no token
		"password": "updatedPa55word",
	})

	assert.Equal(t, code, http.StatusUnprocessableEntity)
}

func TestUpdateUserPasswordMissingPassword(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	user := app.createTestUser(t, withActivated(true))
	passwordResetToken := app.createTestToken(t, user.ID, time.Hour, data.ScopePasswordReset)

	code, _, _ := ts.putJSON(t, "/v1/users/password", map[string]any{
		"token": passwordResetToken.Plaintext,
		// no password
	})

	assert.Equal(t, code, http.StatusUnprocessableEntity)
}
