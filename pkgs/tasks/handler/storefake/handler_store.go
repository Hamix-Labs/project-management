package storefake

// HandlerStoreFake composes TaskCRUDFake with stub implementations for the
// remaining HandlerStore slices. Use NewHandlerStore for handler tests that
// only exercise one store slice (e.g. GET /tasks/{id} → Get).
type HandlerStoreFake struct {
	*TaskCRUDFake
	unimplementedHandlerStore
}

// NewHandlerStore returns a HandlerStoreFake with an embedded TaskCRUDFake.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func NewHandlerStore() *HandlerStoreFake {
	return &HandlerStoreFake{TaskCRUDFake: NewTaskCRUD()}
}

// NewHandlerStoreFromTaskCRUD wraps an existing TaskCRUDFake.
//
//funclogmeasure:skip category=tool-required-noop reason="Handler test fake only; store I/O traces live on production HTTP handler chokepoints."
func NewHandlerStoreFromTaskCRUD(crud *TaskCRUDFake) *HandlerStoreFake {
	if crud == nil {
		crud = NewTaskCRUD()
	}
	return &HandlerStoreFake{TaskCRUDFake: crud}
}

type unimplementedHandlerStore struct{}
