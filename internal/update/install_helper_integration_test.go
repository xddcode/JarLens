//go:build windows && integration

package update

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

const windowsUpdateFixture = `package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	vivaupdate "github.com/vivaui/viva/update"
)

func main() {
	if strings.EqualFold(filepath.Base(os.Args[0]), "setup.exe") {
		if err := os.WriteFile(filepath.Join(os.Getenv("JARLENS_UPDATE_FIXTURE_DIR"), "setup.ok"), []byte("installed"), 0o600); err != nil {
			os.Exit(1)
		}
		return
	}
	mode := flag.String("mode", "install", "fixture mode")
	setup := flag.String("setup", "", "installer path")
	flag.Parse()
	if *mode == "restarted" {
		if err := os.WriteFile(filepath.Join(os.Getenv("JARLENS_UPDATE_FIXTURE_DIR"), "restarted.ok"), []byte("restarted"), 0o600); err != nil {
			os.Exit(1)
		}
		return
	}
	body, err := os.ReadFile(*setup)
	if err != nil {
		os.Exit(1)
	}
	digest := sha256.Sum256(body)
	err = vivaupdate.Install(vivaupdate.InstallOptions{
		Artifact: *setup,
		Asset: vivaupdate.Asset{OS: runtime.GOOS, Arch: runtime.GOARCH, Format: vivaupdate.FormatNSIS, Size: int64(len(body)), SHA256: hex.EncodeToString(digest[:])},
		Version: "9.9.9", Args: []string{"-mode", "restarted"},
	}, func() { os.Exit(0) })
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
`

const windowsPowerShellLogger = `package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	directory := os.Getenv("JARLENS_UPDATE_FIXTURE_DIR")
	arguments, _ := os.OpenFile(filepath.Join(directory, "powershell.args"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	_, _ = arguments.WriteString(strings.Join(os.Args, "\n"))
	_ = arguments.Close()
	output, _ := os.OpenFile(filepath.Join(directory, "powershell.log"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	command := exec.Command("C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe", os.Args[1:]...)
	command.Stdout = output
	command.Stderr = output
	err := command.Run()
	_ = output.Close()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		os.Exit(exit.ExitCode())
	}
	if err != nil {
		os.Exit(1)
	}
}
`

func TestWindowsUpdateHelperInstallsAndRestarts(t *testing.T) {
	directory, err := filepath.Abs(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(directory, "fixture.go")
	if err := os.WriteFile(source, []byte(windowsUpdateFixture), 0o600); err != nil {
		t.Fatal(err)
	}
	application := filepath.Join(directory, "app.exe")
	setup := filepath.Join(directory, "setup.exe")
	for _, output := range []string{application, setup} {
		command := exec.Command("go", "build", "-ldflags", "-H=windowsgui", "-o", output, source)
		if data, err := command.CombinedOutput(); err != nil {
			t.Fatalf("build fixture: %v\n%s", err, data)
		}
	}
	wrapper := filepath.Join(directory, "wrapper")
	if err := os.Mkdir(wrapper, 0o700); err != nil {
		t.Fatal(err)
	}
	logger := filepath.Join(wrapper, "logger.go")
	if err := os.WriteFile(logger, []byte(windowsPowerShellLogger), 0o600); err != nil {
		t.Fatal(err)
	}
	buildLogger := exec.Command("go", "build", "-o", filepath.Join(wrapper, "powershell.exe"), logger)
	if output, err := buildLogger.CombinedOutput(); err != nil {
		t.Fatalf("build PowerShell logger: %v\n%s", err, output)
	}
	command := exec.Command(application, "-setup", setup)
	command.Env = append(os.Environ(), "JARLENS_UPDATE_FIXTURE_DIR="+directory, "PATH="+wrapper+";"+os.Getenv("PATH"))
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("update process: %v\n%s", err, output)
	}
	for _, name := range []string{"setup.ok", "restarted.ok"} {
		path := filepath.Join(directory, name)
		deadline := time.Now().Add(20 * time.Second)
		for {
			if _, err := os.Stat(path); err == nil {
				break
			} else if !os.IsNotExist(err) {
				t.Fatal(err)
			}
			if time.Now().After(deadline) {
				for _, logName := range []string{"powershell.args", "powershell.log"} {
					if data, err := os.ReadFile(filepath.Join(directory, logName)); err == nil {
						t.Logf("%s:\n%s", logName, data)
					}
				}
				t.Fatalf("update helper did not create %s", name)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
}
