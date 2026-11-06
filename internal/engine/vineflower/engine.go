package vineflower

import (
	"context"
	"embed"
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
	engineVersion  = "1.12.0"
	maxLogBytes    = 64 << 10
	maxSourceBytes = 32 << 20
)

//go:embed assets/vineflower-1.12.0.jar
var assets embed.FS

type Engine struct {
	java     string
	once     sync.Once
	jarPath  string
	cacheDir string
	prepErr  error
}

func New() *Engine {
	return &Engine{java: findJava("")}
}

func (e *Engine) Name() string {
	return "Vineflower"
}

func (e *Engine) Version() string {
	return engineVersion
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
	archiveInput := request.ArchivePath != "" && strings.EqualFold(filepath.Ext(request.ArchivePath), ".jar")
	args := []string{"-jar", e.jarPath, "--silent", "--folder", "--thread-count=1", "--include-runtime=current"}
	if archiveInput {
		input = request.ArchivePath
		if err := os.MkdirAll(filepath.Join(outputDir, filepath.Dir(request.ClassName)), 0o700); err != nil {
			return decompile.Result{}, err
		}
		args = append(args, "--only="+strings.TrimSuffix(request.ClassName, filepath.Ext(request.ClassName)))
	} else {
		if err := os.WriteFile(input, request.Data, 0o600); err != nil {
			return decompile.Result{}, err
		}
		if request.ArchivePath != "" {
			args = append(args, "--add-external="+request.ArchivePath)
		}
	}
	args = append(args, input, outputDir)
	command := exec.CommandContext(ctx, e.java, args...)
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
		return decompile.Result{}, fmt.Errorf("%w: %s", decompile.ErrEngineFailed, message)
	}

	output, err := findSource(outputDir, base+".java")
	if err != nil {
		return decompile.Result{}, fmt.Errorf("%w: %v", decompile.ErrEngineFailed, err)
	}
	info, err := os.Stat(output)
	if err != nil {
		return decompile.Result{}, fmt.Errorf("%w: %v", decompile.ErrEngineFailed, err)
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

func (e *Engine) prepare() error {
	e.once.Do(func() {
		if e.java == "" {
			e.prepErr = decompile.ErrJavaNotFound
			return
		}
		e.cacheDir, e.prepErr = os.MkdirTemp("", "jarlens-vineflower-*")
		if e.prepErr != nil {
			return
		}
		data, err := assets.ReadFile("assets/vineflower-" + engineVersion + ".jar")
		if err != nil {
			e.prepErr = err
			return
		}
		e.jarPath = filepath.Join(e.cacheDir, "vineflower-"+engineVersion+".jar")
		e.prepErr = os.WriteFile(e.jarPath, data, 0o600)
	})
	return e.prepErr
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
