// Package parity holds domain↔model field-parity helpers shared by BC
// store/model packages. Kept as a leaf package so model packages can import
// it without cycling through storekernel (which imports some BC models).
package parity

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm/schema"
)

// Pair binds a domain struct prototype to its model counterpart for
// schema- and field-parity guards.
type Pair struct {
	Name   string
	Domain any
	Model  any
	Table  string
	// ModelMigrateExtra lists additional model structs AutoMigrate must run
	// before the primary model type (e.g. parent tables for association FKs).
	ModelMigrateExtra []any
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func gormColumnName(sf reflect.StructField) (string, bool) {
	tag := sf.Tag.Get("gorm")
	if tag == "" {
		return "", false
	}
	if tag == "-" {
		return "", false
	}
	parts := strings.Split(tag, ";")
	for _, p := range parts {
		if strings.HasPrefix(p, "column:") {
			return strings.TrimPrefix(p, "column:"), true
		}
	}
	if strings.Contains(tag, "foreignKey:") || strings.Contains(tag, "references:") {
		return "", false
	}
	return schema.NamingStrategy{}.ColumnName("", sf.Name), true
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func isAssociationField(sf reflect.StructField) bool {
	tag := sf.Tag.Get("gorm")
	if strings.Contains(tag, "foreignKey:") || strings.Contains(tag, "references:") {
		return true
	}
	if sf.Type.Kind() == reflect.Pointer && sf.Type.Elem().Kind() == reflect.Struct {
		if !strings.Contains(tag, "column:") && tag != "-" {
			return true
		}
	}
	return false
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func modelPersistedFields(modelType reflect.Type) []reflect.StructField {
	var out []reflect.StructField
	for i := 0; i < modelType.NumField(); i++ {
		sf := modelType.Field(i)
		if !sf.IsExported() {
			continue
		}
		if isAssociationField(sf) {
			continue
		}
		if _, ok := gormColumnName(sf); !ok {
			continue
		}
		out = append(out, sf)
	}
	return out
}

var domainHydrationOnly = map[string]map[string]bool{
	"Task": {
		"DependsOn": true,
		"CreatedAt": true,
	},
}

// modelPersistOnly lists DB columns intentionally absent from domain wire types.
var modelPersistOnly = map[string]map[string]bool{
	"TaskTemplate": {
		"InstantiateCount": true,
	},
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func domainFieldByName(domainType reflect.Type, name string) (reflect.StructField, bool) {
	for i := 0; i < domainType.NumField(); i++ {
		sf := domainType.Field(i)
		if sf.Name == name {
			return sf, true
		}
	}
	return reflect.StructField{}, false
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func typesCompatible(domainType, modelType reflect.Type) bool {
	if domainType == modelType {
		return true
	}
	raw := reflect.TypeOf(json.RawMessage(nil))
	dt := reflect.TypeOf(datatypes.JSON(nil))
	if domainType == raw && modelType == dt {
		return true
	}
	if domainType.Kind() == reflect.Slice && modelType.Kind() == reflect.Slice {
		return domainType.Elem() == modelType.Elem()
	}
	if domainType.Kind() == reflect.Pointer && modelType.Kind() == reflect.Pointer {
		return typesCompatible(domainType.Elem(), modelType.Elem())
	}
	return false
}

// AssertFieldParity reports whether every persisted model column has a domain
// counterpart with a compatible Go type (json.RawMessage pairs with datatypes.JSON).
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func AssertFieldParity(pair Pair) error {
	dt := reflect.TypeOf(pair.Domain)
	if dt.Kind() == reflect.Pointer {
		dt = dt.Elem()
	}
	mt := reflect.TypeOf(pair.Model)
	if mt.Kind() == reflect.Pointer {
		mt = mt.Elem()
	}
	skip := domainHydrationOnly[pair.Name]
	persistOnly := modelPersistOnly[pair.Name]

	for _, mf := range modelPersistedFields(mt) {
		col, _ := gormColumnName(mf)
		if persistOnly != nil && persistOnly[mf.Name] {
			continue
		}
		df, ok := domainFieldByName(dt, mf.Name)
		if !ok {
			return fmt.Errorf("%s: model field %s (column %q) missing on domain", pair.Name, mf.Name, col)
		}
		if skip != nil && skip[df.Name] {
			return fmt.Errorf("%s: domain field %s is hydration-only but model persists column %q", pair.Name, df.Name, col)
		}
		if !typesCompatible(df.Type, mf.Type) {
			return fmt.Errorf("%s: column %q type mismatch domain=%s model=%s", pair.Name, col, df.Type, mf.Type)
		}
	}
	return nil
}

// SortedStrings returns a sorted copy of ss.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func SortedStrings(ss []string) []string {
	cp := append([]string(nil), ss...)
	sort.Strings(cp)
	return cp
}
