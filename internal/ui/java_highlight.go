package ui

import (
	"strings"
	"unicode"

	"github.com/vivaui/viva/components"
	vivaui "github.com/vivaui/viva/ui"
)

var javaKeywords = map[string]struct{}{
	"abstract": {}, "assert": {}, "boolean": {}, "break": {}, "byte": {}, "case": {}, "catch": {},
	"char": {}, "class": {}, "const": {}, "continue": {}, "default": {}, "do": {}, "double": {},
	"else": {}, "enum": {}, "extends": {}, "final": {}, "finally": {}, "float": {}, "for": {},
	"goto": {}, "if": {}, "implements": {}, "import": {}, "instanceof": {}, "int": {}, "interface": {},
	"long": {}, "native": {}, "new": {}, "package": {}, "private": {}, "protected": {}, "public": {},
	"record": {}, "return": {}, "sealed": {}, "short": {}, "static": {}, "strictfp": {}, "super": {},
	"switch": {}, "synchronized": {}, "this": {}, "throw": {}, "throws": {}, "transient": {},
	"try": {}, "var": {}, "void": {}, "volatile": {}, "while": {}, "yield": {},
	"true": {}, "false": {}, "null": {},
}

func highlightJavaLine(appearance Appearance, _ string, line string, dst []components.CodeEditorRun) []components.CodeEditorRun {
	keyword := vivaui.RGB24(0xcf8e6d)
	str := vivaui.RGB24(0x6aab73)
	comment := vivaui.RGB24(0x7a7e85)
	number := vivaui.RGB24(0x2aacb8)
	annotation := vivaui.RGB24(0xb3ae60)
	if appearance == AppearanceLight {
		keyword = vivaui.RGB24(0x0033b3)
		str = vivaui.RGB24(0x067d17)
		comment = vivaui.RGB24(0x8c8c8c)
		number = vivaui.RGB24(0x1750eb)
		annotation = vivaui.RGB24(0x9e880d)
	}

	for index := 0; index < len(line); {
		switch {
		case strings.HasPrefix(line[index:], "//"):
			return append(dst, components.CodeEditorRun{Start: index, End: len(line), Color: comment})
		case strings.HasPrefix(line[index:], "/*"):
			end := strings.Index(line[index+2:], "*/")
			if end < 0 {
				return append(dst, components.CodeEditorRun{Start: index, End: len(line), Color: comment})
			}
			end += index + 4
			dst = append(dst, components.CodeEditorRun{Start: index, End: end, Color: comment})
			index = end
		case line[index] == '"' || line[index] == '\'':
			quote := line[index]
			end := index + 1
			for end < len(line) {
				if line[end] == '\\' {
					end += 2
					continue
				}
				if line[end] == quote {
					end++
					break
				}
				end++
			}
			end = min(end, len(line))
			dst = append(dst, components.CodeEditorRun{Start: index, End: end, Color: str})
			index = end
		case line[index] == '@':
			end := index + 1
			for end < len(line) && (unicode.IsLetter(rune(line[end])) || line[end] == '_') {
				end++
			}
			dst = append(dst, components.CodeEditorRun{Start: index, End: end, Color: annotation})
			index = end
		case unicode.IsDigit(rune(line[index])):
			end := index + 1
			for end < len(line) && (unicode.IsDigit(rune(line[end])) || unicode.IsLetter(rune(line[end])) || line[end] == '_' || line[end] == '.') {
				end++
			}
			dst = append(dst, components.CodeEditorRun{Start: index, End: end, Color: number})
			index = end
		case unicode.IsLetter(rune(line[index])) || line[index] == '_':
			end := index + 1
			for end < len(line) && (unicode.IsLetter(rune(line[end])) || unicode.IsDigit(rune(line[end])) || line[end] == '_') {
				end++
			}
			if _, ok := javaKeywords[line[index:end]]; ok {
				dst = append(dst, components.CodeEditorRun{Start: index, End: end, Color: keyword})
			}
			index = end
		default:
			index++
		}
	}
	return dst
}
