package ui

import "testing"

func TestInitSetsEnabledFlag(t *testing.T) {
	prev := Enabled
	t.Cleanup(func() { Enabled = prev })

	Enabled = true
	Init(true)
	if Enabled {
		t.Fatalf("Init(true) should disable colors")
	}

	Init(false)
	if !Enabled {
		t.Fatalf("Init(false) should enable colors")
	}
}

func TestColorRespectsEnabledFlag(t *testing.T) {
	prev := Enabled
	t.Cleanup(func() { Enabled = prev })

	Enabled = true
	got := Color("hello", FgGreen)
	want := FgGreen + "hello" + Reset
	if got != want {
		t.Fatalf("Color enabled = %q, want %q", got, want)
	}

	Enabled = false
	if got := Color("world", FgMagenta); got != "world" {
		t.Fatalf("Color disabled = %q, want plain string", got)
	}
}
