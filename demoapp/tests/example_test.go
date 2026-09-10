package tests

/*
|--------------------------------------------------------------------------
| Example
|--------------------------------------------------------------------------
|
| Test suite for Example in the tests package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
| 
| Run from the app directory with go test ./... or gofreight test for
| tests/ packages.
| 
*/

import (
	"testing"

	"demoapp/routes"
	"github.com/lsgser/gofreight/gftest"
)

func TestApp(t *testing.T) {
	gftest.Describe(t, "demoapp", func(d *gftest.DescribeContext) {
		var app *gftest.App

		d.BeforeEach(func(t *testing.T) {
			app = gftest.NewApp(t, gftest.WithDatabase("sqlite://:memory:"))
			app.Draw(routes.Register)
		})

		d.It("returns the homepage", func(t *testing.T) {
			app.Get("/").AssertOk()
		})
	})
}
