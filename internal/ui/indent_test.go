package ui

import "testing"

func TestReindentSource(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		language string
		from     int
		to       int
		want     string
	}{
		{name: "java compact", source: "class A {\n    void run() {\n        work();\n    }\n}\n", language: "java", from: 4, to: 2, want: "class A {\n  void run() {\n    work();\n  }\n}\n"},
		{name: "java normalizes inconsistent indentation", source: "void load() {\n    if (ready) {\n        work();\n    }\n      try {\n            work();\n      } finally {\n            close();\n      }\n}", language: "java", from: 4, to: 4, want: "void load() {\n    if (ready) {\n        work();\n    }\n    try {\n        work();\n    } finally {\n        close();\n    }\n}"},
		{name: "java ignores braces in strings and comments", source: "class A {\n    String s = \"}\";\n    /* { */\n    void run() {\n        work();\n    }\n}", language: "java", from: 4, to: 4, want: "class A {\n    String s = \"}\";\n    /* { */\n    void run() {\n        work();\n    }\n}"},
		{name: "java preserves text block content", source: "class A {\n    String text = \"\"\"\n  keep { this\n    \"\"\";\n}", language: "java", from: 4, to: 4, want: "class A {\n    String text = \"\"\"\n  keep { this\n    \"\"\";\n}"},
		{name: "yaml relaxed", source: "root:\n  child:\n    value: true", language: "yaml", from: 2, to: 6, want: "root:\n      child:\n            value: true"},
		{name: "tab to spaces", source: "\tvalue", language: "yaml", from: 4, to: 2, want: "  value"},
		{name: "continuation scales", source: "      value", language: "yaml", from: 4, to: 2, want: "   value"},
		{name: "unchanged", source: "    value", language: "yaml", from: 4, to: 4, want: "    value"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := reindentSource(test.source, test.language, test.from, test.to); got != test.want {
				t.Fatalf("reindentSource() = %q, want %q", got, test.want)
			}
		})
	}
}
