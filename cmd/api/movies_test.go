package main

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gunturdwiap/greenlight/internal/assert"
	"github.com/gunturdwiap/greenlight/internal/data"
)

// router.HandlerFunc(http.MethodGet, "/v1/movies", app.requirePermission("movies:read", app.listMovieHandler))
// router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermission("movies:write", app.createMovieHandler))
// router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requirePermission("movies:read", app.showMovieHandler))
// router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermission("movies:write", app.updateMovieHandler))
// router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermission("movies:write", app.deleteMovieHandler))

func TestListMovie(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	user := app.createTestUser(t, true)
	app.createTestPermissions(t, user.ID, "movies:read")
	token := app.createTestToken(t, user.ID, time.Hour, data.ScopeAuthentication)

	header := http.Header{}
	header.Add("Authorization", "Bearer "+token.Plaintext)

	code, _, _ := ts.get(t, "/v1/movies", header)

	assert.Equal(t, code, http.StatusOK)
}

func TestCreateMovie(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	user := app.createTestUser(t, true)
	app.createTestPermissions(t, user.ID, "movies:write")
	token := app.createTestToken(t, user.ID, time.Hour, data.ScopeAuthentication)

	header := http.Header{}
	header.Add("Authorization", "Bearer "+token.Plaintext)

	code, _, _ := ts.postJSON(t, "/v1/movies", map[string]any{
		"title":   "Pulp Fiction",
		"year":    1994,
		"runtime": "154 mins",
		"genres":  []string{"crime, drama"},
	}, header)

	assert.Equal(t, code, http.StatusCreated)
}

func TestGetMovie(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	user := app.createTestUser(t, true)
	app.createTestPermissions(t, user.ID, "movies:read")
	token := app.createTestToken(t, user.ID, time.Hour, data.ScopeAuthentication)

	movie := app.createTestMovie(t)

	header := http.Header{}
	header.Add("Authorization", "Bearer "+token.Plaintext)

	code, _, _ := ts.get(t, fmt.Sprintf("/v1/movies/%d", movie.ID), header)

	assert.Equal(t, code, http.StatusOK)
}

func TestUpdateMovie(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	user := app.createTestUser(t, true)
	app.createTestPermissions(t, user.ID, "movies:write")
	token := app.createTestToken(t, user.ID, time.Hour, data.ScopeAuthentication)

	movie := app.createTestMovie(t)

	header := http.Header{}
	header.Add("Authorization", "Bearer "+token.Plaintext)

	code, _, _ := ts.patchJSON(t, fmt.Sprintf("/v1/movies/%d", movie.ID), map[string]any{
		"title": "Pulp Fiction",
	}, header)

	assert.Equal(t, code, http.StatusOK)

	updatedMovie, err := app.models.Movies.Get(movie.ID)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, updatedMovie.Title, "Pulp Fiction")
}

func TestDeleteMovie(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())

	user := app.createTestUser(t, true)
	app.createTestPermissions(t, user.ID, "movies:write")
	token := app.createTestToken(t, user.ID, time.Hour, data.ScopeAuthentication)

	movie := app.createTestMovie(t)

	header := http.Header{}
	header.Add("Authorization", "Bearer "+token.Plaintext)

	code, _, _ := ts.delete(t, fmt.Sprintf("/v1/movies/%d", movie.ID), header)

	assert.Equal(t, code, http.StatusOK)

	_, err := app.models.Movies.Get(movie.ID)
	assert.True(t, errors.Is(err, data.ErrRecordNotFound))
}
