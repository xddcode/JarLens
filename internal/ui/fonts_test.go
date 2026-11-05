package ui

import (
	"testing"

	vivaui "github.com/vivaui/viva/ui"
)

func TestBundledFonts(t *testing.T) {
	configureFonts()
	if fontErr != nil {
		t.Fatal(fontErr)
	}
	if monoFace == vivaui.MonoSystem {
		t.Fatal("monospace font was not registered")
	}
}
