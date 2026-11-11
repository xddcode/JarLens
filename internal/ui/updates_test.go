package ui

import (
	"context"
	"errors"
	"testing"

	jarlensupdate "github.com/xddcode/jarlens/internal/update"
	jarlensworkspace "github.com/xddcode/jarlens/internal/workspace"
)

func TestAutomaticUpdateShowsReleaseDialog(t *testing.T) {
	opened := ""
	m := &Model{
		UpdateChecker: func(context.Context, jarlensupdate.Source) (jarlensupdate.Release, error) {
			return jarlensupdate.Release{Version: "1.1.0", URL: "https://example.com/release", Notes: "- 新功能"}, nil
		},
		openURL: func(url string) error { opened = url; return nil },
	}
	driver := newDriver(t, m)
	defer driver.Close()
	m.CheckForUpdates(true)
	pumpUntil(t, driver, func() bool { return m.updateDialogOpen })
	if m.updateRelease == nil || !shows(driver, "1.1.0") || !shows(driver, "新功能") {
		t.Fatal("new release dialog does not show its version and notes")
	}
	click(t, driver, updateDownload)
	if opened != "https://example.com/release" {
		t.Fatalf("download page = %q", opened)
	}
}

func TestAutomaticUpdateCanBeDisabled(t *testing.T) {
	calls := 0
	m := &Model{UpdateChecker: func(context.Context, jarlensupdate.Source) (jarlensupdate.Release, error) {
		calls++
		return jarlensupdate.Release{}, nil
	}}
	driver := newDriver(t, m)
	defer driver.Close()
	m.settings.CheckUpdates = false
	m.CheckForUpdates(true)
	if calls != 0 {
		t.Fatal("disabled automatic updates contacted the release feed")
	}
}

func TestManualUpdateCheckFromSettings(t *testing.T) {
	m := &Model{
		SettingsStore: &memorySettingsStore{settings: jarlensworkspace.DefaultSettings()},
		UpdateChecker: func(context.Context, jarlensupdate.Source) (jarlensupdate.Release, error) {
			return jarlensupdate.Release{Version: jarlensupdate.Current, URL: "https://example.com/current"}, nil
		},
	}
	driver := newDriver(t, m)
	defer driver.Close()
	openSettingsForTest(t, driver)
	click(t, driver, settingsNav.At(6))
	m.view.ScrollState(settingsContentScroll).ScrollToEnd()
	if err := driver.Move(0, 0); err != nil {
		t.Fatal(err)
	}
	click(t, driver, settingsUpdateCheck)
	pumpUntil(t, driver, func() bool { return m.updateStatus == "当前已是最新版本" })
	if m.updateDialogOpen {
		t.Fatal("same-version release opened an update dialog")
	}
}

func TestUpdateHistoryLoadsExpandsAndPaginates(t *testing.T) {
	type request struct {
		source   jarlensupdate.Source
		page     int
		pageSize int
	}
	requests := []request{}
	m := &Model{
		SettingsStore: &memorySettingsStore{settings: jarlensworkspace.DefaultSettings()},
		UpdateReleasesLoader: func(_ context.Context, source jarlensupdate.Source, page, pageSize int) (jarlensupdate.ReleasePage, error) {
			requests = append(requests, request{source: source, page: page, pageSize: pageSize})
			return jarlensupdate.ReleasePage{
				Releases: []jarlensupdate.Release{{Version: "1." + string(rune('0'+page)) + ".0", URL: "https://example.com/release", Notes: "第 " + string(rune('0'+page)) + " 页更新说明"}},
				Page:     page, HasMore: page == 1,
			}, nil
		},
	}
	driver := newDriver(t, m)
	defer driver.Close()
	openSettingsForTest(t, driver)
	click(t, driver, settingsNav.At(6))
	pumpUntil(t, driver, func() bool { return m.updateReleasesLoaded && len(m.updateReleases) == 1 })
	if !shows(driver, "更新下载源") || !shows(driver, "默认下载源") || !shows(driver, "更新日志") || !shows(driver, "1.1.0") {
		t.Fatal("update management page did not show its source and release history")
	}
	click(t, driver, settingsUpdateMore)
	pumpUntil(t, driver, func() bool { return len(m.updateReleases) == 2 })
	if len(requests) != 2 || requests[0] != (request{source: jarlensupdate.SourceGitHub, page: 1, pageSize: updateReleasesPageSize}) || requests[1].page != 2 || m.updateReleasesHasMore {
		t.Fatalf("history requests = %+v, hasMore=%v", requests, m.updateReleasesHasMore)
	}
	m.view.ScrollState(settingsContentScroll).ScrollTo(0)
	if err := driver.Move(0, 0); err != nil {
		t.Fatal(err)
	}
	click(t, driver, settingsUpdateReleases.At(0))
	if !shows(driver, "第 1 页更新说明") {
		t.Fatal("expanding a release did not show its notes")
	}
}

