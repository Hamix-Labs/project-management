package model

import (
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"
)

// NullableStructJSON marshals a nullable struct pointer for jsonb columns.
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

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func datatypesFromRaw(r json.RawMessage) datatypes.JSON {
	if len(r) == 0 {
		return datatypes.JSON("{}")
	}
	return datatypes.JSON(r)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func rawFromDatatypes(j datatypes.JSON) json.RawMessage {
	if len(j) == 0 {
		return nil
	}
	return json.RawMessage(j)
}
