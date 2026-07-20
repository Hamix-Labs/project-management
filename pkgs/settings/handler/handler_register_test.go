package handler

import (
	"net/http"
	"testing"
)

func TestRegister_requiresNonNilGit(t *testing.T) {
	t.Parallel()
	defer func() {
		if recover() == nil {
			t.Fatal("Register must panic when Deps.Git is nil")
		}
	}()
	Register(http.NewServeMux(), Deps{})
}
