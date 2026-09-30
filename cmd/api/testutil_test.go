package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
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

func newTestApplication(t *testing.T) *application {
	mailer := &mockMailer{}
	return &application{
		config: config{
			env: "testing",
		},
		logger: slog.New(slog.DiscardHandler),
		models: data.NewModels(newTestDB(t)),
		mailer: mailer,
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
