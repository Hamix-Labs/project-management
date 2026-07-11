package model

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	taskeventsdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"
)

func TestTaskEvent_roundTrip(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	resp := "ack"
	respAt := now.Add(time.Minute)
	data := json.RawMessage(`{"status":"ready"}`)
	thread := json.RawMessage(`[{"at":"2026-03-01T12:01:00Z","by":"user","body":"hi"}]`)
	orig := taskeventsdomain.TaskEvent{
		TaskID:         "task-1",
		Seq:            2,
		At:             now,
		Type:           taskeventsdomain.EventStatusChanged,
		By:             taskeventsdomain.ActorUser,
		Data:           data,
		UserResponse:   &resp,
		UserResponseAt: &respAt,
		ResponseThread: thread,
	}
	m := FromDomainTaskEvent(orig)
	m2 := FromDomainTaskEvent(ToDomainTaskEvent(m))
	if !taskEventModelEqual(m, m2) {
		t.Fatal("model round-trip mismatch")
	}
	back := ToDomainTaskEvent(m)
	if !jsonEqual(data, back.Data) || !jsonEqual(thread, back.ResponseThread) {
		t.Fatalf("json columns diverged: data=%s thread=%s", back.Data, back.ResponseThread)
	}
	if back.UserResponse == nil || *back.UserResponse != resp {
		t.Fatalf("user response: got %v", back.UserResponse)
	}
}

func TestTaskEvent_nilOptionalFields(t *testing.T) {
	t.Parallel()
	orig := taskeventsdomain.TaskEvent{
		TaskID: "t",
		Seq:    1,
		At:     time.Now().UTC(),
		Type:   taskeventsdomain.EventTaskCreated,
		By:     taskeventsdomain.ActorAgent,
		Data:   json.RawMessage(`{}`),
	}
	m := FromDomainTaskEvent(orig)
	m2 := FromDomainTaskEvent(ToDomainTaskEvent(m))
	if !taskEventModelEqual(m, m2) {
		t.Fatal("nil optional fields round-trip failed")
	}
}

func taskEventModelEqual(a, b TaskEvent) bool {
	return a.TaskID == b.TaskID &&
		a.Seq == b.Seq &&
		a.At.Equal(b.At) &&
		a.Type == b.Type &&
		a.By == b.By &&
		jsonEqual(a.Data, b.Data) &&
		jsonEqual(a.ResponseThread, b.ResponseThread) &&
		ptrStrEqual(a.UserResponse, b.UserResponse) &&
		ptrTimeEqual(a.UserResponseAt, b.UserResponseAt)
}

func jsonEqual(a, b []byte) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	var ja, jb any
	if err := json.Unmarshal(a, &ja); err != nil {
		return bytes.Equal(a, b)
	}
	if err := json.Unmarshal(b, &jb); err != nil {
		return bytes.Equal(a, b)
	}
	ma, _ := json.Marshal(ja)
	mb, _ := json.Marshal(jb)
	return bytes.Equal(ma, mb)
}

func ptrStrEqual(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func ptrTimeEqual(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Equal(*b)
}
