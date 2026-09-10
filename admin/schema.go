package admin

/*
|--------------------------------------------------------------------------
| Schema
|--------------------------------------------------------------------------
|
| Implements Schema as part of the admin package in the Gofreight
| framework. Key symbols: SchemaNew, SchemaCreate, TableStructure,
| TableDrop, ColumnNew, ColumnAdd.
| 
| The admin package exposes a development-only database browser and schema
| tools.
| 
| It introspects tables and columns so you can inspect SQLite, PostgreSQL,
| or MySQL data without leaving the browser.
| 
| Mount it only in non-production environments; see docs and application
| wiring for route registration.
| 
| Symbols defined here include: SchemaNew (SchemaNew shows the create
| table form.); SchemaCreate (SchemaCreate handles table creation.);
| TableStructure (TableStructure shows column definitions for a table.);
| TableDrop (TableDrop drops a table.); ColumnNew (ColumnNew shows add
| column form.); ColumnAdd (ColumnAdd adds a column to a table.);
| ImportForm (ImportForm shows SQL import form.); ImportRun (ImportRun
| executes imported SQL.).
| 
*/

import (
	"context"
	"net/http"
	"strconv"

	"github.com/lsgser/gofreight/controller"
	"github.com/lsgser/gofreight/database"
	"github.com/lsgser/gofreight/integrations"
	"github.com/lsgser/gofreight/router"
)

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
	a.Get("/export/:table/csv", wrap(p.TableExportCSV))
}

// SchemaNew shows the create table form.
func (p *Panel) SchemaNew(base controller.Base) error {
	data := p.baseData(base)
	data["ColumnTypes"] = database.CommonColumnTypes
	return p.render(base, "schema_new.html", data)
}

// SchemaCreate handles table creation.
func (p *Panel) SchemaCreate(base controller.Base) error {
	if err := base.Request.ParseForm(); err != nil {
		return err
	}

	table := base.Request.FormValue("table_name")
	if table == "" {
		base.Unprocessable(map[string][]string{"table_name": {"is required"}})
		return nil
	}

	columns := parseColumnDefs(base.Request)
	if len(columns) == 0 {
		base.Unprocessable(map[string][]string{"columns": {"add at least one column"}})
		return nil
	}

	if err := database.CreateTable(context.Background(), table, columns); err != nil {
		base.Unprocessable(map[string][]string{"base": {err.Error()}})
		return nil
	}

	base.Redirect(p.cfg.Prefix+"/tables/"+table+"?flash=Table+created", http.StatusSeeOther)
	return nil
}

// TableStructure shows column definitions for a table.
func (p *Panel) TableStructure(base controller.Base) error {
	table := base.Param("table")
	schema, err := database.TableSchema(context.Background(), table)
	if err != nil {
		return err
	}

	data := p.baseData(base)
	data["Table"] = table
	data["Schema"] = schema
	data["Tab"] = "structure"
	return p.render(base, "table_structure.html", data)
}

// TableDrop drops a table.
func (p *Panel) TableDrop(base controller.Base) error {
	table := base.Param("table")
	if err := database.DropTable(context.Background(), table); err != nil {
		base.Unprocessable(map[string][]string{"base": {err.Error()}})
		return nil
	}
	base.Redirect(p.cfg.Prefix+"/?flash=Table+dropped", http.StatusSeeOther)
	return nil
}

// ColumnNew shows add column form.
func (p *Panel) ColumnNew(base controller.Base) error {
	data := p.baseData(base)
	data["Table"] = base.Param("table")
	data["Tab"] = "structure"
	data["ColumnTypes"] = database.CommonColumnTypes
	return p.render(base, "column_new.html", data)
}

// ColumnAdd adds a column to a table.
func (p *Panel) ColumnAdd(base controller.Base) error {
	table := base.Param("table")
	if err := base.Request.ParseForm(); err != nil {
		return err
	}

	col := database.ColumnDef{
		Name:     base.Request.FormValue("name"),
		Type:     base.Request.FormValue("type"),
		Nullable: base.Request.FormValue("nullable") == "1",
		Default:  base.Request.FormValue("default"),
	}

	if err := database.AddColumn(context.Background(), table, col); err != nil {
		base.Unprocessable(map[string][]string{"base": {err.Error()}})
		return nil
	}

	base.Redirect(p.cfg.Prefix+"/tables/"+table+"/structure?flash=Column+added", http.StatusSeeOther)
	return nil
}

// ImportForm shows SQL import form.
func (p *Panel) ImportForm(base controller.Base) error {
	data := p.baseData(base)
	return p.render(base, "import.html", data)
}

// ImportRun executes imported SQL.
func (p *Panel) ImportRun(base controller.Base) error {
	if err := base.Request.ParseForm(); err != nil {
		return err
	}
	sqlText := base.Request.FormValue("sql")
	if err := database.ImportSQL(context.Background(), sqlText); err != nil {
		data := p.baseData(base)
		data["SQL"] = sqlText
		data["Error"] = err.Error()
		return p.render(base, "import.html", data)
	}
	base.Redirect(p.cfg.Prefix+"/?flash=SQL+imported", http.StatusSeeOther)
	return nil
}

