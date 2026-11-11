package update

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	vivaupdate "github.com/vivaui/viva/update"
)

type installFixture struct {
	client    *vivaupdate.Client
	server    *httptest.Server
	artifact  []byte
	signature string
}

func newInstallFixture(t *testing.T, version string) *installFixture {
	t.Helper()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	fixture := &installFixture{artifact: []byte("test installer")}
	mux := http.NewServeMux()
	fixture.server = httptest.NewTLSServer(mux)
	t.Cleanup(fixture.server.Close)
	format := vivaupdate.FormatNSIS
	if runtime.GOOS == "darwin" {
		format = vivaupdate.FormatZIP
	} else if runtime.GOOS == "linux" {
		format = vivaupdate.FormatTarGz
	}
	digest := sha256.Sum256(fixture.artifact)
	feed := vivaupdate.Feed{
		Schema: vivaupdate.SchemaVersion, App: "io.github.xddcode.jarlens", Channel: "stable",
		Latest: vivaupdate.Release{Version: version, Date: time.Now(), Assets: []vivaupdate.Asset{{
			OS: runtime.GOOS, Arch: runtime.GOARCH, Format: format,
			URL: fixture.server.URL + "/setup.exe", Size: int64(len(fixture.artifact)), SHA256: hex.EncodeToString(digest[:]),
		}}},
	}
	manifest, err := json.Marshal(feed)
	if err != nil {
		t.Fatal(err)
	}
	fixture.signature, err = vivaupdate.Sign(manifest, private)
	if err != nil {
		t.Fatal(err)
	}
	mux.HandleFunc("/latest.json", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(manifest) })
	mux.HandleFunc("/latest.json.sig", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(fixture.signature)) })
	mux.HandleFunc("/setup.exe", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(fixture.artifact) })
	fixture.client, err = vivaupdate.New(vivaupdate.Config{
		FeedURL: fixture.server.URL + "/latest.json", PublicKeys: []ed25519.PublicKey{public},
		App: "io.github.xddcode.jarlens", Channel: "stable", Current: "0.1.0", HTTP: fixture.server.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return fixture
}

func TestReleasePublicKeyParses(t *testing.T) {
	if _, err := vivaupdate.ParsePublicKey(releasePublicKeyText); err != nil {
		t.Fatal(err)
	}
}

func TestResolveDownloadVerifiesSignedFeed(t *testing.T) {
	fixture := newInstallFixture(t, "9.9.9")
	pkg, err := resolveDownload(context.Background(), fixture.client, "9.9.9")
	if err != nil {
		t.Fatal(err)
	}
	if pkg.Feed.Latest.Version != "9.9.9" || pkg.Asset.URL != fixture.server.URL+"/setup.exe" {
		t.Fatalf("unexpected package: %+v", pkg)
	}
	if _, err := resolveDownload(context.Background(), fixture.client, "9.9.8"); !errors.Is(err, ErrInstallUnavailable) {
		t.Fatalf("version mismatch error = %v", err)
	}
	fixture.signature = "invalid"
	if _, err := resolveDownload(context.Background(), fixture.client, "9.9.9"); !errors.Is(err, ErrInstallUnavailable) {
		t.Fatalf("invalid signature error = %v", err)
	}
}

func TestPrepareDownloadsAndRejectsChangedBytes(t *testing.T) {
	fixture := newInstallFixture(t, "9.9.9")
	pkg, err := resolveDownload(context.Background(), fixture.client, "9.9.9")
	if err != nil {
		t.Fatal(err)
	}
	installer := &vivaInstaller{dir: t.TempDir(), reason: InstallSupported}
	var last InstallProgress
	staged, err := installer.Prepare(context.Background(), pkg, func(progress InstallProgress) { last = progress })
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(staged.Artifact)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "test installer" || last.Done != int64(len(data)) || last.Total != int64(len(data)) {
		t.Fatalf("download = %q, progress = %+v", data, last)
	}

	fixture.artifact = []byte("tampered setup")
	if _, err := installer.Prepare(context.Background(), pkg, nil); !errors.Is(err, vivaupdate.ErrChecksum) {
		t.Fatalf("tampered download error = %v", err)
	}
	if _, err := os.Stat(staged.Artifact); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("old artifact survived a failed replacement download: %v", err)
	}
}

func TestLaunchRejectsWrongPackageWithoutQuitting(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows installer validation")
	}
	artifact := filepath.Join(t.TempDir(), "setup.exe")
	body := []byte("zip")
	if err := os.WriteFile(artifact, body, 0o600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(body)
	quit := false
	installer := &vivaInstaller{quit: func() { quit = true }}
	err := installer.Launch(StagedUpdate{
		Package: &Download{Feed: &vivaupdate.Feed{Latest: vivaupdate.Release{Version: "9.9.9"}}, Asset: vivaupdate.Asset{
			OS: "windows", Arch: runtime.GOARCH, Format: vivaupdate.FormatZIP,
			Size: int64(len(body)), SHA256: hex.EncodeToString(digest[:]),
		}}, Artifact: artifact,
	})
	if !errors.Is(err, vivaupdate.ErrManifest) || quit {
		t.Fatalf("wrong package error = %v, quit = %v", err, quit)
	}
}

func TestPublishedFeedSignature(t *testing.T) {
	directory := os.Getenv("JARLENS_UPDATE_FEED_DIR")
	if directory == "" {
		t.Skip("set JARLENS_UPDATE_FEED_DIR to verify a release feed")
	}
	manifest, err := os.ReadFile(filepath.Join(directory, "latest.json"))
	if err != nil {
		t.Fatal(err)
	}
	signature, err := os.ReadFile(filepath.Join(directory, "latest.json.sig"))
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyReleaseFeed(manifest, signature) {
		t.Fatal("release feed does not verify with the application's public key")
	}
}
