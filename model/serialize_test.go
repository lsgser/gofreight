package model_test

/*
|--------------------------------------------------------------------------
| Serialize
|--------------------------------------------------------------------------
|
| Test suite for Serialize in the model package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
| 
| The model package is the ORM layer: repositories, queries, associations,
| soft deletes, validation, serialization, collections, and pagination.
| 
| Models map to tables via struct tags; migrations define schema
| separately in db/migrate.
| 
| See docs/models.md, docs/orm.md, and docs/factories.md for
| Laravel-aligned patterns.
| 
| Run with go test ./model/... or go test for this package from the
| framework root.
| 
*/

import (
	"encoding/json"
	"testing"

	"github.com/lsgser/gofreight/model"
)

type serializeUser struct {
	model.Record
	Name     string `json:"name" db:"name"`
	Email    string `json:"email" db:"email"`
	Password string `json:"-" db:"password"`
}

func TestSerializeToArray(t *testing.T) {
	u := &serializeUser{
		Record:   model.Record{ID: 1},
		Name:     "Ada",
		Email:    "ada@example.com",
		Password: "secret",
	}

	m := model.Serialize(u).ToArray()
	if m["name"] != "Ada" {
		t.Fatalf("name: %v", m["name"])
	}
	if _, ok := m["password"]; ok {
		t.Fatal("password should be hidden via json:-")
	}
}

func TestSerializeHiddenVisible(t *testing.T) {
	u := &serializeUser{Name: "Ada", Email: "ada@example.com", Password: "secret"}

	hidden := model.Serialize(u).Hidden("email").ToArray()
	if _, ok := hidden["email"]; ok {
		t.Fatal("email should be hidden")
	}

	visible := model.Serialize(u).Visible("name").ToArray()
	if len(visible) != 1 || visible["name"] != "Ada" {
		t.Fatalf("visible: %v", visible)
	}

	exposed := model.Serialize(u).Hidden("email").MakeVisible("email").ToArray()
	if exposed["email"] != "ada@example.com" {
		t.Fatal("makeVisible failed")
	}
}

func TestSerializeAppends(t *testing.T) {
	u := &serializeUser{Name: "Ada", Email: "ada@example.com"}

	data, err := model.Serialize(u).Append("is_admin", func(record any) any {
		return record.(*serializeUser).Email == "ada@example.com"
	}).ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m["is_admin"] != true {
		t.Fatalf("is_admin: %v", m["is_admin"])
	}

	without := model.Serialize(u).Append("is_admin", func(any) any { return true }).WithoutAppends().ToArray()
	if _, ok := without["is_admin"]; ok {
		t.Fatal("withoutAppends should skip appends")
	}
}

func TestCollectionToArrayJSON(t *testing.T) {
	items := []serializeUser{
		{Name: "A", Email: "a@example.com"},
		{Name: "B", Email: "b@example.com"},
	}
	col := model.NewCollection(items)

	arr := col.ToArray()
	if len(arr) != 2 || arr[0]["name"] != "A" {
		t.Fatalf("toArray: %v", arr)
	}

	raw, err := col.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	var decoded []map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded[1]["name"] != "B" {
		t.Fatalf("toJSON: %v", decoded)
	}
}
