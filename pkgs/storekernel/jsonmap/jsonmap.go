// Package jsonmap holds RawMessage ↔ datatypes.JSON converters shared by
// BC store/model mappers. Kept as a leaf package so model packages can import
// it without cycling through storekernel (which imports some BC models).
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
