package handler

import (
	"context"
	"net/http"
)

//funclogmeasure:skip category=delegate-already-logs reason="Named-payload route helper; list/get/delete handlers emit trace."
func getNamedPayload[T any](
	w http.ResponseWriter,
	r *http.Request,
	op string,
	get func(context.Context, string) (T, error),
) {
	id, err := parseTaskPathID(r.PathValue("id"))
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	row, err := get(r.Context(), id)
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	writeJSON(w, r, op, http.StatusOK, row)
}

//funclogmeasure:skip category=delegate-already-logs reason="Named-payload route helper; list/get/delete handlers emit trace."
func deleteNamedPayload(
	w http.ResponseWriter,
	r *http.Request,
	op string,
	debugIDKey string,
	deleteFn func(context.Context, string) error,
) {
	id, err := parseTaskPathID(r.PathValue("id"))
	if err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	if err := deleteFn(r.Context(), id); err != nil {
		writeStoreError(w, r, op, err)
		return
	}
	debugHTTPOut(r.Context(), op, http.StatusNoContent, debugIDKey, id, "response_empty", true)
	w.WriteHeader(http.StatusNoContent)
}
