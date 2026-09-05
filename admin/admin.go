package admin

import (
	"context"
	"crypto/subtle"
	"embed"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"github.com/lsgser/gofreight/controller"
	"github.com/lsgser/gofreight/database"
	"github.com/lsgser/gofreight/middleware"
	"github.com/lsgser/gofreight/router"
)

//go:embed templates/*
var templateFS embed.FS

const prefix = "/admin"

// Config holds admin dashboard settings.
type Config struct {
	Prefix      string
	Title       string
	PerPage     int
	AllowSQL    bool
	Development bool
	Password    string // from ADMIN_PASSWORD; empty = no auth
}

// DefaultConfig returns sensible defaults for local development.
func DefaultConfig(development bool) Config {
	return Config{
		Prefix:      prefix,
		Title:       "Gofreight Admin",
		PerPage:     25,
		AllowSQL:    development,
		Development: development,
	}
}

// Panel is the database admin dashboard.
type Panel struct {
	cfg  Config
	tmpl *template.Template
}

// New creates an admin panel.
func New(cfg Config) (*Panel, error) {
	funcMap := template.FuncMap{
		"plus":  func(a, b int) int { return a + b },
		"minus": func(a, b int) int { return a - b },
		"seq": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i
			}
			return s
		},
	}
	tmpl, err := template.New("admin").Funcs(funcMap).ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	if cfg.Prefix == "" {
		cfg.Prefix = prefix
	}
	if cfg.PerPage <= 0 {
		cfg.PerPage = 25
	}
	return &Panel{cfg: cfg, tmpl: tmpl}, nil
}

// Mount registers admin routes on the router. Only mounts if Development is true.
func (p *Panel) Mount(r *router.Router) {
	if !p.cfg.Development {
		return
	}

	wrap := func(h func(controller.Base) error) http.HandlerFunc {
		return controller.Handler(func(base controller.Base) error {
			if !p.isPublicPath(base.Request.URL.Path) && !p.authenticated(base.Request) {
				base.Redirect(p.cfg.Prefix+"/login", http.StatusSeeOther)
				return nil
			}
			return h(base)
		})
	}

	r.Group(func(a *router.Router) {
		a.Get("/login", controller.Handler(p.LoginForm))
		a.Post("/login", controller.Handler(p.LoginSubmit))
		a.Get("/", wrap(p.Dashboard))
		a.Get("/sql", wrap(p.SQLConsole))
		a.Post("/sql", wrap(p.SQLRun))
		p.mountSchemaRoutes(a, wrap)
		a.Get("/tables/:table", wrap(p.TableIndex))
		a.Get("/tables/:table/new", wrap(p.TableNew))
		a.Post("/tables/:table", wrap(p.TableCreate))
		a.Get("/tables/:table/:id", wrap(p.TableShow))
		a.Post("/tables/:table/:id", wrap(p.TableUpdate))
		a.Post("/tables/:table/:id/delete", wrap(p.TableDelete))
	}).Prefix(p.cfg.Prefix).Apply()
}

func (p *Panel) mountSchemaRoutes(a *router.Router, wrap func(func(controller.Base) error) http.HandlerFunc) {
	a.Get("/schema/new", wrap(p.SchemaNew))
	a.Post("/schema/create", wrap(p.SchemaCreate))
	a.Get("/import", wrap(p.ImportForm))
	a.Post("/import", wrap(p.ImportRun))
	a.Get("/integrations", wrap(p.IntegrationsStatus))
	a.Get("/tables/:table/structure", wrap(p.TableStructure))
	a.Post("/tables/:table/drop", wrap(p.TableDrop))
	a.Get("/tables/:table/columns/new", wrap(p.ColumnNew))
	a.Post("/tables/:table/columns", wrap(p.ColumnAdd))
	a.Post("/tables/:table/columns/:column/drop", wrap(p.ColumnDrop))
	a.Get("/tables/:table/columns/:column/rename", wrap(p.ColumnRenameForm))
	a.Post("/tables/:table/columns/:column/rename", wrap(p.ColumnRename))
	a.Post("/tables/:table/indexes", wrap(p.IndexAdd))
	a.Get("/tables/:table/migration", wrap(p.TableMigration))
	a.Get("/export/:table", wrap(p.TableExport))
}

