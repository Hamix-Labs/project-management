package model

import (
	"encoding/json"

	"gorm.io/datatypes"
)

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

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func jsonRawObject() json.RawMessage {
	return json.RawMessage("{}")
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func rawJSONObjectFromDatatypes(j datatypes.JSON) json.RawMessage {
	if len(j) == 0 {
		return jsonRawObject()
	}
	return json.RawMessage(j)
}
