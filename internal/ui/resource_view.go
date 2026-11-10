package ui

import (
	"path/filepath"
	"strings"

	"github.com/vivaui/viva/components"
	vivaui "github.com/vivaui/viva/ui"
)

func languageForPath(filePath string) string {
	name := strings.ToLower(filepath.Base(filePath))
	switch filepath.Ext(name) {
	case ".java", ".class":
		return "java"
	case ".xml", ".xsd", ".html", ".htm":
		return "xml"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".md", ".markdown":
		return "markdown"
	case ".properties", ".ini", ".toml", ".conf", ".cfg", ".mf":
		return "properties"
	default:
		return "text"
	}
}

func highlightDocumentLine(appearance Appearance, language, line string, dst []components.CodeEditorRun) []components.CodeEditorRun {
	switch language {
	case "java":
		return highlightJavaLine(appearance, language, line, dst)
	case "xml":
		return highlightXMLLine(appearance, line, dst)
	default:
		return dst
	}
}

func highlightXMLLine(appearance Appearance, line string, dst []components.CodeEditorRun) []components.CodeEditorRun {
	tag := vivaui.RGB24(0xd5b778)
	attribute := vivaui.RGB24(0x92b6e8)
	value := vivaui.RGB24(0x6aab73)
	comment := vivaui.RGB24(0x7a7e85)
	entity := vivaui.RGB24(0x2aacb8)
	if appearance == AppearanceLight {
		tag = vivaui.RGB24(0x0033b3)
		attribute = vivaui.RGB24(0x174ad4)
		value = vivaui.RGB24(0x067d17)
		comment = vivaui.RGB24(0x8c8c8c)
		entity = vivaui.RGB24(0x1750eb)
	}
	for index := 0; index < len(line); {
		switch {
		case strings.HasPrefix(line[index:], "<!--"):
			end := strings.Index(line[index+4:], "-->")
			if end < 0 {
				return append(dst, components.CodeEditorRun{Start: index, End: len(line), Color: comment})
			}
			end += index + 7
			dst = append(dst, components.CodeEditorRun{Start: index, End: end, Color: comment})
			index = end
		case line[index] == '&':
			end := index + 1
			for end < len(line) && xmlNameCharacter(line[end]) {
				end++
			}
			if end < len(line) && line[end] == ';' {
				dst = append(dst, components.CodeEditorRun{Start: index, End: end + 1, Color: entity})
				index = end + 1
			} else {
				index++
			}
		case line[index] == '<':
			index++
			if index < len(line) && (line[index] == '/' || line[index] == '?' || line[index] == '!') {
				index++
			}
			start := index
			for index < len(line) && xmlNameCharacter(line[index]) {
				index++
			}
			if start < index {
				dst = append(dst, components.CodeEditorRun{Start: start, End: index, Color: tag})
			}
			for index < len(line) && line[index] != '>' {
				switch {
				case line[index] == '"' || line[index] == '\'':
					quote := line[index]
					start = index
					index++
					for index < len(line) && line[index] != quote {
						index++
					}
					if index < len(line) {
						index++
					}
					dst = append(dst, components.CodeEditorRun{Start: start, End: index, Color: value})
				case xmlNameCharacter(line[index]):
					start = index
					for index < len(line) && xmlNameCharacter(line[index]) {
						index++
					}
					dst = append(dst, components.CodeEditorRun{Start: start, End: index, Color: attribute})
				default:
					index++
				}
			}
			if index < len(line) {
				index++
			}
		default:
			index++
		}
	}
	return dst
}

func xmlNameCharacter(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9' || value == '_' || value == '-' || value == ':' || value == '.'
}
