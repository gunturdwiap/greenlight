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
	mailer := &mockMailer{}
	app.mailer = mailer

	ts := newTestServer(t, app.routes())

	code, _, _ := ts.postJSON(t, "/v1/users", map[string]string{
		"name":     "test",
		"email":    "test@example.com",
		"password": "pa55word",
	})

	assert.Equal(t, mailer.recipient, "test@example.com")
	assert.Equal(t, mailer.templateFile, "user_welcome.tmpl")
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

	user := &data.User{
		Name:      "test",
		Email:     "test@example.com",
		Activated: false,
	}
	if err := user.Password.Set("pa55word"); err != nil {
		t.Fatal(err)
	}

	if err := app.models.Users.Insert(user); err != nil {
		t.Fatal(err)
	}

	activationToken, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		t.Fatal(err)
	}

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

	user := &data.User{
		Name:      "expired",
		Email:     "expired@example.com",
		Activated: false,
	}
	if err := user.Password.Set("pa55word"); err != nil {
		t.Fatal(err)
	}

	if err := app.models.Users.Insert(user); err != nil {
		t.Fatal(err)
	}

	activationToken, err := app.models.Tokens.New(user.ID, -time.Hour, data.ScopeActivation)
	if err != nil {
		t.Fatal(err)
	}

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

func TestActivateUserAlreadyActivated(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	user := &data.User{
		Name:      "activated",
		Email:     "activated@example.com",
		Activated: true,
	}
	if err := user.Password.Set("pa55word"); err != nil {
		t.Fatal(err)
	}

	if err := app.models.Users.Insert(user); err != nil {
		t.Fatal(err)
	}

	activationToken, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		t.Fatal(err)
	}

	code, _, body := ts.putJSON(t, "/v1/users/activated", map[string]any{
		"token": activationToken.Plaintext,
	})

	assert.Equal(t, code, http.StatusOK)
	assert.StringContains(t, body, `"activated": true`)
}
