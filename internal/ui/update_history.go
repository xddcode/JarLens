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

const updateReleasesPageSize = 5

type updateReleasesResult struct {
	generation uint64
	source     jarlensupdate.Source
	page       int
	append     bool
	releases   jarlensupdate.ReleasePage
	err        error
}

func (m *Model) selectedUpdateSource() jarlensupdate.Source {
	return jarlensupdate.ParseSource(m.settingsDraft.UpdateSource)
}

func (m *Model) ensureUpdateReleaseHistory(f *flex.Frame) {
	source := m.selectedUpdateSource()
	if m.updateReleasesRequested || m.updateReleasesSource == source && (m.updateReleasesLoading || m.updateReleasesLoaded) {
		return
	}
	m.updateReleasesRequested = true
	f.Canvas().AfterPaint(func() {
		m.updateReleasesRequested = false
		if m.settingsOpen && m.settingsCategory == 6 && m.selectedUpdateSource() == source {
			m.loadUpdateReleases(false)
		}
	})
}

func (m *Model) loadUpdateReleases(more bool) {
	if m.updateReleasesLoading {
		if more {
			return
		}
		m.cancelUpdateReleases()
	}
	if more && !m.updateReleasesHasMore {
		return
	}
	loader := m.UpdateReleasesLoader
	if loader == nil {
		loader = func(ctx context.Context, source jarlensupdate.Source, page, pageSize int) (jarlensupdate.ReleasePage, error) {
			return jarlensupdate.Releases(ctx, nil, source, "", page, pageSize)
		}
	}
	source := m.selectedUpdateSource()
	page := 1
	if more {
		page = m.updateReleasesPage + 1
	} else {
		m.updateReleases = nil
		m.updateReleasesPage = 0
		m.updateReleasesHasMore = false
		m.updateReleasesLoaded = false
		m.updateReleaseExpanded = make(map[string]bool)
	}
	m.updateReleasesGeneration++
	generation := m.updateReleasesGeneration
	m.updateReleasesSource = source
	m.updateReleasesLoading = true
	m.updateReleasesError = ""
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	m.updateReleasesCancel = cancel
	go func() {
		defer cancel()
		releases, err := loader(ctx, source, page, updateReleasesPageSize)
		m.mu.Lock()
		m.pendingUpdateReleases = &updateReleasesResult{
			generation: generation, source: source, page: page, append: more, releases: releases, err: err,
		}
		m.mu.Unlock()
		m.invalidator.Invalidate()
	}()
}

func (m *Model) cancelUpdateReleases() {
	m.updateReleasesGeneration++
	if m.updateReleasesCancel != nil {
		m.updateReleasesCancel()
		m.updateReleasesCancel = nil
	}
	m.updateReleasesLoading = false
	m.updateReleasesRequested = false
}

func (m *Model) applyUpdateReleasesResult(result *updateReleasesResult) {
	if result.generation != m.updateReleasesGeneration {
		return
	}
	m.updateReleasesLoading = false
	m.updateReleasesCancel = nil
	m.updateReleasesLoaded = true
	m.updateReleasesSource = result.source
	if result.err != nil && !errors.Is(result.err, jarlensupdate.ErrNoRelease) {
		m.updateReleasesError = m.t("暂时无法获取更新日志，请检查网络后重试")
		return
	}
	m.updateReleasesPage = result.page
	m.updateReleasesHasMore = result.releases.HasMore
	if result.append {
		m.updateReleases = append(m.updateReleases, result.releases.Releases...)
	} else {
		m.updateReleases = result.releases.Releases
	}
}

func (m *Model) openUpdateSource() {
	opener := m.openURL
	if opener == nil {
		opener = app.OpenURL
	}
	if err := opener(m.selectedUpdateSource().ReleasesURL()); err != nil {
		m.updateStatus = m.t("无法打开版本发布页面")
		return
	}
	m.updateStatus = m.t("已打开版本发布页面")
}

