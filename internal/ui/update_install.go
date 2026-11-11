package ui

import (
	"context"
	"errors"
	"fmt"

	"github.com/vivaui/viva/components"
	"github.com/vivaui/viva/flex"
	vivaui "github.com/vivaui/viva/ui"
	jarlensupdate "github.com/xddcode/jarlens/internal/update"
)

type updateInstallPhase uint8

const (
	updateInstallIdle updateInstallPhase = iota
	updateInstallDownloading
	updateInstallApplying
	updateInstallRestarting
	updateInstallFailed
)

type updateInstallState struct {
	phase   updateInstallPhase
	serial  uint64
	cancel  context.CancelFunc
	done    int64
	total   int64
	text    string
	failure string
}

type updateInstallResult struct {
	serial   uint64
	phase    updateInstallPhase
	done     int64
	total    int64
	text     string
	staged   jarlensupdate.StagedUpdate
	err      error
	finished bool
}

func (m *Model) canInstallUpdate() bool {
	return m.updateRelease != nil && m.updateDownload != nil && m.UpdateInstaller != nil && m.UpdateInstaller.Support() == jarlensupdate.InstallSupported
}

func (m *Model) updateInstallUnavailableReason() string {
	if m.UpdateInstaller == nil {
		return ""
	}
	switch m.UpdateInstaller.Support() {
	case jarlensupdate.InstallNotUserInstall:
		return m.t("当前是开发版、便携版或非 Windows 安装，请打开下载页手动更新。")
	case jarlensupdate.InstallNotWritable:
		return m.t("当前安装目录不可写，请打开下载页手动更新。")
	case jarlensupdate.InstallSystemInstall:
		return m.t("系统级安装不能由当前用户直接替换，请使用系统安装包更新。")
	default:
		if m.updateDownload == nil {
			return m.t("该版本暂未提供可用的签名更新包，请打开下载页手动更新。")
		}
		return ""
	}
}

func (m *Model) startUpdateInstall() {
	if !m.canInstallUpdate() || m.updateInstall.cancel != nil {
		return
	}
	serial := m.updateInstall.serial + 1
	ctx, cancel := context.WithCancel(context.Background())
	m.updateInstall = updateInstallState{phase: updateInstallDownloading, serial: serial, cancel: cancel, text: m.t("正在下载更新...")}
	installer := m.UpdateInstaller
	pkg := m.updateDownload
	downloadingText := m.t("正在下载更新...")
	progressFormat := m.t("正在下载更新：%s / %s")
	go func() {
		staged, err := installer.Prepare(ctx, pkg, func(progress jarlensupdate.InstallProgress) {
			text := downloadingText
			if progress.Total > 0 {
				text = fmt.Sprintf(progressFormat, formatUpdateBytes(progress.Done), formatUpdateBytes(progress.Total))
			}
			m.publishUpdateInstall(updateInstallResult{
				serial: serial, phase: updateInstallDownloading, done: progress.Done, total: progress.Total, text: text,
			})
		})
		m.publishUpdateInstall(updateInstallResult{serial: serial, phase: updateInstallApplying, staged: staged, err: err, finished: true})
	}()
}

func (m *Model) publishUpdateInstall(result updateInstallResult) {
	m.mu.Lock()
	m.pendingUpdateInstall = &result
	m.mu.Unlock()
	m.invalidator.Invalidate()
}

func (m *Model) applyUpdateInstallResult(result *updateInstallResult) {
	install := &m.updateInstall
	if result.serial != install.serial || install.cancel == nil {
		return
	}
	if !result.finished {
		if install.phase != updateInstallFailed {
			install.phase = result.phase
			install.done = result.done
			install.total = result.total
			install.text = result.text
		}
		return
	}
	install.cancel()
	install.cancel = nil
	if errors.Is(result.err, context.Canceled) {
		return
	}
	if result.err == nil {
		install.phase = updateInstallApplying
		install.text = m.t("正在校验并准备安装...")
		result.err = m.UpdateInstaller.Launch(result.staged)
	}
	if result.err != nil {
		install.phase = updateInstallFailed
		install.failure = result.err.Error()
		return
	}
	install.phase = updateInstallRestarting
	install.text = m.t("更新已就绪，正在重启 JarLens...")
}

func (m *Model) cancelUpdateInstall() {
	m.updateInstall.serial++
	if m.updateInstall.cancel != nil {
		m.updateInstall.cancel()
		m.updateInstall.cancel = nil
	}
	m.updateInstall.phase = updateInstallIdle
	m.updateInstall.failure = ""
	m.updateDialogOpen = false
}

func (m *Model) closeUpdateDialog() {
	if m.updateInstall.cancel != nil {
		m.cancelUpdateInstall()
		return
	}
	m.updateInstall.phase = updateInstallIdle
	m.updateInstall.failure = ""
	m.updateDialogOpen = false
}

func (m *Model) paintUpdateInstallProgress(f *flex.Frame) {
	install := &m.updateInstall
	f.Column().Gap(12).Children(func() {
		if install.phase == updateInstallFailed {
			f.Paragraph(m.tf("无法完成自动更新：%s", install.failure), 4).FontSize(13).Color(WorkspaceTheme(m.Theme).Danger)
			f.Paragraph(m.t("旧版本未被替换，你仍然可以打开下载页手动更新。"), 3).FontSize(13).Muted()
			return
		}
		f.Text(install.text).FontSize(13).Muted()
		if install.phase == updateInstallDownloading {
			if install.total > 0 {
				f.Progress(float32(float64(install.done) / float64(install.total))).Height(8)
			} else {
				f.Progress(0).Indeterminate().Height(8)
			}
		} else {
			f.Progress(0).Indeterminate().Height(8)
		}
	})
}

func (m *Model) paintUpdateNotes(d *flex.DrawContext) {
	height := m.updateNotesView.Layout(d.Canvas, components.MarkdownStyle{Size: 12, LineHeight: 18, MonoFace: m.codeFace}, &m.updateNotesDoc, d.Rect.Width)
	maximum := max(height-d.Rect.Height, 0)
	offset := min(max(m.updateNotesOffset, 0), maximum)
	m.updateNotesMax = maximum
	area := components.BeginScrollArea(d.Canvas, *d.Theme, d.Rect, vivaui.Size{Width: d.Rect.Width, Height: height}).
		Offset(0, offset).Viewport(updateNotes, m.t("更新说明"))
	geometry := area.BeginContent()
	m.updateNotesView.Paint(d.Canvas, components.MarkdownPaint{
		Look: components.MarkdownLookFor(*d.Theme), X: geometry.Content.X, Y: geometry.Content.Y, Clip: geometry.Viewport,
	})
	area.EndContent()
}

func (m *Model) scrollUpdateNotes(delta float32) {
	m.updateNotesOffset = min(max(m.updateNotesOffset+delta, 0), m.updateNotesMax)
}

func formatUpdateBytes(value int64) string {
	const unit = 1024 * 1024
	if value >= unit {
		return fmt.Sprintf("%.1f MB", float64(value)/unit)
	}
	if value >= 1024 {
		return fmt.Sprintf("%.1f KB", float64(value)/1024)
	}
	return fmt.Sprintf("%d B", value)
}
