package vineflower

import (
	"context"
	"errors"
	"testing"

	"github.com/xddcode/jarlens/internal/decompile"
)

func TestEngineIdentity(t *testing.T) {
	engine := New()
	defer engine.Close()
	if engine.Name() != "Vineflower" || engine.Version() != "1.12.0" {
		t.Fatalf("unexpected engine identity: %s %s", engine.Name(), engine.Version())
	}
}

func TestEngineWithoutJava(t *testing.T) {
	engine := &Engine{}
	defer engine.Close()
	_, err := engine.Decompile(context.Background(), decompile.Request{ClassName: "Empty.class"})
	if !errors.Is(err, decompile.ErrJavaNotFound) {
		t.Fatalf("missing Java error is %v", err)
	}
}

func TestLimitedBuffer(t *testing.T) {
	buffer := &limitedBuffer{}
	input := make([]byte, maxLogBytes+100)
	if _, err := buffer.Write(input); err != nil {
		t.Fatal(err)
	}
	if len(buffer.data) != maxLogBytes {
		t.Fatalf("log size is %d", len(buffer.data))
	}
}