func TestUpdateSourceSwitchPersistsAndScopesCheck(t *testing.T) {
	store := &memorySettingsStore{settings: jarlensworkspace.DefaultSettings()}
	checked := jarlensupdate.Source("")
	m := &Model{
		SettingsStore: store,
		UpdateReleasesLoader: func(_ context.Context, source jarlensupdate.Source, _, _ int) (jarlensupdate.ReleasePage, error) {
			version := "1.1.0"
			if source == jarlensupdate.SourceCNB {
				version = "1.2.0"
			}
			return jarlensupdate.ReleasePage{Releases: []jarlensupdate.Release{{Version: version, URL: "https://example.com/release"}}}, nil
		},
		UpdateChecker: func(_ context.Context, source jarlensupdate.Source) (jarlensupdate.Release, error) {
			checked = source
			return jarlensupdate.Release{Version: jarlensupdate.Current, URL: "https://example.com/current"}, nil
		},
	}
	driver := newDriver(t, m)
	defer driver.Close()
	openSettingsForTest(t, driver)
	click(t, driver, settingsNav.At(6))
	pumpUntil(t, driver, func() bool { return m.updateReleasesLoaded })
	click(t, driver, settingsUpdateSource.At(2))
	click(t, driver, settingsUpdateSource.At(1))
	pumpUntil(t, driver, func() bool {
		return m.updateReleasesLoaded && m.updateReleasesSource == jarlensupdate.SourceCNB && len(m.updateReleases) == 1 && m.updateReleases[0].Version == "1.2.0"
	})
	click(t, driver, settingsUpdateCheck)
	pumpUntil(t, driver, func() bool { return checked == jarlensupdate.SourceCNB })
	click(t, driver, settingsApply)
	if store.saves != 1 || store.settings.UpdateSource != "cnb" || m.settings.UpdateSource != "cnb" {
		t.Fatalf("update source was not persisted: %+v", store.settings)
	}
}

type fakeUpdateInstaller struct {
	reason    jarlensupdate.InstallReason
	prepare   func(context.Context, func(jarlensupdate.InstallProgress)) (jarlensupdate.StagedUpdate, error)
	launches  int
	launchErr error
}

func (installer *fakeUpdateInstaller) Support() jarlensupdate.InstallReason {
	return installer.reason
}

func (installer *fakeUpdateInstaller) Prepare(ctx context.Context, _ *jarlensupdate.Download, progress func(jarlensupdate.InstallProgress)) (jarlensupdate.StagedUpdate, error) {
	if installer.prepare != nil {
		return installer.prepare(ctx, progress)
	}
	return jarlensupdate.StagedUpdate{}, nil
}

func (installer *fakeUpdateInstaller) Launch(jarlensupdate.StagedUpdate) error {
	installer.launches++
	return installer.launchErr
}

func updateInstallModel(installer *fakeUpdateInstaller, opened *string) *Model {
	return &Model{
		UpdateChecker: func(context.Context, jarlensupdate.Source) (jarlensupdate.Release, error) {
			return jarlensupdate.Release{Version: "9.9.9", URL: "https://example.com/release", Notes: "## 中文\n\n- 新功能"}, nil
		},
		UpdateDownloadResolver: func(context.Context, string) (*jarlensupdate.Download, error) {
			return &jarlensupdate.Download{}, nil
		},
		UpdateInstaller: installer,
		openURL:         func(url string) error { *opened = url; return nil },
	}
}

func TestUpdateInstallDownloadsLaunchesAndRestarts(t *testing.T) {
	opened := ""
	installer := &fakeUpdateInstaller{}
	started, allow := make(chan struct{}), make(chan struct{})
	installer.prepare = func(_ context.Context, progress func(jarlensupdate.InstallProgress)) (jarlensupdate.StagedUpdate, error) {
		progress(jarlensupdate.InstallProgress{Done: 5, Total: 10})
		close(started)
		<-allow
		return jarlensupdate.StagedUpdate{}, nil
	}
	m := updateInstallModel(installer, &opened)
	driver := newDriver(t, m)
	defer driver.Close()
	m.CheckForUpdates(true)
	pumpUntil(t, driver, func() bool { return m.updateDialogOpen })
	if !hasHit(driver, updateInstall, "立即更新") || shows(driver, "## 中文") {
		t.Fatal("signed update is not offered, or Markdown was left unrendered")
	}
	click(t, driver, updateInstall)
	<-started
	pumpUntil(t, driver, func() bool { return m.updateInstall.done == 5 })
	close(allow)
	pumpUntil(t, driver, func() bool { return m.updateInstall.phase == updateInstallRestarting })
	if installer.launches != 1 || opened != "" {
		t.Fatalf("launches = %d, manual page = %q", installer.launches, opened)
	}
}

