package mcp_test

import (
	"context"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/domain"
	draftmcp "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/mcp"
	draftassiststore "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/store"
)

func TestUpdatePrompt_nonceFailClosed(t *testing.T) {
	store := draftassiststore.NewMemoryStore()
	sess, err := store.CreateSession(context.Background(), contract.CreateSessionInput{
		Snapshot: domain.FormSnapshot{Prompt: "<p>old</p>"},
	})
	if err != nil {
		t.Fatal(err)
	}
	host := &draftmcp.ToolHost{
		Bind:  &draftmcp.BindFile{SessionID: sess.ID, Nonce: sess.Nonce},
		Store: store,
	}
	if _, err := host.Store.UpdatePrompt(context.Background(), host.Bind.SessionID, "bad", "<p>new</p>"); err == nil {
		t.Fatal("expected nonce mismatch")
	}
	if _, err := host.Store.UpdatePrompt(context.Background(), host.Bind.SessionID, host.Bind.Nonce, "<p>new</p>"); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetSession(context.Background(), sess.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Snapshot.Prompt != "<p>new</p>" {
		t.Fatalf("prompt=%q", got.Snapshot.Prompt)
	}
}
