package javaengine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/xddcode/jarlens/internal/decompile"
)

const (
	maxLogBytes    = 64 << 10
	maxSourceBytes = 32 << 20
)

type Config struct {
	Name         string
	Version      string
	ArtifactName string
	SHA256       string
	Artifact     []byte
	UseArchive   bool
	Arguments    func(decompile.Request, string, string) []string
}

type Engine struct {
	config   Config
	javaMu   sync.RWMutex
	java     string
	once     sync.Once
	jarPath  string
	cacheDir string
	prepErr  error
}

func New(config Config) *Engine {
	return &Engine{config: config, java: findJava("")}
}

func (e *Engine) Name() string {
	return e.config.Name
}

func (e *Engine) Version() string {
	return e.config.Version
}

func (e *Engine) SetJava(path string) {
	if path == "" {
		path = findJava("")
	}
	e.javaMu.Lock()
	e.java = path
	e.javaMu.Unlock()
}

func (e *Engine) javaPath() string {
	e.javaMu.RLock()
	defer e.javaMu.RUnlock()
	return e.java
}

func (e *Engine) prepare() error {
	e.once.Do(func() {
		if e.javaPath() == "" {
			e.prepErr = decompile.ErrJavaNotFound
			return
		}
		digest := sha256.Sum256(e.config.Artifact)
		if hex.EncodeToString(digest[:]) != e.config.SHA256 {
			e.prepErr = fmt.Errorf("%w: %s artifact checksum mismatch", decompile.ErrEngineFailed, e.config.Name)
			return
		}
		e.cacheDir, e.prepErr = os.MkdirTemp("", "jarlens-"+strings.ToLower(e.config.Name)+"-*")
		if e.prepErr != nil {
			return
		}
		e.jarPath = filepath.Join(e.cacheDir, e.config.ArtifactName)
		e.prepErr = os.WriteFile(e.jarPath, e.config.Artifact, 0o600)
	})
	return e.prepErr
}

func (e *Engine) Decompile(ctx context.Context, request decompile.Request) (decompile.Result, error) {
	started := time.Now()
	if err := e.prepare(); err != nil {
		return decompile.Result{}, err
	}
	workspace, err := os.MkdirTemp("", "jarlens-decompile-*")
	if err != nil {
		return decompile.Result{}, err
	}
	defer os.RemoveAll(workspace)
	inputDir := filepath.Join(workspace, "input")
	outputDir := filepath.Join(workspace, "output")
	if err := os.MkdirAll(inputDir, 0o700); err != nil {
		return decompile.Result{}, err
	}
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		return decompile.Result{}, err
	}
	base := strings.TrimSuffix(filepath.Base(request.ClassName), ".class")
	if base == "" || base == "." {
		base = "Input"
	}
	input := filepath.Join(inputDir, base+".class")
	if e.config.UseArchive && request.ArchivePath != "" && strings.EqualFold(filepath.Ext(request.ArchivePath), ".jar") {
		input = request.ArchivePath
	} else if err := os.WriteFile(input, request.Data, 0o600); err != nil {
		return decompile.Result{}, err
	}
	args := []string{"-jar", e.jarPath}
	args = append(args, e.config.Arguments(request, input, outputDir)...)
	command := exec.CommandContext(ctx, e.javaPath(), args...)
	command.WaitDelay = 5 * time.Second
	var logs limitedBuffer
	command.Stdout = &logs
	command.Stderr = &logs
	if err := command.Run(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return decompile.Result{}, ctxErr
		}
		message := strings.TrimSpace(logs.String())
		if message == "" {
			message = err.Error()
		}
		return decompile.Result{}, fmt.Errorf("%w: %s: %s", decompile.ErrEngineFailed, e.config.Name, message)
	}
	output, err := findSource(outputDir, base+".java")
	if err != nil {
		return decompile.Result{}, fmt.Errorf("%w: %s: %v", decompile.ErrEngineFailed, e.config.Name, err)
	}
	info, err := os.Stat(output)
	if err != nil {
		return decompile.Result{}, fmt.Errorf("%w: %s: %v", decompile.ErrEngineFailed, e.config.Name, err)
	}
	if info.Size() > maxSourceBytes {
		return decompile.Result{}, decompile.ErrOutputTooLarge
	}
	data, err := os.ReadFile(output)
	if err != nil {
		return decompile.Result{}, err
	}
	result := decompile.Result{
		Source: string(data), Engine: e.Name(), EngineVersion: e.Version(), Duration: time.Since(started),
	}
	if message := strings.TrimSpace(logs.String()); message != "" {
		result.Diagnostics = []string{message}
	}
	return result, nil
}

func (e *Engine) Close() error {
	if e.cacheDir == "" {
		return nil
	}
	return os.RemoveAll(e.cacheDir)
}

func findSource(root, name string) (string, error) {
	var matches []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && entry.Name() == name {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("expected one %s output, found %d", name, len(matches))
	}
	return matches[0], nil
}

type limitedBuffer struct {
	data []byte
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	remaining := maxLogBytes - len(b.data)
	if remaining > 0 {
		b.data = append(b.data, p[:min(len(p), remaining)]...)
	}
	return len(p), nil
}

func (b *limitedBuffer) String() string {
	return string(b.data)
}

var _ decompile.Engine = (*Engine)(nil)
var _ io.Writer = (*limitedBuffer)(nil)
