//go:build integration

package vineflower

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xddcode/jarlens/internal/decompile"
)

func TestVineflowerDecompilesJavaClass(t *testing.T) {
	if _, err := exec.LookPath("javac"); err != nil {
		t.Skip("javac is required for the engine integration fixture")
	}
	if _, err := exec.LookPath("java"); err != nil {
		t.Skip("java is required for the engine integration fixture")
	}

	workspace := t.TempDir()
	source := []byte("package example;\npublic class Hello {\n  public String greeting(String name) {\n    return \"Hello, \" + name;\n  }\n}\n")
	sourcePath := filepath.Join(workspace, "Hello.java")
	if err := os.WriteFile(sourcePath, source, 0o600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command("javac", "-parameters", "-d", workspace, sourcePath).CombinedOutput(); err != nil {
		t.Fatalf("javac failed: %v\n%s", err, output)
	}
	classData, err := os.ReadFile(filepath.Join(workspace, "example", "Hello.class"))
	if err != nil {
		t.Fatal(err)
	}

	engine := New()
	defer engine.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	request := decompile.Request{ClassName: "example/Hello.class", Data: classData}
	result, err := engine.Decompile(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"package example;", "class Hello", "String greeting(String name)", "Hello, "} {
		if !strings.Contains(result.Source, want) {
			t.Errorf("decompiled source does not contain %q:\n%s", want, result.Source)
		}
	}
	if result.Engine != "Vineflower" || result.EngineVersion != engineVersion {
		t.Fatal("result lost engine provenance")
	}

	cancelled, cancelNow := context.WithCancel(context.Background())
	cancelNow()
	if _, err := engine.Decompile(cancelled, request); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled engine request returned %v", err)
	}
}