func (p *Panel) render(base controller.Base, name string, data any) error {
	base.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	return p.tmpl.ExecuteTemplate(base.Response, name, data)
}

func (p *Panel) baseData(base controller.Base) map[string]any {
	return map[string]any{
		"Title":    p.cfg.Title,
		"Prefix":   p.cfg.Prefix,
		"AllowSQL": p.cfg.AllowSQL,
		"Flash":    base.Query("flash"),
	}
}

// Dashboard lists all tables.
func (p *Panel) Dashboard(base controller.Base) error {
	ctx := context.Background()
	tables, err := database.ListTables(ctx)
	if err != nil {
		return err
	}

	type tableRow struct {
		Name  string
		Count int64
	}
	var rows []tableRow
	for _, t := range tables {
		count, _ := database.TableCount(ctx, t)
		rows = append(rows, tableRow{Name: t, Count: count})
	}

	data := p.baseData(base)
	data["Tables"] = rows
	return p.render(base, "dashboard.html", data)
}

// TableIndex lists records in a table.
func (p *Panel) TableIndex(base controller.Base) error {
	ctx := context.Background()
	table := base.Param("table")
	page, _ := strconv.Atoi(base.Query("page"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * p.cfg.PerPage

	schema, err := database.TableSchema(ctx, table)
	if err != nil {
		return err
	}

	rows, err := database.TableRows(ctx, table, p.cfg.PerPage, offset)
	if err != nil {
		return err
	}
	total, _ := database.TableCount(ctx, table)
	lastPage := int(total) / p.cfg.PerPage
	if int(total)%p.cfg.PerPage > 0 {
		lastPage++
	}
	if lastPage == 0 {
		lastPage = 1
	}

	data := p.baseData(base)
	data["Table"] = table
	data["Schema"] = schema
	data["Rows"] = rows
	data["Columns"] = schema.Columns
	data["Page"] = page
	data["LastPage"] = lastPage
	data["Total"] = total
	return p.render(base, "table_index.html", data)
}

// TableShow shows edit form for a record.
func (p *Panel) TableShow(base controller.Base) error {
	ctx := context.Background()
	table := base.Param("table")
	id := base.Param("id")

	schema, err := database.TableSchema(ctx, table)
	if err != nil {
		return err
	}
	row, err := database.FindRow(ctx, table, id)
	if err != nil {
		base.NotFound("Record not found")
		return nil
	}

	data := p.baseData(base)
	data["Table"] = table
	data["Schema"] = schema
	data["Row"] = row
	data["ID"] = id
	data["Edit"] = true
	return p.render(base, "table_form.html", data)
}

// TableNew shows create form.
func (p *Panel) TableNew(base controller.Base) error {
	ctx := context.Background()
	table := base.Param("table")

	schema, err := database.TableSchema(ctx, table)
	if err != nil {
		return err
	}

	data := p.baseData(base)
	data["Table"] = table
	data["Schema"] = schema
	data["Row"] = map[string]any{}
	data["Edit"] = false
	return p.render(base, "table_form.html", data)
}

// TableCreate handles record creation.
func (p *Panel) TableCreate(base controller.Base) error {
	ctx := context.Background()
	table := base.Param("table")

	if err := base.Request.ParseForm(); err != nil {
		return err
	}
	values := formToMap(base.Request)
	if err := database.InsertRow(ctx, table, values); err != nil {
		base.Unprocessable(map[string][]string{"base": {err.Error()}})
		return nil
	}
	base.Redirect(p.cfg.Prefix+"/tables/"+table+"?flash=Created", http.StatusSeeOther)
	return nil
}

// TableUpdate handles record update.
func (p *Panel) TableUpdate(base controller.Base) error {
	ctx := context.Background()
	table := base.Param("table")
	id := base.Param("id")

	if err := base.Request.ParseForm(); err != nil {
		return err
	}
	values := formToMap(base.Request)
	if err := database.UpdateRow(ctx, table, id, values); err != nil {
		base.Unprocessable(map[string][]string{"base": {err.Error()}})
		return nil
	}
	base.Redirect(p.cfg.Prefix+"/tables/"+table+"/"+id+"?flash=Saved", http.StatusSeeOther)
	return nil
}

// TableDelete handles record deletion.
func (p *Panel) TableDelete(base controller.Base) error {
	ctx := context.Background()
	table := base.Param("table")
	id := base.Param("id")

	if err := database.DeleteRow(ctx, table, id); err != nil {
		return err
	}
	base.Redirect(p.cfg.Prefix+"/tables/"+table+"?flash=Deleted", http.StatusSeeOther)
	return nil
}

// SQLConsole shows the SQL query interface.
func (p *Panel) SQLConsole(base controller.Base) error {
	if !p.cfg.AllowSQL {
		http.NotFound(base.Response, base.Request)
		return nil
	}
	data := p.baseData(base)
	return p.render(base, "sql.html", data)
}

// SQLRun executes a read-only SQL query.
func (p *Panel) SQLRun(base controller.Base) error {
	if !p.cfg.AllowSQL {
		http.NotFound(base.Response, base.Request)
		return nil
	}
	if err := base.Request.ParseForm(); err != nil {
		return err
	}
	sql := base.Request.FormValue("query")
	rows, err := database.ExecQuery(context.Background(), sql)

	data := p.baseData(base)
	data["Query"] = sql
	if err != nil {
		data["Error"] = err.Error()
	} else {
		data["Results"] = rows
		if len(rows) > 0 {
			cols := make([]string, 0, len(rows[0]))
			for k := range rows[0] {
				cols = append(cols, k)
			}
			data["ResultColumns"] = cols
		}
	}
	return p.render(base, "sql.html", data)
}

func formToMap(r *http.Request) map[string]any {
	values := make(map[string]any)
	for k, v := range r.PostForm {
		if len(v) > 0 {
			values[k] = v[0]
		}
	}
	return values
}

// MustNew creates a panel or panics.
func MustNew(cfg Config) *Panel {
	p, err := New(cfg)
	if err != nil {
		panic(err)
	}
	return p
}

func (p *Panel) adminPassword() string {
	if p.cfg.Password != "" {
		return p.cfg.Password
	}
	return os.Getenv("ADMIN_PASSWORD")
}

func (p *Panel) isPublicPath(path string) bool {
	return path == p.cfg.Prefix+"/login"
}

func (p *Panel) authenticated(r *http.Request) bool {
	if p.adminPassword() == "" {
		return true
	}
	session := middleware.SessionFromContext(r.Context())
	if session == nil {
		return false
	}
	authed, _ := session.Get("_admin_authed").(bool)
	return authed
}

// LoginForm shows admin login page.
func (p *Panel) LoginForm(base controller.Base) error {
	data := p.baseData(base)
	return p.render(base, "login.html", data)
}

// LoginSubmit handles admin login.
func (p *Panel) LoginSubmit(base controller.Base) error {
	password := p.adminPassword()
	if err := base.Request.ParseForm(); err != nil {
		return err
	}
	submitted := base.Request.FormValue("password")
	if subtle.ConstantTimeCompare([]byte(submitted), []byte(password)) != 1 {
		data := p.baseData(base)
		data["Error"] = "Invalid password"
		return p.render(base, "login.html", data)
	}
	session := middleware.SessionFromContext(base.Request.Context())
	if session != nil {
		session.Set("_admin_authed", true)
	}
	base.Redirect(p.cfg.Prefix+"/", http.StatusSeeOther)
	return nil
}
