package update

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	vivaupdate "github.com/vivaui/viva/update"
)

const InstallFeed = "https://github.com/xddcode/JarLens/releases/latest/download/latest.json"
const releasePublicKeyText = "7IlaSXjMzIcvIZ5Vde2pj7OZhYD1AOgHU1/FtqI9hEQ="

var ErrInstallUnavailable = errors.New("update: no signed installer is available for this release")

type InstallReason uint8

const (
	InstallSupported InstallReason = iota
	InstallNotUserInstall
	InstallNotWritable
	InstallSystemInstall
)

type InstallProgress struct {
	Done  int64
	Total int64
}

type Download struct {
	Feed   *vivaupdate.Feed
	Asset  vivaupdate.Asset
	client *vivaupdate.Client
}

type StagedUpdate struct {
	Package  *Download
	Artifact string
}

type Installer interface {
	Support() InstallReason
	Prepare(context.Context, *Download, func(InstallProgress)) (StagedUpdate, error)
	Launch(StagedUpdate) error
}

func newUpdateClient() (*vivaupdate.Client, error) {
	key, err := vivaupdate.ParsePublicKey(releasePublicKeyText)
	if err != nil {
		return nil, err
	}
	return vivaupdate.New(vivaupdate.Config{
		FeedURL: InstallFeed, PublicKeys: []ed25519.PublicKey{key},
		App: "io.github.xddcode.jarlens", Channel: "stable", Current: Current,
	})
}

func VerifyReleaseFeed(manifest, signatureText []byte) bool {
	key, err := vivaupdate.ParsePublicKey(releasePublicKeyText)
	if err != nil {
		return false
	}
	signature, err := base64.StdEncoding.DecodeString(string(signatureText))
	return err == nil && ed25519.Verify(key, manifest, signature)
}

func ResolveDownload(ctx context.Context, version string) (*Download, error) {
	client, err := newUpdateClient()
	if err != nil {
		return nil, err
	}
	return resolveDownload(ctx, client, version)
}

func resolveDownload(ctx context.Context, client *vivaupdate.Client, version string) (*Download, error) {
	if client == nil {
		return nil, ErrInstallUnavailable
	}
	feed, err := client.Check(ctx)
	if err != nil || feed == nil || feed.Latest.Version != version {
		return nil, ErrInstallUnavailable
	}
	asset, err := feed.Asset(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return nil, ErrInstallUnavailable
	}
	return &Download{Feed: feed, Asset: asset, client: client}, nil
}

type vivaInstaller struct {
	dir    string
	reason InstallReason
	quit   func()
}

func NewInstaller(quit func()) Installer {
	directory, err := os.UserConfigDir()
	if err != nil {
		return &vivaInstaller{reason: InstallNotUserInstall, quit: quit}
	}
	return &vivaInstaller{
		dir: filepath.Join(directory, "JarLens", "updates"), reason: detectInstallReason(), quit: quit,
	}
}

func (installer *vivaInstaller) Support() InstallReason {
	return installer.reason
}

func (installer *vivaInstaller) Prepare(ctx context.Context, pkg *Download, progress func(InstallProgress)) (StagedUpdate, error) {
	if installer.reason != InstallSupported || pkg == nil || pkg.Feed == nil || pkg.client == nil {
		return StagedUpdate{}, ErrInstallUnavailable
	}
	if installer.dir == "" {
		return StagedUpdate{}, ErrInstallUnavailable
	}
	version := pkg.Feed.Latest.Version
	directory := filepath.Join(installer.dir, version)
	if filepath.Base(directory) != version {
		return StagedUpdate{}, ErrInstallUnavailable
	}
	if err := os.RemoveAll(directory); err != nil {
		return StagedUpdate{}, err
	}
	last := time.Time{}
	artifact, err := pkg.client.Download(ctx, pkg.Asset, directory, func(done, total int64) {
		if progress == nil || done != total && !last.IsZero() && time.Since(last) < 100*time.Millisecond {
			return
		}
		last = time.Now()
		progress(InstallProgress{Done: done, Total: total})
	})
	if err != nil {
		return StagedUpdate{}, err
	}
	return StagedUpdate{Package: pkg, Artifact: artifact}, nil
}

func (installer *vivaInstaller) Launch(staged StagedUpdate) error {
	if staged.Package == nil || staged.Package.Feed == nil || staged.Artifact == "" {
		return ErrInstallUnavailable
	}
	if installer.quit == nil {
		return fmt.Errorf("update: install needs a quit function")
	}
	return vivaupdate.Install(vivaupdate.InstallOptions{
		Artifact: staged.Artifact, Asset: staged.Package.Asset,
		Version: staged.Package.Feed.Latest.Version, App: "jarlens", Args: os.Args[1:],
	}, installer.quit)
}
