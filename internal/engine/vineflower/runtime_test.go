package vineflower

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFindJavaPrefersExplicitOverride(t *testing.T) {
	t.Setenv(javaOverrideEnv, filepath.Join("custom", "java"))
	if got := findJava("ignored"); got != filepath.Join("custom", "java") {
		t.Fatalf("override is %q", got)
	}
}

func TestFindJavaUsesBundledRuntime(t *testing.T) {
	t.Setenv(javaOverrideEnv, "")
	directory := t.TempDir()
	executable := filepath.Join(directory, "jarlens")
	java := filepath.Join(directory, "jre", "bin", javaName())
	if err := os.MkdirAll(filepath.Dir(java), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(java, []byte("runtime"), 0o700); err != nil {
		t.Fatal(err)
	}
	if got := findJava(executable); got != java {
		t.Fatalf("bundled Java is %q, want %q", got, java)
	}
}

func TestJavaCandidatesIncludeMacOSBundleResources(t *testing.T) {
	executable := filepath.Join("JarLens.app", "Contents", "MacOS", "jarlens")
	candidates := javaCandidates(executable)
	if len(candidates) != 2 || candidates[1] != filepath.Join("JarLens.app", "Contents", "Resources", "jre", "bin", javaName()) {
		t.Fatalf("unexpected candidates: %+v", candidates)
	}
}

func javaName() string {
	if runtime.GOOS == "windows" {
		return "java.exe"
	}
	return "java"
}
