package model

// ParityPair binds a domain struct prototype to its model counterpart for
// schema- and field-parity guards. BC-local registries live under
// pkgs/*/store/model/parity.go (ADR-0071); this hub slice is retained only
// for shared parity helper tests that import this package.
type ParityPair struct {
	Name   string
	Domain any
	Model  any
	Table  string
	// ModelMigrateExtra lists additional model structs AutoMigrate must run
	// before the primary model type (e.g. parent tables for association FKs).
	ModelMigrateExtra []any
}

// ParityPairs is empty — per-BC registries own schema/field parity tests.
var ParityPairs = []ParityPair{}
