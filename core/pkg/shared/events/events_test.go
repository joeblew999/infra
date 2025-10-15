package events

import "testing"

type sample struct {
	Name string
}

func TestWrapAndDecode(t *testing.T) {
	env, err := Wrap("core.test", "created", sample{Name: "demo"})
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	var out sample
	if err := env.Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Name != "demo" {
		t.Fatalf("expected demo got %s", out.Name)
	}
	if env.Subject != "core.test" {
		t.Fatalf("unexpected subject %s", env.Subject)
	}
}

func TestOptions(t *testing.T) {
	env, err := Wrap("core.test", "meta", sample{}, WithMetadata(map[string]string{"env": "test"}), WithID("fixed"))
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	if env.ID != "fixed" {
		t.Fatalf("expected fixed id got %s", env.ID)
	}
	if env.Metadata["env"] != "test" {
		t.Fatalf("expected metadata")
	}
}
