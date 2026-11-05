package workspace

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/xddcode/jarlens/internal/archive"
)

var ErrNotText = errors.New("resource is not a supported text document")

func (w *Workspace) TextResource(entryID int) (string, error) {
	if w == nil || w.Archive == nil || entryID < 0 || entryID >= len(w.Archive.Entries) {
		return "", archive.ErrEntryNotFound
	}
	entry := w.Archive.Entries[entryID]
	if entry.Kind != archive.EntryResource {
		return "", fmt.Errorf("%w: %s", ErrNotText, entry.Path)
	}
	data, err := w.Archive.Read(entryID)
	if err != nil {
		return "", err
	}
	text, err := decodeText(data)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrNotText, entry.Path)
	}
	return text, nil
}

func decodeText(data []byte) (string, error) {
	var text string
	switch {
	case len(data) >= 3 && string(data[:3]) == "\xef\xbb\xbf":
		text = string(data[3:])
	case len(data) >= 2 && data[0] == 0xff && data[1] == 0xfe:
		if len(data)%2 != 0 {
			return "", ErrNotText
		}
		text = decodeUTF16(data[2:], true)
	case len(data) >= 2 && data[0] == 0xfe && data[1] == 0xff:
		if len(data)%2 != 0 {
			return "", ErrNotText
		}
		text = decodeUTF16(data[2:], false)
	case utf8.Valid(data):
		text = string(data)
	default:
		return "", ErrNotText
	}
	if !utf8.ValidString(text) || strings.ContainsRune(text, '\x00') {
		return "", ErrNotText
	}
	for _, r := range text {
		if r < 0x20 && r != '\n' && r != '\r' && r != '\t' {
			return "", ErrNotText
		}
	}
	return text, nil
}

func decodeUTF16(data []byte, littleEndian bool) string {
	units := make([]uint16, 0, len(data)/2)
	for index := 0; index+1 < len(data); index += 2 {
		if littleEndian {
			units = append(units, uint16(data[index])|uint16(data[index+1])<<8)
		} else {
			units = append(units, uint16(data[index])<<8|uint16(data[index+1]))
		}
	}
	return string(utf16.Decode(units))
}
