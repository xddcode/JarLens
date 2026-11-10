package ui

import (
	"context"
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
