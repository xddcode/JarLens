package workspace

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/xddcode/jarlens/internal/archive"
)

func TestWorkspaceCreatesClassRequest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fixture.jar")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create("example/Hello.class")
	if err != nil {
		t.Fatal(err)
	}
	class := []byte{0xca, 0xfe, 0xba, 0xbe, 0, 0, 0, 65}
	if _, err := entry.Write(class); err != nil {
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
	request, err := workspace.DecompileRequest(0)
	if err != nil {
		t.Fatal(err)
	}
	if request.ClassName != "example/Hello.class" || request.ArchivePath != path || len(request.Data) != len(class) {
		t.Fatalf("unexpected request: %+v", request)
	}
}

func TestWorkspaceRejectsMissingEntry(t *testing.T) {
	workspace := &Workspace{Archive: &archive.Archive{}}
	if _, err := workspace.DecompileRequest(0); err == nil {
		t.Fatal("missing entry was accepted")
	}
}
