package handler

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/wire"

// HandlerAPI is the composed persistence contract required by pkgs/tasks/handler.
// Canonical definition: pkgs/tasks/wire.HandlerAPI.
type HandlerAPI = wire.HandlerAPI

// HandlerStore is a deprecated alias of HandlerAPI (B-40 / F-COMPLEX-9).
// Prefer HandlerAPI in new code.
type HandlerStore = HandlerAPI