// TableExportCSV exports table as CSV download.
func (p *Panel) TableExportCSV(base controller.Base) error {
	table := base.Param("table")
	csvText, err := database.ExportTableCSV(context.Background(), table)
	if err != nil {
		return err
	}
	base.Response.Header().Set("Content-Type", "text/csv")
	base.Response.Header().Set("Content-Disposition", "attachment; filename="+table+".csv")
	base.Response.Write([]byte(csvText))
	return nil
}

// TableExport exports table as SQL download.
func (p *Panel) TableExport(base controller.Base) error {
	table := base.Param("table")
	sqlText, err := database.ExportTableSQL(context.Background(), table)
	if err != nil {
		return err
	}

	base.Response.Header().Set("Content-Type", "application/sql")
	base.Response.Header().Set("Content-Disposition", "attachment; filename="+table+".sql")
	base.Response.Write([]byte(sqlText))
	return nil
}

// IntegrationsStatus shows configured integrations.
func (p *Panel) IntegrationsStatus(base controller.Base) error {
	_ = integrations.ConfigureAll(integrations.OsEnv{})
	all := integrations.Default().All()
	type row struct {
		Name    string
		Enabled bool
	}
	var rows []row
	for name, i := range all {
		rows = append(rows, row{Name: name, Enabled: i.Enabled()})
	}

	data := p.baseData(base)
	data["Integrations"] = rows
	return p.render(base, "integrations.html", data)
}

func parseColumnDefs(r *http.Request) []database.ColumnDef {
	count, _ := strconv.Atoi(r.FormValue("column_count"))
	if count == 0 {
		count = 5
	}

	var cols []database.ColumnDef
	for i := 0; i < count; i++ {
		prefix := "col_" + strconv.Itoa(i) + "_"
		name := r.FormValue(prefix + "name")
		if name == "" {
			continue
		}
		cols = append(cols, database.ColumnDef{
			Name:          name,
			Type:          r.FormValue(prefix + "type"),
			Nullable:      r.FormValue(prefix+"nullable") == "1",
			PrimaryKey:    r.FormValue(prefix+"pk") == "1",
			AutoIncrement: r.FormValue(prefix+"auto") == "1",
			Default:       r.FormValue(prefix + "default"),
		})
	}
	return cols
}

// ColumnDrop removes a column.
func (p *Panel) ColumnDrop(base controller.Base) error {
	table := base.Param("table")
	column := base.Param("column")
	if err := database.DropColumn(context.Background(), table, column); err != nil {
		base.Unprocessable(map[string][]string{"base": {err.Error()}})
		return nil
	}
	base.Redirect(p.cfg.Prefix+"/tables/"+table+"/structure?flash=Column+dropped", http.StatusSeeOther)
	return nil
}

// ColumnRenameForm shows rename column form.
func (p *Panel) ColumnRenameForm(base controller.Base) error {
	data := p.baseData(base)
	data["Table"] = base.Param("table")
	data["Column"] = base.Param("column")
	data["Tab"] = "structure"
	return p.render(base, "column_rename.html", data)
}

// ColumnRename renames a column.
func (p *Panel) ColumnRename(base controller.Base) error {
	table := base.Param("table")
	column := base.Param("column")
	if err := base.Request.ParseForm(); err != nil {
		return err
	}
	newName := base.Request.FormValue("new_name")
	if err := database.RenameColumn(context.Background(), table, column, newName); err != nil {
		base.Unprocessable(map[string][]string{"base": {err.Error()}})
		return nil
	}
	base.Redirect(p.cfg.Prefix+"/tables/"+table+"/structure?flash=Column+renamed", http.StatusSeeOther)
	return nil
}

// IndexAdd adds an index to a table.
func (p *Panel) IndexAdd(base controller.Base) error {
	table := base.Param("table")
	if err := base.Request.ParseForm(); err != nil {
		return err
	}
	name := base.Request.FormValue("index_name")
	column := base.Request.FormValue("column")
	unique := base.Request.FormValue("unique") == "1"
	if err := database.AddIndex(context.Background(), table, name, column, unique); err != nil {
		base.Unprocessable(map[string][]string{"base": {err.Error()}})
		return nil
	}
	base.Redirect(p.cfg.Prefix+"/tables/"+table+"/structure?flash=Index+added", http.StatusSeeOther)
	return nil
}

// TableMigration exports table schema as a migration file.
func (p *Panel) TableMigration(base controller.Base) error {
	table := base.Param("table")
	path, err := database.SaveTableAsMigration(context.Background(), "db/migrate", table)
	if err != nil {
		base.Unprocessable(map[string][]string{"base": {err.Error()}})
		return nil
	}
	base.Redirect(p.cfg.Prefix+"/tables/"+table+"/structure?flash=Migration+saved:+ "+path, http.StatusSeeOther)
	return nil
}
