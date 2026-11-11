package ui

import (
	"strings"
	"unicode/utf8"

	"github.com/vivaui/viva/text"
)

func (m *Model) revealDocument(doc *document, name string) {
	if name == "" || doc == nil || doc.editor == nil {
		return
	}
	offset := declarationOffset(doc.source, name)
	if offset < 0 {
		return
	}
	start := utf8.RuneCountInString(doc.source[:offset])
	end := start + utf8.RuneCountInString(name)
	doc.editor.SetSelection(text.Range{Start: end, End: start})
	doc.code.Reveal(offset + len(name))
	doc.pendingCenter = true
	doc.pendingCenterOffset = offset
	if m.currentDocument() == doc {
		m.code = doc.code
		m.focusSource = !m.focusProject
	}
}

func nestedDeclarationName(name string) string {
	name = strings.TrimSuffix(name, ".class")
	separator := strings.LastIndex(name, "$")
	if separator < 0 {
		return name
	}
	return strings.TrimLeft(name[separator+1:], "0123456789")
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
