// Package jsonmap holds RawMessage ↔ datatypes.JSON converters shared by
// BC store/model mappers. Kept as a leaf package so model packages can import
// it without pulling the rest of storekernel.
package jsonmap

import (
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"
)

// NullableStructJSON marshals a nullable struct pointer for jsonb columns.
// nil returns nil datatypes.JSON so GORM writes SQL NULL instead of invalid empty json.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func NullableStructJSON[T any](v *T) (datatypes.JSON, error) {
	if v == nil {
		return nil, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}
	return datatypes.JSON(b), nil
}

// DatatypesFromRaw maps json.RawMessage to datatypes.JSON, using "{}" for empty input.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func DatatypesFromRaw(r json.RawMessage) datatypes.JSON {
	if len(r) == 0 {
		return datatypes.JSON("{}")
	}
	return datatypes.JSON(r)
}

// RawFromDatatypes maps datatypes.JSON to json.RawMessage, returning nil for empty input.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func RawFromDatatypes(j datatypes.JSON) json.RawMessage {
	if len(j) == 0 {
		return nil
	}
	return json.RawMessage(j)
}

// JSONRawObject returns a canonical empty JSON object as RawMessage.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func JSONRawObject() json.RawMessage {
	return json.RawMessage("{}")
}

// RawJSONObjectFromDatatypes maps datatypes.JSON to RawMessage, using "{}" for empty input.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func RawJSONObjectFromDatatypes(j datatypes.JSON) json.RawMessage {
	if len(j) == 0 {
		return JSONRawObject()
	}
	return json.RawMessage(j)
}

// JSONStringSlice maps a string slice to datatypes.JSONSlice for jsonb columns.
// nil and empty both become a non-nil empty slice so GormValue writes "[]"
// (a nil JSONSlice writes JSON null / empty driver values that Postgres rejects
// on NOT NULL jsonb columns when paired with GORM serializer:json).
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func JSONStringSlice(s []string) datatypes.JSONSlice[string] {
	if s == nil {
		s = []string{}
	}
	return datatypes.NewJSONSlice(s)
}

// StringSliceFromJSON maps datatypes.JSONSlice back to a plain []string.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func StringSliceFromJSON(s datatypes.JSONSlice[string]) []string {
	if len(s) == 0 {
		return []string{}
	}
	return append([]string(nil), s...)
}
