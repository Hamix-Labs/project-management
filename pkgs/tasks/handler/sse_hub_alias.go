package handler

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"

// SSEHub is the concrete in-process fanout hub. Ownership lives in realtime;
// aliases keep same-package whitebox tests and historical type names compiling.
type SSEHub = realtime.SSEHub

// SSEHubOptions tunes hub construction.
type SSEHubOptions = realtime.SSEHubOptions

// NewSSEHub constructs a test-friendly hub (coalescing off).
//
//funclogmeasure:skip category=hot-path reason="Thin realtime re-export for handler whitebox tests; ownership and tracing live in realtime."
func NewSSEHub() *SSEHub { return realtime.NewSSEHub() }

// NewSSEHubWith constructs a hub with caller-supplied options.
//
//funclogmeasure:skip category=hot-path reason="Thin realtime re-export for handler whitebox tests; ownership and tracing live in realtime."
func NewSSEHubWith(opts SSEHubOptions) *SSEHub { return realtime.NewSSEHubWith(opts) }

// DefaultSSEHubOptions returns production hub defaults.
//
//funclogmeasure:skip category=hot-path reason="Thin realtime re-export for handler whitebox tests; ownership and tracing live in realtime."
func DefaultSSEHubOptions() SSEHubOptions { return realtime.DefaultSSEHubOptions() }
