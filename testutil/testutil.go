package testutil

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofreight/gofreight/application"
	"github.com/gofreight/gofreight/config"
	"github.com/gofreight/gofreight/gftest"
	"github.com/gofreight/gofreight/router"
)

// TestApp wraps a gftest.App for backward compatibility.
type TestApp struct {
	*gftest.App
}

// SetupTestApp creates an application in test mode.
func SetupTestApp(t *testing.T, databaseURL string) *TestApp {
	t.Helper()
	opts := []gftest.AppOption{}
	if databaseURL != "" {
		opts = append(opts, gftest.WithDatabase(databaseURL))
	}
	return &TestApp{App: gftest.NewApp(t, opts...)}
}

// Draw registers routes on the test application.
func (ta *TestApp) Draw(routes application.RoutesFunc) {
	ta.App.Draw(routes)
}

// Request performs an HTTP request against the test app.
func (ta *TestApp) Request(method, path string, body io.Reader) *httptest.ResponseRecorder {
	resp := ta.App.Request(method, path, body)
	return resp.Recorder
}

// Get performs a GET request.
func (ta *TestApp) Get(path string) *httptest.ResponseRecorder {
	return ta.App.Get(path).Recorder
}

// Post performs a POST request with JSON body.
func (ta *TestApp) Post(t *testing.T, path string, payload any) *httptest.ResponseRecorder {
	return ta.App.PostJSON(path, payload).Recorder
}

// Put performs a PUT request with JSON body.
func (ta *TestApp) Put(t *testing.T, path string, payload any) *httptest.ResponseRecorder {
	return ta.App.PutJSON(path, payload).Recorder
}

// Delete performs a DELETE request.
func (ta *TestApp) Delete(path string) *httptest.ResponseRecorder {
	return ta.App.Delete(path).Recorder
}

// AssertStatus fails if the response status doesn't match.
func AssertStatus(t *testing.T, rec *httptest.ResponseRecorder, expected int) {
	t.Helper()
	gftest.Expect(rec.Code).Bind(t).ToEqual(expected)
}

// AssertJSON decodes the response body into dest.
func AssertJSON(t *testing.T, rec *httptest.ResponseRecorder, dest any) {
	t.Helper()
	(&gftest.Response{T: t, Recorder: rec}).AssertJSON(dest)
}

// AssertContains fails if the response body doesn't contain the given string.
func AssertContains(t *testing.T, rec *httptest.ResponseRecorder, substr string) {
	t.Helper()
	gftest.Expect(rec.Body.String()).Bind(t).ToContain(substr)
}

// TestConfig returns a config loaded for the test environment.
func TestConfig() *config.Config {
	return config.Load()
}

// Router returns the app router (backward compat).
func (ta *TestApp) RouterCompat() *router.Router {
	return ta.Router
}
