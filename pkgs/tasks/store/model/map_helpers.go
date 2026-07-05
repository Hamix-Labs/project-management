package model

// MapSlice maps each element of rows through fn, returning nil for an empty input slice.
//
//funclogmeasure:skip category=hot-path reason="Pure mapper helper without I/O; entity mappers emit trace at store boundary."
func MapSlice[A, B any](rows []A, fn func(A) B) []B {
	if len(rows) == 0 {
		return nil
	}
	out := make([]B, len(rows))
	for i := range rows {
		out[i] = fn(rows[i])
	}
	return out
}

// MapPtr applies fn to *p when non-nil; otherwise returns nil.
//
//funclogmeasure:skip category=hot-path reason="Pure mapper helper without I/O; entity mappers emit trace at store boundary."
func MapPtr[A, B any](p *A, fn func(A) B) *B {
	if p == nil {
		return nil
	}
	b := fn(*p)
	return &b
}
