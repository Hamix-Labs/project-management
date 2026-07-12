package handler

import (
	"context"
	"encoding/json"
	taskcorehandler "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/handler"
)

// normalizeComposePayloadRaw decodes, validates against app settings, and re-marshals
// a template/draft compose payload. Callers use the returned struct for name fallback.
//
//funclogmeasure:skip category=delegate-already-logs reason="Compose validation helper; template save/patch handlers emit trace."
func (h *Handler) normalizeComposePayloadRaw(ctx context.Context, raw json.RawMessage) (json.RawMessage, taskcorehandler.TaskComposePayloadJSON, error) {
	compose, err := taskcorehandler.DecodeComposePayload(raw)
	if err != nil {
		return nil, taskcorehandler.TaskComposePayloadJSON{}, err
	}
	settings, err := h.store.GetSettings(ctx)
	if err != nil {
		return nil, taskcorehandler.TaskComposePayloadJSON{}, err
	}
	if err := h.validateComposePayload(ctx, compose, settings); err != nil {
		return nil, taskcorehandler.TaskComposePayloadJSON{}, err
	}
	payloadRaw, err := taskcorehandler.ComposePayloadToRaw(compose)
	if err != nil {
		return nil, taskcorehandler.TaskComposePayloadJSON{}, err
	}
	return payloadRaw, compose, nil
}
