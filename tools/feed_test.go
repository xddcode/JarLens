package tools

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	vivaupdate "github.com/vivaui/viva/update"
)

func TestReleaseFeedIncludesInstallablePackages(t *testing.T) {
	directory := t.TempDir()
	files := map[string]string{
		"jarlens-9.9.9-windows-amd64-setup.exe":     "setup",
		"jarlens-9.9.9-windows-amd64.zip":           "standard windows",
		"jarlens-9.9.9-windows-amd64-portable.zip":  "portable windows",
		"jarlens-9.9.9-darwin-arm64.zip":            "mac",
		"jarlens-9.9.9-linux-amd64.tar.gz":          "linux",
		"jarlens-9.9.9-linux-amd64-portable.tar.gz": "portable linux",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(directory, "latest.json")
	command := exec.Command("go", "tool", "viva", "feed", "-C", "..",
		"-dir", directory, "-out", manifest,
		"-app", "io.github.xddcode.jarlens", "-version", "9.9.9", "-channel", "stable",
		"-base-url", "https://example.com/releases/latest/download",
		"-notes", "https://example.com/releases/v9.9.9",
	)
	command.Env = append(os.Environ(), "VIVA_UPDATE_KEY="+base64.StdEncoding.EncodeToString(private))
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("viva feed: %v\n%s", err, output)
	}
	body, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	signatureText, err := os.ReadFile(manifest + ".sig")
	if err != nil {
		t.Fatal(err)
	}
	signature, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(signatureText)))
	if err != nil || !ed25519.Verify(public, body, signature) {
		t.Fatalf("feed signature does not verify: %v", err)
	}
	var feed vivaupdate.Feed
	if err := json.Unmarshal(body, &feed); err != nil {
		t.Fatal(err)
	}
	if len(feed.Latest.Assets) != 4 {
		t.Fatalf("feed has %d assets, want 4: %+v", len(feed.Latest.Assets), feed.Latest.Assets)
	}
	asset, err := feed.Asset("windows", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if asset.Format != vivaupdate.FormatNSIS || asset.Size != int64(len("setup")) || !strings.HasSuffix(asset.URL, "jarlens-9.9.9-windows-amd64-setup.exe") {
		t.Fatalf("Windows update asset = %+v", asset)
	}
	for _, asset := range feed.Latest.Assets {
		if strings.Contains(asset.URL, "portable") {
			t.Fatalf("duplicate portable archive entered the feed: %+v", asset)
		}
	}

	cnbManifest := filepath.Join(directory, "cnb-latest.json")
	command = exec.Command("go", "tool", "viva", "feed", "-C", "..",
		"-dir", directory, "-out", cnbManifest,
		"-app", "io.github.xddcode.jarlens", "-version", "9.9.9", "-channel", "stable",
		"-base-url", "https://cnb.cool/xddcode/JarLens/-/releases/download/v9.9.9",
		"-notes", "https://cnb.cool/xddcode/JarLens/-/releases/tag/v9.9.9",
	)
	command.Env = append(os.Environ(), "VIVA_UPDATE_KEY="+base64.StdEncoding.EncodeToString(private))
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("CNB viva feed: %v\n%s", err, output)
	}
	body, err = os.ReadFile(cnbManifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &feed); err != nil {
		t.Fatal(err)
	}
	asset, err = feed.Asset("windows", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(asset.URL, "https://cnb.cool/xddcode/JarLens/-/releases/download/v9.9.9/") {
		t.Fatalf("CNB update asset points to the wrong host: %+v", asset)
	}
}
