package ui

import (
	"path/filepath"
	"strings"
	"unicode"

	vivaui "github.com/vivaui/viva/ui"
	"github.com/xddcode/jarlens/internal/archive"
)

type classLink struct {
	start   int
	end     int
	entryID int
}

type classLinkIndex struct {
	qualified map[string]int
	simple    map[string]int
}

func newClassLinkIndex(entries []archive.Entry) *classLinkIndex {
	index := &classLinkIndex{qualified: make(map[string]int), simple: make(map[string]int)}
	for _, entry := range entries {
		if entry.Kind != archive.EntryClass {
			continue
		}
		name := strings.TrimSuffix(entry.Path, filepath.Ext(entry.Path))
		index.qualified[strings.ReplaceAll(name, "/", ".")] = entry.ID
		simple := strings.TrimSuffix(entry.Name, filepath.Ext(entry.Name))
		if dot := strings.LastIndex(simple, "$"); dot >= 0 {
			simple = simple[dot+1:]
		}
		if _, exists := index.simple[simple]; !exists {
			index.simple[simple] = entry.ID
		}
	}
	return index
}

func buildClassLinks(source, filePath string, index *classLinkIndex) []classLink {
	if index == nil {
		return nil
	}
	if strings.HasPrefix(filepath.ToSlash(filePath), "META-INF/services/") {
		return serviceClassLinks(source, index)
	}
	if !strings.EqualFold(filepath.Ext(filePath), ".java") && !strings.EqualFold(filepath.Ext(filePath), ".class") {
		return nil
	}
	return javaClassLinks(source, index)
}

func serviceClassLinks(source string, index *classLinkIndex) []classLink {
	var links []classLink
	for start := 0; start < len(source); {
		end := strings.IndexByte(source[start:], '\n')
		if end < 0 {
			end = len(source)
		} else {
			end += start
		}
		line := source[start:end]
		left := strings.TrimLeft(line, " \t\r")
		trimmed := strings.TrimSpace(left)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			if id, ok := index.qualified[trimmed]; ok {
				begin := start + len(line) - len(left)
				links = append(links, classLink{start: begin, end: begin + len(trimmed), entryID: id})
			}
		}
		if end == len(source) {
			break
		}
		start = end + 1
	}
	return links
}

func javaClassLinks(source string, index *classLinkIndex) []classLink {
	var links []classLink
	for at := 0; at < len(source); {
		switch {
		case strings.HasPrefix(source[at:], "//"):
			end := strings.IndexByte(source[at:], '\n')
			if end < 0 {
				return links
			}
			at += end + 1
		case strings.HasPrefix(source[at:], "/*"):
			end := strings.Index(source[at+2:], "*/")
			if end < 0 {
				return links
			}
			at += end + 4
		case source[at] == '"' || source[at] == '\'':
			quote := source[at]
			at++
			for at < len(source) {
				if source[at] == '\\' {
					at += 2
					continue
				}
				if source[at] == quote {
					at++
					break
				}
				at++
			}
		case unicode.IsLetter(rune(source[at])) || source[at] == '_' || source[at] == '$':
			start := at
			for at < len(source) && (unicode.IsLetter(rune(source[at])) || unicode.IsDigit(rune(source[at])) || source[at] == '_' || source[at] == '$' || source[at] == '.') {
				at++
			}
			token := strings.TrimRight(source[start:at], ".")
			if id, ok := index.qualified[token]; ok {
				links = append(links, classLink{start: start, end: start + len(token), entryID: id})
			} else if id, ok := index.simple[token]; ok {
				links = append(links, classLink{start: start, end: start + len(token), entryID: id})
			}
		default:
			at++
		}
	}
	return links
}

func underlineClassLinks(canvas *vivaui.Canvas, filePath string, index *classLinkIndex) {
	for _, text := range canvas.Texts() {
		links := buildClassLinks(text.Value, filePath, index)
		if len(links) == 0 {
			continue
		}
		runs := canvas.TextRuns(text)
		offset := 0
		for index := range runs {
			start := offset
			offset += int(runs[index].Bytes)
			for _, link := range links {
				if start < link.end && offset > link.start {
					runs[index].Underline = true
					break
				}
			}
		}
	}
}

func classLinkAt(links []classLink, offset int) (int, bool) {
	for _, link := range links {
		if offset >= link.start && offset <= link.end {
			return link.entryID, true
		}
	}
	return -1, false
}

func (m *Model) sourceLinkPointer(event vivaui.PointerEvent) (vivaui.EventResult, bool) {
	doc := m.currentDocument()
	if doc == nil {
		m.linkDownOffset = -1
		return vivaui.EventResult{}, false
	}
	switch event.Kind {
	case vivaui.PointerDown:
		m.linkDownOffset = -1
		if offset, ok := m.code.OffsetAt(vivaui.Point{X: event.X, Y: event.Y}); ok {
			if _, linked := classLinkAt(doc.links, offset); linked {
				m.linkDownOffset = offset
			}
		}
	case vivaui.PointerUp:
		down := m.linkDownOffset
		m.linkDownOffset = -1
		if down < 0 {
			return vivaui.EventResult{}, false
		}
		offset, ok := m.code.OffsetAt(vivaui.Point{X: event.X, Y: event.Y})
		if !ok || offset != down {
			return vivaui.EventResult{}, false
		}
		if entryID, linked := classLinkAt(doc.links, offset); linked {
			m.startDecompile(entryID)
			return vivaui.EventResult{Handled: true, Repaint: true}, true
		}
	}
	return vivaui.EventResult{}, false
}
