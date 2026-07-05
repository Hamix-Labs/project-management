package handler

import (
	"context"
	"encoding/json"
)

// normalizeComposePayloadRaw decodes, validates against app settings, and re-marshals
// a template/draft compose payload. Callers use the returned struct for name fallback.
//
//funclogmeasure:skip category=delegate-already-logs reason="Compose validation helper; template save/patch handlers emit trace."
func (h *Handler) normalizeComposePayloadRaw(ctx context.Context, raw json.RawMessage) (json.RawMessage, taskComposePayloadJSON, error) {
	compose, err := decodeComposePayload(raw)
	if err != nil {
		return nil, taskComposePayloadJSON{}, err
	}
	settings, err := h.store.GetSettings(ctx)
	if err != nil {
		return nil, taskComposePayloadJSON{}, err
	}
	if err := h.validateComposePayload(ctx, compose, settings); err != nil {
		return nil, taskComposePayloadJSON{}, err
	}
	payloadRaw, err := composePayloadToRaw(compose)
	if err != nil {
		return nil, taskComposePayloadJSON{}, err
	}
	return payloadRaw, compose, nil
}
