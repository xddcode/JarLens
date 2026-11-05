package workspace

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestTextResource(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fixture.jar")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create("config/application.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("<configuration><enabled>true</enabled></configuration>")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	workspace, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	text, err := workspace.TextResource(0)
	if err != nil {
		t.Fatal(err)
	}
	if text != "<configuration><enabled>true</enabled></configuration>" {
		t.Fatalf("unexpected XML: %q", text)
	}
}

func TestDecodeText(t *testing.T) {
	for _, test := range []struct {
		name string
		data []byte
		want string
	}{
		{name: "utf8", data: []byte("hello 世界"), want: "hello 世界"},
		{name: "utf8 bom", data: append([]byte{0xef, 0xbb, 0xbf}, []byte("hello")...), want: "hello"},
		{name: "utf16le", data: []byte{0xff, 0xfe, 'h', 0, 'i', 0}, want: "hi"},
		{name: "utf16be", data: []byte{0xfe, 0xff, 0, 'h', 0, 'i'}, want: "hi"},
		{name: "empty", data: nil, want: ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := decodeText(test.data)
			if err != nil || got != test.want {
				t.Fatalf("decodeText() = %q, %v; want %q", got, err, test.want)
			}
		})
	}
}

func TestDecodeTextRejectsBinaryAndMalformedData(t *testing.T) {
	for _, data := range [][]byte{{0, 1, 2}, {0xff}, {0xff, 0xfe, 0}, {'a', 1, 'b'}} {
		if _, err := decodeText(data); !errors.Is(err, ErrNotText) {
			t.Fatalf("binary data %v was accepted: %v", data, err)
		}
	}
}
