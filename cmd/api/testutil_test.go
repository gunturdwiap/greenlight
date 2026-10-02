package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/gunturdwiap/greenlight/internal/data"
	_ "github.com/lib/pq"
)

type mockMailer struct {
	recipient    string
	templateFile string
	data         any
}

func (m *mockMailer) Send(recipient string, templateFile string, data any) error {
	m.recipient = recipient
	m.templateFile = templateFile
	m.data = data
	return nil
}

type testApplication struct {
	*application
	mailer *mockMailer
}

func newTestApplication(t *testing.T) *testApplication {
	mailer := &mockMailer{}
	db := newTestDB(t)
	app := &application{
		config: config{
			env: "testing",
		},
		logger: slog.New(slog.DiscardHandler),
		models: data.NewModels(db),
		mailer: mailer,
	}

	return &testApplication{
		application: app,
		mailer:      mailer,
	}
}

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("postgres", "postgres://greenlight_test:pa55word@localhost:5433/greenlight_test?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		t.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations",
		"postgres", driver,
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		defer db.Close()

		if err = m.Down(); err != nil {
			t.Fatal(err)
		}
	})

	return db
}

type userOption func(*data.User)

func withActivated(activated bool) userOption {
	return func(u *data.User) {
		u.Activated = activated
	}
}

func (app *testApplication) createTestUser(t *testing.T, opts ...userOption) *data.User {
	t.Helper()

	user := &data.User{
		Name: "test",
		// TODO
		Email:     fmt.Sprintf("user-%d@example.com", time.Now().UnixNano()),
		Activated: false,
	}
	if err := user.Password.Set("pa55word"); err != nil {
		t.Fatal(err)
	}

	// TODO
	for _, opt := range opts {
		opt(user)
	}

	if err := app.models.Users.Insert(user); err != nil {
		t.Fatal(err)
	}

	return user
}

func (app *testApplication) createTestToken(t *testing.T, userID int64, ttl time.Duration, scope string) *data.Token {
	t.Helper()

	token, err := app.models.Tokens.New(userID, ttl, scope)
	if err != nil {
		t.Fatal(err)
	}

	return token
}

type testServer struct {
	*httptest.Server
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	t.Helper()

	s := httptest.NewServer(h)
	return &testServer{s}
}

func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {
	t.Helper()

	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	return rs.StatusCode, rs.Header, string(body)
}

func (ts *testServer) sendJSON(t *testing.T, method, urlPath string, data any) (int, http.Header, string) {
	t.Helper()

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(data)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(method, ts.URL+urlPath, &buf)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rs, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()

	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	return rs.StatusCode, rs.Header, string(body)
}

func (ts *testServer) postJSON(t *testing.T, urlPath string, data any) (int, http.Header, string) {
	t.Helper()
	return ts.sendJSON(t, http.MethodPost, urlPath, data)
}

func (ts *testServer) putJSON(t *testing.T, urlPath string, data any) (int, http.Header, string) {
	t.Helper()
	return ts.sendJSON(t, http.MethodPut, urlPath, data)
}
