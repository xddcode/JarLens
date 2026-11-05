package archive

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

var classBytes = []byte{0xca, 0xfe, 0xba, 0xbe, 0, 0, 0, 65}

func TestOpenJARAndReadEntries(t *testing.T) {
	path := writeJAR(t, map[string][]byte{
		"com/example/Hello.class": classBytes,
		"META-INF/MANIFEST.MF":    []byte("Manifest-Version: 1.0\n"),
	})
	archive, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(archive.Entries) != 2 {
		t.Fatalf("got %d entries", len(archive.Entries))
	}
	entryID := -1
	for _, entry := range archive.Entries {
		if entry.Kind == EntryClass {
			entryID = entry.ID
			break
		}
	}
	data, err := archive.Read(entryID)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, classBytes) {
		t.Fatal("class bytes changed while reading")
	}
	if _, err := archive.Read(-1); !errors.Is(err, ErrEntryNotFound) {
		t.Fatalf("invalid entry error is %v", err)
	}
}

func TestOpenJARSkipsUnsafeEntries(t *testing.T) {
	path := writeJAR(t, map[string][]byte{
		"../escape.class": classBytes,
		"safe/Ok.class":   classBytes,
	})
	archive, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(archive.Entries) != 1 || archive.Entries[0].Path != "safe/Ok.class" {
		t.Fatalf("unsafe entries were not ignored: %+v", archive.Entries)
	}
	if len(archive.Warnings) != 1 {
		t.Fatal("unsafe entry did not produce a diagnostic")
	}
}

func TestOpenStandaloneClass(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Hello.class")
	if err := os.WriteFile(path, classBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	archive, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(archive.Entries) != 1 || !archive.standalone {
		t.Fatal("standalone class was not recognized")
	}
}

func TestOpenRejectsInvalidInputs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.class")
	if err := os.WriteFile(path, []byte("not a class"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(path); !errors.Is(err, ErrInvalidClass) {
		t.Fatalf("invalid class error is %v", err)
	}
	if _, err := Open("file.txt"); !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("unsupported format error is %v", err)
	}
}

func writeJAR(t *testing.T, entries map[string][]byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fixture.jar")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for name, data := range entries {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}
