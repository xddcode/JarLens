package ui

import (
	"strings"
	"unicode/utf8"
)

func (m *Model) revealDocument(doc *document, name string) {
	if name == "" || doc == nil || doc.editor == nil {
		return
	}
	offset := declarationOffset(doc.source, name)
	if offset < 0 {
		return
	}
	doc.editor.MoveTo(utf8.RuneCountInString(doc.source[:offset]), false)
	doc.code.Reveal(offset)
}

func declarationOffset(source, name string) int {
	if source == "" || name == "" {
		return -1
	}
	offset := 0
	for _, line := range strings.Split(source, "\n") {
		for _, kind := range []string{"class", "interface", "record", "enum"} {
			needle := kind + " " + name
			from := 0
			for {
				index := strings.Index(line[from:], needle)
				if index < 0 {
					break
				}
				index += from
				end := index + len(needle)
				if end == len(line) || !isIdentifierPart(line[end]) {
					return offset + index + len(kind) + 1
				}
				from = end
			}
		}
		offset += len(line) + 1
	}
	return -1
}

func isIdentifierPart(value byte) bool {
	return value == '_' || value == '$' || value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9'
}
