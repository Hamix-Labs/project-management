package handler

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/wire"

// HandlerStore is the composed persistence contract required by pkgs/tasks/handler.
// Canonical definition: pkgs/tasks/wire.HandlerAPI.
type HandlerStore = wire.HandlerAPI
