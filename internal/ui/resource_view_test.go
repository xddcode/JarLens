package ui

import "testing"

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
	if runs[0].Start != 0 || runs[0].End != 15 {
		t.Fatalf("unexpected first XML tag: %+v", runs[0])
	}
}
