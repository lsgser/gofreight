package application

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofreight/gofreight/admin"
	"github.com/gofreight/gofreight/assets"
	"github.com/gofreight/gofreight/cache"
	"github.com/gofreight/gofreight/config"
	"github.com/gofreight/gofreight/controller"
	"github.com/gofreight/gofreight/database"
	"github.com/gofreight/gofreight/health"
	"github.com/gofreight/gofreight/integrations"
	"github.com/gofreight/gofreight/jobs"
	"github.com/gofreight/gofreight/mail"
	"github.com/gofreight/gofreight/middleware"
	"github.com/gofreight/gofreight/plugins"
	"github.com/gofreight/gofreight/router"
	"github.com/gofreight/gofreight/view"
	"github.com/joho/godotenv"
)

// Application is the central hub of a Gofreight app, like Rails.application.
type Application struct {
	Config   *config.Config
	Router   *router.Router
	Server   *http.Server
	Views    *view.Engine
	Assets   *assets.Pipeline
	Sessions *middleware.Sessions
	CSRF     *middleware.CSRF
	Admin    *admin.Panel
	Cache    cache.Cacher
	Mailer   mail.Mailer
	Jobs     *jobs.Queue
	Worker   *jobs.Worker
}

// New creates and configures a new Application instance.
func New() *Application {
	_ = godotenv.Load()

	cfg := config.Load()
	r := router.New()
	if cfg.IsProduction() {
		r.Use(middleware.StructuredLogger, middleware.Recovery, middleware.SecurityHeaders)
	} else {
		r.Use(middleware.Logger, middleware.Recovery)
	}

	sessions := middleware.NewSessions(cfg.SecretKey)
	csrf := middleware.NewCSRF()
	r.Use(sessions.Middleware)

	views := view.New("app/views")
	assetPipeline := assets.New("public")
	jobQueue := jobs.New()

	return &Application{
		Config:   cfg,
		Router:   r,
		Views:    views,
		Sessions: sessions,
		CSRF:     csrf,
		Assets:   assetPipeline,
		Cache:    cache.New(),
		Mailer:   mail.NewLogMailer(),
		Jobs:     jobQueue,
		Worker:   jobs.NewWorker(jobQueue),
	}
}

// ConnectDatabase establishes the database connection using config.
func (app *Application) ConnectDatabase() error {
	_, err := database.Connect(app.Config.DatabaseURL)
	return err
}

// ConfigureIntegrations initializes third-party services from environment variables.
func (app *Application) ConfigureIntegrations() error {
	svc, err := integrations.BuildServices(integrations.OsEnv{})
	if err != nil {
		return err
	}
	if svc.Cache != nil {
		app.Cache = svc.Cache
	}
	if svc.Mail != nil {
		app.Mailer = svc.Mail
	}
	return nil
}

// MountHealth registers GET /health.
func (app *Application) MountHealth() {
	checker := health.New()
	app.Router.Get("/health", checker.Handler())
}

// LoadViews parses all view templates.
func (app *Application) LoadViews() error {
	controller.SetViews(app.Views)
	return app.Views.Load()
}

// PrecompileAssets builds the asset digest map for production.
func (app *Application) PrecompileAssets() error {
	return app.Assets.Precompile()
}

// MountAssets registers the asset pipeline on the router.
func (app *Application) MountAssets() {
	app.Router.Mount(app.Assets.Prefix, app.Assets.Handler())
}

// MountAdmin registers the database admin dashboard (development/test only).
func (app *Application) MountAdmin() {
	if app.Config.IsProduction() {
		return
	}
	cfg := admin.DefaultConfig(!app.Config.IsProduction())
	panel, err := admin.New(cfg)
	if err != nil {
		log.Printf("warning: admin panel failed to load: %v", err)
		return
	}
	app.Admin = panel
	panel.Mount(app.Router)
	if app.Config.IsDevelopment() {
		log.Printf("Database admin available at http://%s:%d/admin", app.Config.Host, app.Config.Port)
	}
}

// UseCSRF enables CSRF protection on mutating requests.
func (app *Application) UseCSRF() {
	app.Router.Use(app.CSRF.Middleware)
}

// UseCORS enables CORS with optional allowed origins.
func (app *Application) UseCORS(origins ...string) {
	app.Router.Use(middleware.CORS(origins...))
}

// UseRateLimit adds rate limiting middleware.
func (app *Application) UseRateLimit(limit int, window time.Duration) {
	rl := middleware.NewRateLimiter(limit, window)
	app.Router.Use(rl.Middleware)
}

// StartJobs starts the background job worker.
func (app *Application) StartJobs(concurrency int) {
	app.Worker.Start(concurrency)
}

// Routes is a callback where apps define their routes (like config/routes.rb).
type RoutesFunc func(*router.Router)

// Draw registers routes using the provided callback.
func (app *Application) Draw(routes RoutesFunc) {
	routes(app.Router)
}

// Run starts the HTTP server with graceful shutdown.
func (app *Application) Run() {
	_ = plugins.Run("boot", app)

	if err := app.LoadViews(); err != nil {
		log.Printf("warning: views not loaded: %v", err)
	}

	if app.Config.IsProduction() {
		if err := app.PrecompileAssets(); err != nil {
			log.Printf("warning: asset precompile failed: %v", err)
		}
	}

	app.MountHealth()
	app.MountAssets()

	if app.Config.IsDevelopment() {
		app.MountAdmin()
	}

	if err := app.ConfigureIntegrations(); err != nil {
		log.Printf("warning: integrations: %v", err)
	}

	_ = plugins.Run("before_run", app)

	addr := fmt.Sprintf("%s:%d", app.Config.Host, app.Config.Port)

	app.Server = &http.Server{
		Addr:         addr,
		Handler:      app.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if app.Config.IsDevelopment() {
		log.Println("Routes:")
		for _, route := range app.Router.Routes() {
			log.Println(" ", route)
		}
	}

	go func() {
		log.Printf("Gofreight server starting on http://%s [%s]", addr, app.Config.Environment)
		if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if app.Worker != nil {
		app.Worker.Stop()
	}

	if err := app.Server.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("Server stopped")
}
