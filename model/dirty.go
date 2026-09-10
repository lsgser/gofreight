package model

/*
|--------------------------------------------------------------------------
| Dirty
|--------------------------------------------------------------------------
|
| Implements Dirty as part of the model package in the Gofreight
| framework. Key symbols: Dirty, Snapshot, Changed, ChangedFields,
| Changes, ChangedAttr.
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
| Symbols defined here include: Dirty (exported type); Snapshot (Snapshot
| captures the current state for change detection.); Changed (Changed
| returns true if any attribute has changed since the snapshot.);
| ChangedFields (ChangedFields returns the names of changed attributes.);
| Changes (Changes returns a map of changed attributes with [original,
| current] values.); ChangedAttr (ChangedAttr returns true if a specific
| attribute changed.); Previous (Previous returns the original value of an
| attribute.); Clear (Clear resets dirty tracking after save.).
| 
*/

import (
	"reflect"
)

// Dirty tracks attribute changes on a model between reads and saves.
type Dirty struct {
	original map[string]any
	current  any
}

// Snapshot captures the current state for change detection.
func (d *Dirty) Snapshot(record any) {
	d.original = recordToMap(record)
	d.current = record
}

// Changed returns true if any attribute has changed since the snapshot.
func (d *Dirty) Changed() bool {
	if d.original == nil {
		return false
	}
	current := recordToMap(d.current)
	for k, origVal := range d.original {
		if curVal, ok := current[k]; ok {
			if !reflect.DeepEqual(origVal, curVal) {
				return true
			}
		}
	}
	return false
}

// ChangedFields returns the names of changed attributes.
func (d *Dirty) ChangedFields() []string {
	if d.original == nil {
		return nil
	}
	current := recordToMap(d.current)
	var fields []string
	for k, origVal := range d.original {
		if curVal, ok := current[k]; ok {
			if !reflect.DeepEqual(origVal, curVal) {
				fields = append(fields, k)
			}
		}
	}
	return fields
}

// Changes returns a map of changed attributes with [original, current] values.
func (d *Dirty) Changes() map[string][2]any {
	if d.original == nil {
		return nil
	}
	current := recordToMap(d.current)
	changes := make(map[string][2]any)
	for k, origVal := range d.original {
		if curVal, ok := current[k]; ok {
			if !reflect.DeepEqual(origVal, curVal) {
				changes[k] = [2]any{origVal, curVal}
			}
		}
	}
	return changes
}

// ChangedAttr returns true if a specific attribute changed.
func (d *Dirty) ChangedAttr(column string) bool {
	if d.original == nil {
		return false
	}
	current := recordToMap(d.current)
	origVal, ok := d.original[column]
	if !ok {
		return false
	}
	curVal, ok := current[column]
	if !ok {
		return false
	}
	return !reflect.DeepEqual(origVal, curVal)
}

// Previous returns the original value of an attribute.
func (d *Dirty) Previous(column string) any {
	if d.original == nil {
		return nil
	}
	return d.original[column]
}

// Clear resets dirty tracking after save.
func (d *Dirty) Clear(record any) {
	d.Snapshot(record)
}

// Dirtyable models can embed Dirty for change tracking.
type Dirtyable struct {
	Dirty Dirty
}

// TrackChanges snapshots the record for dirty tracking.
func (d *Dirtyable) TrackChanges(record any) {
	d.Dirty.Snapshot(record)
}

// SoftDeletable enables soft delete on a model.
type SoftDeletable struct {
	DeletedAt *string `db:"deleted_at" json:"deleted_at,omitempty"`
}

// IsDeleted returns true if the record is soft-deleted.
func (s *SoftDeletable) IsDeleted() bool {
	return s.DeletedAt != nil && *s.DeletedAt != ""
}
