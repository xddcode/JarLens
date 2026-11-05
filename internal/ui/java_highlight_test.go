package ui

import (
	"testing"

	"github.com/vivaui/viva/components"
	vivaui "github.com/vivaui/viva/ui"
)

func TestJavaHighlighter(t *testing.T) {
	line := `public String greeting = "hello"; // comment`
	runs := highlightJavaLine(AppearanceDark, "java", line, nil)
	assertRun(t, runs, 0, 6)
	assertRun(t, runs, 25, 32)
	assertRun(t, runs, 34, len(line))
}

func TestJavaHighlighterLightTheme(t *testing.T) {
	dark := highlightJavaLine(AppearanceDark, "java", "public", nil)
	light := highlightJavaLine(AppearanceLight, "java", "public", nil)
	if dark[0].Color == light[0].Color {
		t.Fatal("syntax colors did not follow the appearance")
	}
	if light[0].Color != vivaui.RGB24(0x0033b3) {
		t.Fatal("light keyword color does not match the Islands palette")
	}
}

func assertRun(t *testing.T, runs []components.CodeEditorRun, start, end int) {
	t.Helper()
	for _, run := range runs {
		if run.Start == start && run.End == end {
			return
		}
	}
	t.Fatalf("missing highlighted run %d..%d in %+v", start, end, runs)
}
