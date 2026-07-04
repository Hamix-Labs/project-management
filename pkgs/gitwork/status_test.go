package gitwork_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestCheckoutStatus_clean(t *testing.T) {
	main := initRepo(t)
	st, err := svc().CheckoutStatus(context.Background(), main)
	if err != nil {
		t.Fatalf("CheckoutStatus: %v", err)
	}
	if st.Dirty {
		t.Fatal("expected clean")
	}
	if st.Detached {
		t.Fatal("expected attached HEAD")
	}
	if st.HeadSHA == "" {
		t.Fatal("expected head sha")
	}
	if st.HeadCommitAt.IsZero() {
		t.Fatal("expected head commit time")
	}
	if st.HeadCommitAt.After(time.Now().UTC().Add(time.Minute)) {
		t.Fatalf("head commit time in future: %v", st.HeadCommitAt)
	}
}

func TestCheckoutStatus_dirty(t *testing.T) {
	main := initRepo(t)
	writeFile(t, filepath.Join(main, "dirty.txt"), "x\n")
	st, err := svc().CheckoutStatus(context.Background(), main)
	if err != nil {
		t.Fatalf("CheckoutStatus: %v", err)
	}
	if !st.Dirty {
		t.Fatal("expected dirty")
	}
}

func TestCheckoutStatus_detached(t *testing.T) {
	main := initRepo(t)
	runGit(t, main, "checkout", "--detach")
	st, err := svc().CheckoutStatus(context.Background(), main)
	if err != nil {
		t.Fatalf("CheckoutStatus: %v", err)
	}
	if !st.Detached {
		t.Fatal("expected detached")
	}
	if st.HasUpstream {
		t.Fatal("detached checkout should not report upstream")
	}
}

func TestCheckoutStatus_upstreamCounts(t *testing.T) {
	main := initRepo(t)
	repo := openRepo(t, main)
	if _, err := svc().CreateBranch(context.Background(), repo, "track-me", "main"); err != nil {
		t.Fatal(err)
	}
	runGit(t, main, "checkout", "track-me")
	runGit(t, main, "branch", "--set-upstream-to", "main", "track-me")

	st, err := svc().CheckoutStatus(context.Background(), main)
	if err != nil {
		t.Fatalf("CheckoutStatus: %v", err)
	}
	if !st.HasUpstream {
		t.Fatal("expected upstream")
	}
	if st.Upstream == "" {
		t.Fatal("expected upstream name")
	}
	if st.Ahead != 0 || st.Behind != 0 {
		t.Fatalf("same commit as upstream: ahead=%d behind=%d", st.Ahead, st.Behind)
	}
}
