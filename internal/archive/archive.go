package archive

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	MaxEntries    = 200000
	MaxEntryBytes = 16 << 20
)

var (
	ErrUnsupportedFormat = errors.New("only JAR and Class files are supported")
	ErrInvalidClass      = errors.New("invalid Java class file")
	ErrEntryTooLarge     = errors.New("archive entry exceeds the 16 MiB read limit")
	ErrTooManyEntries    = errors.New("archive exceeds the 200000 entry limit")
	ErrEntryNotFound     = errors.New("archive entry not found")
)

type EntryKind uint8

const (
	EntryClass EntryKind = iota
	EntryResource
)

type Entry struct {
	ID   int
	Path string
	Name string
	Kind EntryKind
	Size int64
}

type Archive struct {
	Path       string
	Entries    []Entry
	Warnings   []string
	zipIndexes []int
	standalone bool
}

func Open(filePath string) (*Archive, error) {
	extension := strings.ToLower(filepath.Ext(filePath))
	if extension != ".jar" && extension != ".class" {
		return nil, ErrUnsupportedFormat
	}
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", filePath)
	}

	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".class":
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		var magic [4]byte
		if _, err := io.ReadFull(file, magic[:]); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidClass, err)
		}
		if magic != [4]byte{0xca, 0xfe, 0xba, 0xbe} {
			return nil, ErrInvalidClass
		}
		entry := Entry{ID: 0, Path: filepath.Base(filePath), Name: filepath.Base(filePath), Kind: EntryClass, Size: info.Size()}
		return &Archive{Path: filePath, Entries: []Entry{entry}, standalone: true}, nil
	case ".jar":
		return openJAR(filePath)
	default:
		return nil, ErrUnsupportedFormat
	}
}

func openJAR(filePath string) (*Archive, error) {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	result := &Archive{Path: filePath}
	for index, entry := range reader.File {
		if entry.FileInfo().IsDir() {
			continue
		}
		if len(result.Entries) >= MaxEntries {
			return nil, ErrTooManyEntries
		}
		name := entry.Name
		if !safeEntryPath(name) {
			result.Warnings = append(result.Warnings, "Ignored unsafe archive entry: "+name)
			continue
		}
		kind := EntryResource
		if strings.EqualFold(filepath.Ext(name), ".class") {
			kind = EntryClass
		}
		id := len(result.Entries)
		result.Entries = append(result.Entries, Entry{
			ID: id, Path: name, Name: entry.FileInfo().Name(), Kind: kind,
			Size: int64(entry.UncompressedSize64),
		})
		result.zipIndexes = append(result.zipIndexes, index)
	}
	return result, nil
}

func (a *Archive) Read(id int) ([]byte, error) {
	if id < 0 || id >= len(a.Entries) {
		return nil, ErrEntryNotFound
	}
	entry := a.Entries[id]
	if entry.Size > MaxEntryBytes {
		return nil, ErrEntryTooLarge
	}
	if a.standalone {
		file, err := os.Open(a.Path)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		return readEntry(file, entry)
	}

	reader, err := zip.OpenReader(a.Path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	if id >= len(a.zipIndexes) || a.zipIndexes[id] >= len(reader.File) {
		return nil, ErrEntryNotFound
	}
	entryReader, err := reader.File[a.zipIndexes[id]].Open()
	if err != nil {
		return nil, err
	}
	defer entryReader.Close()
	return readEntry(entryReader, entry)
}

func readEntry(reader io.Reader, entry Entry) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, MaxEntryBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > MaxEntryBytes {
		return nil, ErrEntryTooLarge
	}
	if entry.Kind == EntryClass && (len(data) < 4 || string(data[:4]) != "\xca\xfe\xba\xbe") {
		return nil, ErrInvalidClass
	}
	return data, nil
}

func safeEntryPath(name string) bool {
	if name == "" || strings.HasPrefix(name, "/") || strings.ContainsRune(name, '\x00') || strings.Contains(name, "\\") {
		return false
	}
	parts := strings.Split(name, "/")
	if len(parts) == 0 || strings.Contains(parts[0], ":") {
		return false
	}
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}
