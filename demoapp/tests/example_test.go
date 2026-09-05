package tests

/*
|--------------------------------------------------------------------------
| Example Feature Test
|--------------------------------------------------------------------------
|
| Tests live in tests/ and use gftest for HTTP assertions and database
| setup. Run the suite with:
|
|   gofreight test
|
| Generate a test:
|   gofreight make:test Posts
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
