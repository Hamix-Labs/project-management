package storefake

import "errors"

// errNotImplemented is returned by TaskCRUDFake methods that the test has not
// configured. HandlerStoreFake slice stubs return the same sentinel.
var errNotImplemented = errors.New("storefake: not implemented")
