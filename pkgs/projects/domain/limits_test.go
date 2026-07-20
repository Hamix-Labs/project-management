package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateProjectContextTitle(t *testing.T) {
	t.Parallel()
	if err := ValidateProjectContextTitle(""); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("empty title err = %v", err)
	}
	if err := ValidateProjectContextTitle("ok"); err != nil {
		t.Fatalf("short title: %v", err)
	}
	long := strings.Repeat("ä", MaxProjectContextTitleChars+1)
	if err := ValidateProjectContextTitle(long); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("oversize title err = %v", err)
	}
	atLimit := strings.Repeat("a", MaxProjectContextTitleChars)
	if err := ValidateProjectContextTitle(atLimit); err != nil {
		t.Fatalf("at-limit title: %v", err)
	}
}

func TestValidateProjectContextBody(t *testing.T) {
	t.Parallel()
	if err := ValidateProjectContextBody(""); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("empty body err = %v", err)
	}
	if err := ValidateProjectContextBody("note"); err != nil {
		t.Fatalf("short body: %v", err)
	}
	over := strings.Repeat("x", MaxProjectContextBodyBytes+1)
	if err := ValidateProjectContextBody(over); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("oversize body err = %v", err)
	}
	atLimit := strings.Repeat("y", MaxProjectContextBodyBytes)
	if err := ValidateProjectContextBody(atLimit); err != nil {
		t.Fatalf("at-limit body: %v", err)
	}
}

func TestValidateProjectContextDescription(t *testing.T) {
	t.Parallel()
	if err := ValidateProjectContextDescription(""); err != nil {
		t.Fatalf("empty description: %v", err)
	}
	if err := ValidateProjectContextDescription("short blurb for selection"); err != nil {
		t.Fatalf("short description: %v", err)
	}
	long := strings.Repeat("ä", MaxProjectContextDescriptionChars+1)
	if err := ValidateProjectContextDescription(long); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("oversize description err = %v", err)
	}
	atLimit := strings.Repeat("a", MaxProjectContextDescriptionChars)
	if err := ValidateProjectContextDescription(atLimit); err != nil {
		t.Fatalf("at-limit description: %v", err)
	}
}
