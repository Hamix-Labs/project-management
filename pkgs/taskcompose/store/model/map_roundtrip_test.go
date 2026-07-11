package model

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/domain"
)

func TestTaskDraft_roundTrip(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	payload := json.RawMessage(`{"title":"draft"}`)
	orig := domain.TaskDraft{
		ID: "draft-1", Name: "My draft", PayloadJSON: payload,
		CreatedAt: now, UpdatedAt: now,
	}
	m := FromDomainTaskDraft(orig)
	back := ToDomainTaskDraft(m)
	if !reflect.DeepEqual(orig, back) {
		t.Fatalf("round-trip mismatch: %+v vs %+v", orig, back)
	}
}

func TestTaskTemplate_roundTrip(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	payload := json.RawMessage(`{"title":"template"}`)
	orig := domain.TaskTemplate{
		ID: "tmpl-1", Name: "My template", PayloadJSON: payload,
		CreatedAt: now, UpdatedAt: now,
	}
	m := FromDomainTaskTemplate(orig)
	back := ToDomainTaskTemplate(m)
	if !reflect.DeepEqual(orig, back) {
		t.Fatalf("round-trip mismatch: %+v vs %+v", orig, back)
	}
}
