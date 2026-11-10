package ui

import (
	"strings"
	"testing"

	"github.com/vivaui/viva/components"
	vivaui "github.com/vivaui/viva/ui"
)

func TestLanguageForPath(t *testing.T) {
	for name, want := range map[string]string{
		"Hello.class":      "java",
		"application.xml":  "xml",
		"config.json":      "json",
		"application.yaml": "yaml",
		"MANIFEST.MF":      "properties",
		"README":           "text",
	} {
		if got := languageForPath(name); got != want {
			t.Errorf("languageForPath(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestXMLHighlighter(t *testing.T) {
	runs := highlightXMLLine(AppearanceDark, "<configuration><enabled>true</enabled></configuration>", nil)
	if len(runs) != 4 {
		t.Fatalf("XML tags were not highlighted: %+v", runs)
	}
	if runs[0].Start != 1 || runs[0].End != 14 {
		t.Fatalf("unexpected first XML tag: %+v", runs[0])
	}
}

func TestXMLAttributesCommentsAndDeclarations(t *testing.T) {
	line := `<root enabled="true" count="3"><!-- note --></root>`
	runs := highlightXMLLine(AppearanceDark, line, nil)
	for text, color := range map[string]vivaui.Color{
		"root": vivaui.RGB24(0xd5b778), "enabled": vivaui.RGB24(0x92b6e8), "count": vivaui.RGB24(0x92b6e8),
		`"true"`: vivaui.RGB24(0x6aab73), `"3"`: vivaui.RGB24(0x6aab73), "<!-- note -->": vivaui.RGB24(0x7a7e85),
	} {
		if !hasXMLRun(line, runs, text, color) {
			t.Errorf("XML %q does not use its expected color: %+v", text, runs)
		}
	}

	line = `<?xml version="1.0" encoding="UTF-8"?>`
	runs = highlightXMLLine(AppearanceDark, line, nil)
	for text, color := range map[string]vivaui.Color{
		"xml": vivaui.RGB24(0xd5b778), "version": vivaui.RGB24(0x92b6e8), "encoding": vivaui.RGB24(0x92b6e8),
		`"1.0"`: vivaui.RGB24(0x6aab73), `"UTF-8"`: vivaui.RGB24(0x6aab73),
	} {
		if !hasXMLRun(line, runs, text, color) {
			t.Errorf("XML declaration %q does not use its expected color: %+v", text, runs)
		}
	}
}

func TestXMLEntities(t *testing.T) {
	line := `<root>&amp;</root>`
	runs := highlightXMLLine(AppearanceDark, line, nil)
	if !hasXMLRun(line, runs, "&amp;", vivaui.RGB24(0x2aacb8)) {
		t.Fatalf("XML entity was not highlighted: %+v", runs)
	}
}

func hasXMLRun(line string, runs []components.CodeEditorRun, text string, color vivaui.Color) bool {
	start := strings.Index(line, text)
	if start < 0 {
		return false
	}
	for _, run := range runs {
		if run.Start == start && run.End == start+len(text) && run.Color == color {
			return true
		}
	}
	return false
}
