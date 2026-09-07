package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lsgser/gofreight/router"
)

func TestRouteConstraints(t *testing.T) {
	r := router.New()
	r.Get("/posts/:id", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(req.PathValue("id")))
	}).WhereParam("id", "[0-9]+")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/posts/42", nil))
	if rec.Body.String() != "42" {
		t.Fatalf("expected 42, got %q", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/posts/abc", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-numeric id, got %d", rec.Code)
	}
}

func TestCatchAllWildcard(t *testing.T) {
	r := router.New()
	r.Get("/files/{path*}", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(req.PathValue("path")))
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/files/docs/guide.md", nil))
	if rec.Body.String() != "docs/guide.md" {
		t.Fatalf("expected catch-all path, got %q", rec.Body.String())
	}
}

func TestOptionalParam(t *testing.T) {
	r := router.New()
	r.Get("/posts/{id?}", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(req.PathValue("id")))
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/posts", nil))
	if rec.Body.String() != "" {
		t.Fatalf("expected empty id, got %q", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/posts/7", nil))
	if rec.Body.String() != "7" {
		t.Fatalf("expected 7, got %q", rec.Body.String())
	}
}

func TestDomainRouting(t *testing.T) {
	r := router.New()
	r.SetDefaultDomain("example.com")
	r.Group(func(api *router.Router) {
		api.Get("/status", func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte("api"))
		})
	}).Subdomain("api").Apply()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	req.Host = "api.example.com"
	r.ServeHTTP(rec, req)
	if rec.Body.String() != "api" {
		t.Fatalf("expected api subdomain route, got %q", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/status", nil)
	req.Host = "www.example.com"
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 on wrong host, got %d", rec.Code)
	}
}

func TestParameterizedDomain(t *testing.T) {
	r := router.New()
	r.Group(func(tenant *router.Router) {
		tenant.Get("/", func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte(req.PathValue("host_tenant")))
		})
	}).Domain("{tenant}.example.com").Apply()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "acme.example.com"
	r.ServeHTTP(rec, req)
	if rec.Body.String() != "acme" {
		t.Fatalf("expected tenant acme, got %q", rec.Body.String())
	}
}

func TestSignedURL(t *testing.T) {
	signer := router.NewURLSigner("test-secret-key")
	signed := signer.SignRelative("/invites/accept", time.Hour)
	if !signer.Verify(signed) {
		t.Fatal("expected valid signature")
	}
	if signer.Verify("/invites/accept") {
		t.Fatal("unsigned URL should fail verification")
	}
}

func TestSignedRouteMiddleware(t *testing.T) {
	signer := router.NewURLSigner("test-secret-key")
	r := router.New()
	r.Get("/secret", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("ok"))
	}).Signed(signer)

	unsigned := httptest.NewRecorder()
	r.ServeHTTP(unsigned, httptest.NewRequest(http.MethodGet, "/secret", nil))
	if unsigned.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", unsigned.Code)
	}

	signed := signer.SignRelative("/secret", time.Hour)
	authorized := httptest.NewRecorder()
	r.ServeHTTP(authorized, httptest.NewRequest(http.MethodGet, signed, nil))
	if authorized.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", authorized.Code)
	}
}

func TestURLWithCatchAll(t *testing.T) {
	r := router.New()
	r.Get("/files/:path*", func(w http.ResponseWriter, req *http.Request) {}, "files.show")

	url, err := r.URL("files.show", map[string]string{"path": "a/b/c"})
	if err != nil {
		t.Fatal(err)
	}
	if url != "/files/a/b/c" {
		t.Fatalf("unexpected url %q", url)
	}
}

func TestCustomBind(t *testing.T) {
	r := router.New()
	r.Get("/users/:id", func(w http.ResponseWriter, req *http.Request) {
		v, ok := router.Bound(req, "id")
		if !ok {
			t.Fatal("missing bound value")
		}
		w.Write([]byte(v.(string)))
	}).Bind("id", func(req *http.Request) (any, error) {
		return "user-" + req.PathValue("id"), nil
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/users/42", nil))
	if rec.Body.String() != "user-42" {
		t.Fatalf("got %q", rec.Body.String())
	}
}
