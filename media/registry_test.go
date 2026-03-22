package media

import (
	"testing"
)

func TestNew_AllProviders(t *testing.T) {
	for _, name := range AllProviders() {
		p, err := New(name, "test-key")
		if err != nil {
			t.Errorf("New(%q): unexpected error: %v", name, err)
			continue
		}
		if p == nil {
			t.Errorf("New(%q): returned nil provider", name)
			continue
		}
		if p.ID() == "" {
			t.Errorf("New(%q): provider has empty ID", name)
		}
	}
}

func TestNew_UnknownProvider(t *testing.T) {
	_, err := New("nonexistent", "key")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestMustNew_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for unknown provider")
		}
	}()
	MustNew("nonexistent", "key")
}

func TestMustNew_Works(t *testing.T) {
	p := MustNew(ElevenLabs, "test-key")
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestAllProviders_Count(t *testing.T) {
	providers := AllProviders()
	if len(providers) != 10 {
		t.Errorf("expected 10 providers, got %d", len(providers))
	}
}
