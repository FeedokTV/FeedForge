package profile_test

import (
	"feedforge/internal/profile"
	"testing"
)

func TestProfileLoadBuiltin(t *testing.T) {
	builtinName := "urlhaus"

	got, err := profile.LoadBuiltin(builtinName)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got == nil {
		t.Fatal("got nil profile")
	}

	if got.Name != builtinName {
		t.Errorf("got %q, want %q", got.Name, builtinName)
	}
}

func TestProfileLoadBuiltinUnknownNotFOund(t *testing.T) {
	builtinName := "nonexisting!!!"

	got, err := profile.LoadBuiltin(builtinName)
	if err == nil {
		t.Errorf("want error, but got nil")
	}

	if got != nil {
		t.Errorf("want: nil, got: %v", got.Name)
	}
}