func TestUpdateInstallFailureKeepsManualDownload(t *testing.T) {
	opened := ""
	installer := &fakeUpdateInstaller{prepare: func(context.Context, func(jarlensupdate.InstallProgress)) (jarlensupdate.StagedUpdate, error) {
		return jarlensupdate.StagedUpdate{}, errors.New("checksum differs")
	}}
	m := updateInstallModel(installer, &opened)
	driver := newDriver(t, m)
	defer driver.Close()
	m.CheckForUpdates(true)
	pumpUntil(t, driver, func() bool { return m.updateDialogOpen })
	click(t, driver, updateInstall)
	pumpUntil(t, driver, func() bool { return m.updateInstall.phase == updateInstallFailed })
	if !shows(driver, "checksum differs") || !shows(driver, "旧版本未被替换") {
		t.Fatal("failed update did not explain the failure and preserve the old version")
	}
	click(t, driver, updateDownload)
	if opened != "https://example.com/release" || installer.launches != 0 {
		t.Fatalf("manual fallback opened %q, launches = %d", opened, installer.launches)
	}
}

func TestUpdateInstallCanCancelDownload(t *testing.T) {
	opened := ""
	started, stopped := make(chan struct{}), make(chan struct{})
	installer := &fakeUpdateInstaller{prepare: func(ctx context.Context, _ func(jarlensupdate.InstallProgress)) (jarlensupdate.StagedUpdate, error) {
		close(started)
		<-ctx.Done()
		close(stopped)
		return jarlensupdate.StagedUpdate{}, ctx.Err()
	}}
	m := updateInstallModel(installer, &opened)
	driver := newDriver(t, m)
	defer driver.Close()
	m.CheckForUpdates(true)
	pumpUntil(t, driver, func() bool { return m.updateDialogOpen })
	click(t, driver, updateInstall)
	<-started
	clicked := false
	for _, hit := range driver.Snapshot().Hits {
		if hit.ID == updateDismiss && hit.Label == "取消" {
			if err := driver.Click(hit.Bounds.X+hit.Bounds.Width/2, hit.Bounds.Y+hit.Bounds.Height/2); err != nil {
				t.Fatal(err)
			}
			clicked = true
			break
		}
	}
	if !clicked {
		t.Fatal("download cancel button is missing")
	}
	<-stopped
	if m.updateDialogOpen || installer.launches != 0 || m.updateInstall.cancel != nil {
		t.Fatal("cancelling left the update active")
	}
}

func TestUpdateInstallUnsupportedKeepsManualDownload(t *testing.T) {
	opened := ""
	installer := &fakeUpdateInstaller{reason: jarlensupdate.InstallNotUserInstall}
	m := updateInstallModel(installer, &opened)
	resolved := false
	m.UpdateDownloadResolver = func(context.Context, string) (*jarlensupdate.Download, error) {
		resolved = true
		return &jarlensupdate.Download{}, nil
	}
	driver := newDriver(t, m)
	defer driver.Close()
	m.CheckForUpdates(true)
	pumpUntil(t, driver, func() bool { return m.updateDialogOpen })
	if hasHit(driver, updateInstall, "立即更新") || !shows(driver, "开发版、便携版") || resolved {
		t.Fatal("unsupported installation was offered a signed download")
	}
	click(t, driver, updateDownload)
	if opened != "https://example.com/release" {
		t.Fatalf("manual page = %q", opened)
	}
}

func TestUpdateInstallMissingSignedPackageFallsBack(t *testing.T) {
	opened := ""
	installer := &fakeUpdateInstaller{}
	m := updateInstallModel(installer, &opened)
	m.UpdateDownloadResolver = func(context.Context, string) (*jarlensupdate.Download, error) {
		return nil, jarlensupdate.ErrInstallUnavailable
	}
	driver := newDriver(t, m)
	defer driver.Close()
	m.CheckForUpdates(true)
	pumpUntil(t, driver, func() bool { return m.updateDialogOpen })
	if hasHit(driver, updateInstall, "立即更新") || !shows(driver, "签名更新包") {
		t.Fatal("missing signed package did not retain the manual fallback")
	}
}
