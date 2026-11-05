package workspace

import (
	"fmt"

	"github.com/xddcode/jarlens/internal/archive"
	"github.com/xddcode/jarlens/internal/decompile"
)

type Workspace struct {
	Archive *archive.Archive
	Tree    *archive.Tree
}

func Open(path string) (*Workspace, error) {
	opened, err := archive.Open(path)
	if err != nil {
		return nil, err
	}
	return &Workspace{Archive: opened, Tree: archive.BuildTree(opened.Entries)}, nil
}

func (w *Workspace) DecompileRequest(entryID int) (decompile.Request, error) {
	if w == nil || w.Archive == nil || entryID < 0 || entryID >= len(w.Archive.Entries) {
		return decompile.Request{}, archive.ErrEntryNotFound
	}
	entry := w.Archive.Entries[entryID]
	if entry.Kind != archive.EntryClass {
		return decompile.Request{}, fmt.Errorf("%s is not a Java class", entry.Path)
	}
	data, err := w.Archive.Read(entryID)
	if err != nil {
		return decompile.Request{}, err
	}
	return decompile.Request{ClassName: entry.Path, Data: data, ArchivePath: w.Archive.Path}, nil
}
