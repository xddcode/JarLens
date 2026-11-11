package ui

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/vivaui/viva/text"
)

func (m *Model) showProjectFilter() {
	m.focusProject = true
	m.filterOpen = true
	m.filterFocus = true
	m.findOpen = false
	m.findFocus = false
}

func (m *Model) showSourceFind() {
	m.focusProject = false
	m.findOpen = true
	m.findFocus = true
	m.filterOpen = false
	m.filterFocus = false
}

func (m *Model) updateFind() {
	query := m.findField.Text()
	if m.source == "" {
		m.findMatches = nil
		m.findQuery = query
		m.findSource = ""
		m.findIndex = 0
		return
	}
	if query == m.findQuery && m.source == m.findSource {
		return
	}
	m.findQuery = query
	m.findSource = m.source
	m.findMatches = findOccurrences(m.source, query)
	m.findIndex = 0
}

func findOccurrences(source, query string) []text.Span {
	if query == "" {
		return nil
	}
	lowerSource := strings.ToLower(source)
	lowerQuery := strings.ToLower(query)
	var matches []text.Span
	for offset := 0; offset < len(source); {
		index := strings.Index(lowerSource[offset:], lowerQuery)
		if index < 0 {
			break
		}
		start := offset + index
		end := start + len(query)
		matches = append(matches, text.Span{Start: start, End: end})
		offset = end
	}
	return matches
}

func (m *Model) findNext(delta int) {
	if len(m.findMatches) == 0 {
		return
	}
	m.findIndex = (m.findIndex + delta + len(m.findMatches)) % len(m.findMatches)
	span := m.findMatches[m.findIndex]
	m.editor.MoveTo(utf8.RuneCountInString(m.source[:span.Start]), false)
	m.code.Reveal(span.Start)
}

func (m *Model) findStatus() string {
	if m.findQuery == "" {
		return ""
	}
	if len(m.findMatches) == 0 {
		return m.t("无匹配项")
	}
	return fmt.Sprintf("%d/%d", m.findIndex+1, len(m.findMatches))
}

func (m *Model) goToLine(value string) bool {
	line, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || line < 1 {
		return false
	}
	lineCount := strings.Count(m.source, "\n") + 1
	line = min(line, lineCount)
	offset := 0
	for current := 1; current < line; current++ {
		index := strings.IndexByte(m.source[offset:], '\n')
		if index < 0 {
			break
		}
		offset += index + 1
	}
	m.editor.MoveTo(utf8.RuneCountInString(m.source[:offset]), false)
	m.code.Reveal(offset)
	return true
}
