package javaengine

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/xddcode/jarlens/internal/decompile"
)

func TestEngineRejectsArtifactChecksumMismatch(t *testing.T) {
	engine := New(Config{Name: "Broken", Version: "1", ArtifactName: "broken.jar", Artifact: []byte("broken"), SHA256: "0000000000000000000000000000000000000000000000000000000000000000"})
	defer engine.Close()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	engine.SetJava(executable)
	_, err = engine.Decompile(context.Background(), decompile.Request{ClassName: "Input.class", Data: []byte("input")})
	if err == nil || !strings.Contains(err.Error(), "artifact checksum mismatch") {
		t.Fatalf("checksum mismatch was not rejected: %v", err)
	}
}
