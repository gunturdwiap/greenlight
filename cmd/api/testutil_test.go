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

	"github.com/gunturdwiap/greenlight/internal/data"
)

func newTestApplication(t *testing.T) *application {
	return &application{
		config: config{
			env: "testing",
		},
		logger: slog.New(slog.DiscardHandler),
		models: data.NewModels(newTestDB(t)),
		// mailer: ,
	}
}

func newTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", "postgres://greenlight:pa55word@localhost/greenlight_test?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		t.Fatal(err)
	}

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

	rs, err := ts.Client().Get(urlPath)
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

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(data)
	if err != nil {
		t.Fatal(err)
	}

	rs, err := ts.Client().Post(urlPath, "application/json", &buf)
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
