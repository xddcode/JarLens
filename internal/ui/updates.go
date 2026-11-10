package ui

import (
	"context"
	"errors"
	"time"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/flex"
	vivaui "github.com/vivaui/viva/ui"
	jarlensupdate "github.com/xddcode/jarlens/internal/update"
)

func (m *Model) CheckForUpdates(automatic bool) {
	if m.updateChecking || automatic && !m.settings.CheckUpdates {
		return
	}
	checker := m.UpdateChecker
	if checker == nil {
		checker = func(ctx context.Context, source jarlensupdate.Source) (jarlensupdate.Release, error) {
			return jarlensupdate.LatestFromSource(ctx, nil, source)
		}
	}
	source := jarlensupdate.ParseSource(m.settings.UpdateSource)
	if m.settingsOpen {
		source = jarlensupdate.ParseSource(m.settingsDraft.UpdateSource)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	m.updateCancel = cancel
	m.updateChecking = true
	if !automatic {
		m.updateStatus = m.t("正在检查更新...")
	}
	generation := m.generation
	go func() {
		release, err := checker(ctx, source)
		result := &taskResult{kind: taskUpdate, generation: generation, automatic: automatic, err: err}
		if err == nil {
			result.update = &release
		}
		m.publish(result)
	}()
}

func (m *Model) applyUpdateResult(result *taskResult) {
	m.updateChecking = false
	m.updateCancel = nil
	if result.err != nil {
		if !result.automatic {
			if errors.Is(result.err, jarlensupdate.ErrNoRelease) {
				m.updateStatus = m.t("暂未发现可用更新")
			} else {
				m.updateStatus = m.t("暂时无法检查更新，请稍后重试")
			}
		}
		return
	}
	if result.update == nil {
		return
	}
	if !jarlensupdate.Newer(jarlensupdate.Current, result.update.Version) {
		if !result.automatic {
			m.updateStatus = m.t("当前已是最新版本")
		}
		return
	}
	m.updateRelease = result.update
	m.updateDialogOpen = true
	m.updateStatus = m.tf("发现新版本 %s", result.update.Version)
}

func (m *Model) openUpdateRelease() {
	if m.updateRelease == nil {
		return
	}
	opener := m.openURL
	if opener == nil {
		opener = app.OpenURL
	}
	if err := opener(m.updateRelease.URL); err != nil {
		m.updateStatus = m.t("无法打开下载页面")
		return
	}
	m.updateStatus = m.t("已打开新版本下载页面")
}

func (m *Model) paintUpdateDialog(f *flex.Frame, theme vivaui.Theme) {
	release := m.updateRelease
	if release == nil {
		return
	}
	dialog := f.Dialog(flex.DialogSpec{
		Title: m.tf("发现新版本 %s", release.Version), Width: 600, Layer: 4,
		Scope: updateDialog, DismissID: updateDismiss, CloseID: updateClose,
	})
	dialog.Body(func() {
		f.Column().Gap(12).Children(func() {
			f.Text(m.tf("当前版本 %s，新版本 %s。", jarlensupdate.Current, release.Version)).FontSize(13)
			notes := release.Notes
			if notes == "" {
				notes = m.t("该版本没有提供更新说明。")
			}
			f.ScrollArea(updateNotes).Height(220).Children(func() {
				f.Paragraph(notes, 12).FontSize(12).Muted()
			})
		})
	})
	dialog.Footer(func() {
		if f.Button(updateDismiss, m.t("稍后")).Outline().Clicked() {
			m.updateDialogOpen = false
		}
		if f.Button(updateDownload, m.t("打开下载页")).Clicked() {
			m.openUpdateRelease()
		}
	})
	if dialog.Dismissed() {
		m.updateDialogOpen = false
	}
	_ = theme
}
