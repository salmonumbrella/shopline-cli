package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/salmonumbrella/shopline-cli/internal/secrets"
)

type runtimeProfileStoreMock struct {
	names []string
	creds map[string]*secrets.StoreCredentials
}

func (m *runtimeProfileStoreMock) List() ([]string, error) {
	return append([]string(nil), m.names...), nil
}

func (m *runtimeProfileStoreMock) Get(name string) (*secrets.StoreCredentials, error) {
	if c, ok := m.creds[name]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("not found")
}

func TestResolveStoreCredentials_ProfileNotFoundIncludesDesirePath(t *testing.T) {
	store := &runtimeProfileStoreMock{
		names: []string{"demoshop"},
		creds: map[string]*secrets.StoreCredentials{
			"demoshop": {Name: "demoshop", Handle: "demoshop"},
		},
	}

	_, err := resolveStoreCredentials(store, "does-not-exist")
	if err == nil {
		t.Fatal("expected error for missing profile")
	}

	msg := err.Error()
	for _, want := range []string{
		"profile not found: does-not-exist",
		"spl auth ls",
		"spl auth login",
	} {
		if !strings.Contains(msg, want) {
			t.Fatalf("expected %q in error, got: %s", want, msg)
		}
	}
}

func TestResolveStoreCredentials_AmbiguousMatchIncludesSelectionGuidance(t *testing.T) {
	store := &runtimeProfileStoreMock{
		names: []string{"demo-ca", "demo-us"},
		creds: map[string]*secrets.StoreCredentials{
			"demo-ca": {Name: "demo-ca", Handle: "demoshop"},
			"demo-us": {Name: "demo-us", Handle: "demousa"},
		},
	}

	_, err := resolveStoreCredentials(store, "demo")
	if err == nil {
		t.Fatal("expected error for ambiguous profile")
	}

	msg := err.Error()
	for _, want := range []string{
		"profile not found: demo",
		"multiple matches:",
		"use --store with an exact profile name",
		"spl auth ls",
	} {
		if !strings.Contains(msg, want) {
			t.Fatalf("expected %q in error, got: %s", want, msg)
		}
	}
}

func TestResolveStoreCredentials_PrefixMatchPrintsStderr(t *testing.T) {
	store := &runtimeProfileStoreMock{
		names: []string{"production"},
		creds: map[string]*secrets.StoreCredentials{
			"production": {Name: "production", Handle: "prod-store"},
		},
	}

	var buf bytes.Buffer
	oldWriter := stderrWriter
	stderrWriter = &buf
	defer func() { stderrWriter = oldWriter }()

	creds, err := resolveStoreCredentials(store, "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.Name != "production" {
		t.Fatalf("expected production, got %s", creds.Name)
	}
	if !strings.Contains(buf.String(), "production") {
		t.Errorf("expected stderr to mention resolved profile, got: %q", buf.String())
	}
}

func TestResolveStoreAlias_CaseInsensitive(t *testing.T) {
	t.Setenv("SHOPLINE_STORE_ALIASES", "DS:demoshop,TS:testshop")

	tests := []struct{ input, want string }{
		{"ds", "demoshop"},
		{"DS", "demoshop"},
		{"Ds", "demoshop"},
		{"ts", "testshop"},
		{"TS", "testshop"},
		{"unknown", "unknown"},
	}
	for _, tt := range tests {
		got := resolveStoreAlias(tt.input)
		if got != tt.want {
			t.Errorf("resolveStoreAlias(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
