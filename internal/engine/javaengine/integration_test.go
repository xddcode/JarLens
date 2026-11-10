//go:build integration

package javaengine_test

import (
	"archive/zip"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xddcode/jarlens/internal/decompile"
	"github.com/xddcode/jarlens/internal/engine/cfr"
	"github.com/xddcode/jarlens/internal/engine/procyon"
)

func TestCFRAndProcyonDecompileClassAndJAR(t *testing.T) {
	if _, err := exec.LookPath("javac"); err != nil {
		t.Skip("javac is required for the engine integration fixture")
	}
	directory := t.TempDir()
	source := filepath.Join(directory, "Hello.java")
	if err := os.WriteFile(source, []byte("package example;\npublic class Hello { public String greet(String name) { return \"Hello \" + name; } }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	classes := filepath.Join(directory, "classes")
	if err := os.MkdirAll(classes, 0o700); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command("javac", "-d", classes, source).CombinedOutput(); err != nil {
		t.Fatalf("javac: %v: %s", err, output)
	}
	classPath := filepath.Join(classes, "example", "Hello.class")
	data, err := os.ReadFile(classPath)
	if err != nil {
		t.Fatal(err)
	}
	archivePath := filepath.Join(directory, "hello.jar")
	archive, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(archive)
	entry, err := writer.Create("example/Hello.class")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name    string
		version string
		engine  decompile.Engine
	}{
		{name: "CFR", version: "0.152", engine: cfr.New()},
		{name: "Procyon", version: "0.6.0", engine: procyon.New()},
	} {
		t.Run(test.name, func(t *testing.T) {
			defer test.engine.(interface{ Close() error }).Close()
			if test.engine.Name() != test.name || test.engine.Version() != test.version {
				t.Fatalf("unexpected engine identity: %s %s", test.engine.Name(), test.engine.Version())
			}
			for _, useArchive := range []bool{false, true} {
				t.Run(map[bool]string{false: "class", true: "jar"}[useArchive], func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					request := decompile.Request{ClassName: "example/Hello.class", Data: data}
					if useArchive {
						request.ArchivePath = archivePath
					}
					result, err := test.engine.Decompile(ctx, request)
					if err != nil {
						t.Fatal(err)
					}
					if result.Engine != test.name || result.EngineVersion != test.version {
						t.Fatalf("unexpected result identity: %+v", result)
					}
					for _, want := range []string{"package example;", "public class Hello", "greet"} {
						if !strings.Contains(result.Source, want) {
							t.Errorf("%s output does not contain %q:\n%s", test.name, want, result.Source)
						}
					}
				})
			}
		})
	}
}
