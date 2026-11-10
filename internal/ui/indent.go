package ui

import (
	"strings"

	"github.com/vivaui/viva/components"
	"github.com/vivaui/viva/text"
)

type javaIndentState struct {
	blockComment bool
	textBlock    bool
}

func sourceIndentWidth(language string) int {
	if language == "java" {
		return 4
	}
	return 2
}

func reindentSource(source, language string, from, to int) string {
	if source == "" || from <= 0 || to <= 0 {
		return source
	}
	if language == "java" {
		return reindentJava(source, to)
	}
	if from == to {
		return source
	}
	lines := strings.Split(source, "\n")
	for index, line := range lines {
		content := 0
		columns := 0
		for content < len(line) {
			switch line[content] {
			case ' ':
				columns++
			case '\t':
				columns += from
			default:
				goto leadingDone
			}
			content++
		}
	leadingDone:
		indent := columns * to / from
		if content == len(line) {
			lines[index] = strings.Repeat(" ", indent)
			continue
		}
		lines[index] = strings.Repeat(" ", indent) + line[content:]
	}
	return strings.Join(lines, "\n")
}

func reindentJava(source string, width int) string {
	lines := strings.Split(source, "\n")
	depth := 0
	state := javaIndentState{}
	for index, line := range lines {
		ending := ""
		content := strings.TrimSuffix(line, "\r")
		if len(content) != len(line) {
			ending = "\r"
		}
		trimmed := strings.TrimLeft(content, " \t")
		if trimmed == "" {
			lines[index] = ending
			continue
		}
		if state.textBlock {
			if strings.Contains(trimmed, `"""`) {
				lines[index] = strings.Repeat(" ", depth*width) + trimmed + ending
				state.textBlock = false
			}
			continue
		}
		leadingClosers := 0
		if !state.blockComment {
			for leadingClosers < len(trimmed) && trimmed[leadingClosers] == '}' {
				leadingClosers++
			}
		}
		lines[index] = strings.Repeat(" ", max(depth-leadingClosers, 0)*width) + trimmed + ending
		depth = max(depth+javaBraceDelta(content, &state), 0)
	}
	return strings.Join(lines, "\n")
}

func javaBraceDelta(line string, state *javaIndentState) int {
	delta := 0
	for index := 0; index < len(line); {
		if state.blockComment {
			end := strings.Index(line[index:], "*/")
			if end < 0 {
				return delta
			}
			index += end + 2
			state.blockComment = false
			continue
		}
		if strings.HasPrefix(line[index:], "//") {
			return delta
		}
		if strings.HasPrefix(line[index:], "/*") {
			state.blockComment = true
			index += 2
			continue
		}
		if strings.HasPrefix(line[index:], `"""`) {
			if !strings.Contains(line[index+3:], `"""`) {
				state.textBlock = true
			}
			return delta
		}
		if line[index] == '"' || line[index] == '\'' {
			quote := line[index]
			index++
			for index < len(line) {
				if line[index] == '\\' {
					index += 2
					continue
				}
				current := line[index]
				index++
				if current == quote {
					break
				}
			}
			continue
		}
		switch line[index] {
		case '{':
			delta++
		case '}':
			delta--
		}
		index++
	}
	return delta
}

func (m *Model) reindentDocuments(from, to int) {
	if from == to {
		return
	}
	m.walkPanes(func(pane *editorPane) {
		for index, doc := range pane.documents {
			if doc.source == "" {
				continue
			}
			doc.source = reindentSource(doc.source, doc.language, from, to)
			doc.editor = text.NewEditor(doc.source)
			doc.code = components.CodeEditor{}
			doc.links = buildClassLinks(doc.source, doc.path, m.classLinks)
			doc.pendingCenter = false
			if index == pane.active {
				pane.editor = doc.editor
				pane.code = doc.code
			}
		}
	})
	if doc := m.currentDocument(); doc != nil {
		m.source = doc.source
		m.editor = doc.editor
		m.code = doc.code
		m.syncActivePane()
	}
	m.findSource = ""
}