func (m *Model) paintUpdateSettings(f *flex.Frame, theme vivaui.Theme) {
	m.ensureUpdateReleaseHistory(f)
	settingsGroup(f, theme, "", func() {
		m.settingsRow(f, theme, "更新下载源", func() {
			selected := 0
			if m.selectedUpdateSource() == jarlensupdate.SourceCNB {
				selected = 1
			}
			selector := f.Select(settingsUpdateSource, &m.settingsUpdateSourceSelect, []string{m.t("默认下载源"), "CNB"}, selected).Width(240).Label(m.t("更新下载源"))
			if m.settingsUpdateSourceSelect.Open {
				selector = selector.DrawOver(func(d *flex.DrawContext) { d.Canvas.EndPresence() })
			}
			if index, ok := selector.Picked(); ok {
				source := jarlensupdate.SourceGitHub
				if index == 1 {
					source = jarlensupdate.SourceCNB
				}
				if source != m.selectedUpdateSource() {
					m.settingsDraft.UpdateSource = string(source)
					m.loadUpdateReleases(false)
				}
			}
		})
	})
	settingsGroup(f, theme, m.t("更新"), func() {
		m.settingsToggleRow(f, theme, settingsCheckUpdates, "自动检查更新", m.settingsDraft.CheckUpdates, func() {
			m.settingsDraft.CheckUpdates = !m.settingsDraft.CheckUpdates
		})
		m.settingsRow(f, theme, "当前版本", func() {
			f.Text(jarlensupdate.Current).FontSize(13).Muted()
		})
		if m.updateStatus != "" {
			f.Text(m.updateStatus).FontSize(12).Muted().PadSides(2, 0, 4, 0)
		}
	})
	settingsGroup(f, theme, m.t("更新日志"), func() {
		f.Row().MinHeight(32).Gap(8).Items(flex.Center).Children(func() {
			f.Text(m.t("查看所选来源发布的各版本更新内容")).FontSize(12).Muted()
			f.Spacer()
			if f.Button(settingsUpdateCheck, m.t("检查更新")).Outline().Small().Disabled(m.updateChecking).Clicked() {
				m.CheckForUpdates(false)
			}
			if f.Button(settingsUpdateOpen, m.t("在官网查看")).Outline().Small().Clicked() {
				m.openUpdateSource()
			}
			if smallIconButton(f, settingsUpdateRefresh, m.t("刷新更新日志"), refreshGlyph, theme, !m.updateReleasesLoading) {
				m.loadUpdateReleases(false)
			}
		})
		m.paintUpdateReleaseHistory(f, theme)
	})
}

func (m *Model) paintUpdateReleaseHistory(f *flex.Frame, theme vivaui.Theme) {
	if m.updateReleasesLoading && len(m.updateReleases) == 0 {
		f.Row().Height(64).Center().Gap(8).Children(func() {
			f.Spinner()
			f.Text(m.t("正在获取更新日志...")).FontSize(12).Muted()
		})
		return
	}
	if m.updateReleasesError != "" && len(m.updateReleases) == 0 {
		f.Column().Height(64).Center().Gap(6).Children(func() {
			f.Text(m.updateReleasesError).FontSize(12).Color(theme.Danger)
			if f.Button(settingsUpdateRetry, m.t("重试")).Outline().Small().Clicked() {
				m.loadUpdateReleases(false)
			}
		})
		return
	}
	if len(m.updateReleases) == 0 {
		f.Row().Height(56).Center().Children(func() {
			f.Text(m.t("该来源还没有发布正式版本")).FontSize(12).Muted()
		})
		return
	}
	f.Column().Gap(6).PadSides(4, 0, 0, 0).Children(func() {
		for index, release := range m.updateReleases {
			key := string(m.updateReleasesSource) + "/" + release.Version
			expanded := m.updateReleaseExpanded[key]
			f.Column().FillWidth().Radius(6).Border(1, theme.Border).Background(theme.Card).Children(func() {
				header := f.Row().MinHeight(38).PadSides(0, 10, 0, 10).Gap(8).Items(flex.Center).
					HitWith(settingsUpdateReleases.At(index), m.tf("查看 %s 更新日志", release.Version), vivaui.HitOptions{
						Role: vivaui.RoleButton, Expanded: expanded, Expandable: true, Cursor: vivaui.PointerCursorHand,
					})
				header.Children(func() {
					f.Text(release.Version).FontSize(12).Weight(vivaui.WeightSemibold)
					if release.Version == jarlensupdate.Current {
						f.Text(m.t("当前")).FontSize(10).Color(theme.AccentForeground).Background(theme.Accent).Radius(3).PadSides(0, 5, 0, 5)
					} else if jarlensupdate.Newer(jarlensupdate.Current, release.Version) {
						f.Text("NEW").FontSize(10).Color(theme.PrimaryForeground).Background(theme.Primary).Radius(3).PadSides(0, 5, 0, 5)
					}
					if !release.PublishedAt.IsZero() {
						f.Text(m.t("发布于 ") + release.PublishedAt.Local().Format(m.t("2006年1月2日"))).FontSize(11).Muted()
					}
					f.Spacer()
					glyph := chevronRightGlyph
					if expanded {
						glyph = chevronDownGlyph
					}
					f.Icon(glyph).Size(14, 14).Color(theme.MutedForeground)
				})
				if header.Clicked() {
					if m.updateReleaseExpanded == nil {
						m.updateReleaseExpanded = make(map[string]bool)
					}
					m.updateReleaseExpanded[key] = !expanded
				}
				if expanded {
					notes := release.Notes
					if notes == "" {
						notes = m.t("该版本没有提供更新说明。")
					}
					f.Divider()
					f.Paragraph(notes, 100).FontSize(12).Muted().Pad(12)
				}
			})
		}
		if m.updateReleasesHasMore {
			if f.Button(settingsUpdateMore, m.t("加载更多")).Outline().Small().FillWidth().Disabled(m.updateReleasesLoading).Clicked() {
				m.loadUpdateReleases(true)
			}
		}
		if m.updateReleasesError != "" {
			f.Text(m.updateReleasesError).FontSize(12).Color(theme.Danger)
		}
	})
}
