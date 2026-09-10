package application

/*
|--------------------------------------------------------------------------
| Wiring
|--------------------------------------------------------------------------
|
| Implements Wiring as part of the application package in the Gofreight
| framework. Key symbols: ConfigureStorage, UseExceptionHandler, UseVite,
| UseRedisBroadcast, CacheRoutes, CacheRoutesIfRequested.
| 
| The application package is the framework kernel: it constructs the
| Application value that owns the router, ORM, mailer, cache, queue, and
| view engine.
| 
| bootstrap/app.go in your project returns application.New() with your
| bindings; Run() serves HTTP and optional background workers.
| 
| Most cross-cutting services are configured here or via
| ConfigureIntegrations from environment variables.
| 
| Symbols defined here include: ConfigureStorage (ConfigureStorage wires
| the local filesystem disk from config.); UseExceptionHandler
| (UseExceptionHandler registers panic recovery with HTML/JSON error
| pages.); UseVite (UseVite enables Vite dev server asset tags when
| public/hot exists.); UseRedisBroadcast (UseRedisBroadcast enables
| multi-instance WebSocket broadcasting.); CacheRoutes (CacheRoutes writes
| route metadata for faster URL generation.); CacheRoutesIfRequested
| (CacheRoutesIfRequested exits after writing route cache (CLI:
| GOFREIGHT_ROUTE_CACHE=1).); NewScheduler (NewScheduler creates the
| application task scheduler.); NewNotifier (NewNotifier creates a
| notification sender using app mailer.).
| 
*/

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/lsgser/gofreight/assets"
	"github.com/lsgser/gofreight/channels"
	"github.com/lsgser/gofreight/config"
	"github.com/lsgser/gofreight/middleware"
	"github.com/lsgser/gofreight/notification"
	"github.com/lsgser/gofreight/schedule"
	"github.com/lsgser/gofreight/storage"
)

// ConfigureStorage wires the local filesystem disk from config.
func (app *Application) ConfigureStorage() error {
	if config.ResolveFilesystemDisk() != "local" {
		return nil
	}
	root := os.Getenv("STORAGE_LOCAL_ROOT")
	if root == "" {
		root = "storage/app"
	}
	disk, err := storage.NewLocalDisk(root)
	if err != nil {
		return err
	}
	app.Storage = disk
	return nil
}

// UseExceptionHandler registers panic recovery with HTML/JSON error pages.
func (app *Application) UseExceptionHandler() {
	var render middleware.ExceptionRenderer
	if app.Views != nil {
		views := app.Views
		render = func(w http.ResponseWriter, name string, data map[string]any) error {
			return views.Render(w, name, data)
		}
	}
	handler := middleware.NewExceptionHandler(render, app.Config.AppDebug)
	app.Router.Use(handler.Middleware)
}

// UseVite enables Vite dev server asset tags when public/hot exists.
func (app *Application) UseVite() *assets.Vite {
	v := assets.NewVite()
	if app.Assets != nil {
		app.Assets.Vite = v
	}
	if app.Views != nil {
		app.Views.RegisterFunc("vite", func(entry string) template.HTML {
			return template.HTML(v.ClientTags(entry))
		})
	}
	app.Router.Use(assets.ViteProxyMiddleware(v))
	return v
}

// UseRedisBroadcast enables multi-instance WebSocket broadcasting.
func (app *Application) UseRedisBroadcast(redisURL string) error {
	rb, err := channels.NewRedisBroadcaster(app.Channels, redisURL, "gofreight:broadcast")
	if err != nil {
		return err
	}
	app.Channels.UseRedisBroadcast(rb)
	return nil
}

// CacheRoutes writes route metadata for faster URL generation.
func (app *Application) CacheRoutes() error {
	path := filepath.Join("bootstrap", "cache", "routes.json")
	return app.Router.SaveCache(path)
}

// CacheRoutesIfRequested exits after writing route cache (CLI: GOFREIGHT_ROUTE_CACHE=1).
func (app *Application) CacheRoutesIfRequested() bool {
	if os.Getenv("GOFREIGHT_ROUTE_CACHE") != "1" {
		return false
	}
	if err := app.CacheRoutes(); err != nil {
		log.Fatalf("route cache failed: %v", err)
	}
	log.Println("Route cache written to bootstrap/cache/routes.json")
	os.Exit(0)
	return true
}

// NewScheduler creates the application task scheduler.
func NewScheduler() *schedule.Scheduler {
	return schedule.New()
}

// NewNotifier creates a notification sender using app mailer.
func (app *Application) NewNotifier(store notification.DatabaseStore) *notification.Sender {
	return &notification.Sender{Mailer: app.Mailer, Store: store}
}

// WireDefaults configures file/redis drivers from environment.
func (app *Application) WireDefaults() {
	switch config.ResolveSessionDriver() {
	case "file":
		if err := app.UseFileSessions(""); err != nil {
			log.Printf("warning: file sessions: %v", err)
		}
	case "redis":
		_ = app.UseRedisSessions(config.ResolveRedisURL())
	}
	if err := app.ConfigureStorage(); err != nil {
		log.Printf("warning: storage: %v", err)
	}
	_ = app.ConfigureIntegrations()
}
