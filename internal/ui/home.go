package ui

import (
	"path/filepath"

	"github.com/vivaui/viva/flex"
	vivaui "github.com/vivaui/viva/ui"
)

func (m *Model) paintHome(f *flex.Frame, theme vivaui.Theme) {
	f.Column().Grow(1).Margin(6).Radius(10).Background(theme.Background).Center().Children(func() {
		f.Column().Width(480).MaxHeightPercent(90).Gap(16).Children(func() {
			f.Row().Gap(12).Items(flex.Center).Children(func() {
				if m.appIcon.Width > 0 {
					f.Image(m.appIcon).Size(40, 40)
				}
				f.Column().Gap(4).Children(func() {
					f.Text("JarLens").FontSize(24).Bold()
					f.Text(m.t("打开 Java 归档，浏览和反编译类文件")).FontSize(13).Muted()
				})
			})
			open := f.Row().FillWidth().Height(44).PadX(14).Gap(10).Items(flex.Center).Radius(6).
				Background(theme.Secondary).Hover(flex.Look{}.Background(theme.Hover)).
				HitWith(homeOpen, m.t("打开..."), vivaui.HitOptions{Role: vivaui.RoleButton, Hint: m.t("打开 JAR 或 Class 文件 (Ctrl+O)"), Cursor: vivaui.PointerCursorHand})
			open.Children(func() {
				f.Icon(openGlyph).Size(18, 18).Color(theme.SecondaryForeground)
				f.Text(m.t("打开...")).FontSize(14).Color(theme.AccentForeground)
				f.Spacer()
				f.Text("Ctrl+O").FontSize(12).Muted()
			})
			if open.Clicked() {
				f.Canvas().AfterPaint(m.chooseArchive)
			}
			f.Text(m.t("或将 JAR / Class 文件拖放到此处")).FontSize(13).Muted().TextCenter()
			if len(m.recentProjects) > 0 {
				f.Text(m.t("最近项目")).FontSize(13).Weight(vivaui.WeightSemibold).Color(theme.AccentForeground).MarginSides(6, 0, 0, 0)
				f.ScrollArea(homeRecentList).FillWidth().MaxHeight(420).Gap(4).Children(func() {
					for index, path := range m.recentProjects {
						name := filepath.Base(path)
						row := f.Row().FillWidth().Height(56).PadX(10).Gap(12).Items(flex.Center).Radius(6).
							Hover(flex.Look{}.Background(theme.Hover)).
							HitWith(homeRecentProjects.At(index), name, vivaui.HitOptions{Role: vivaui.RoleButton, Hint: path, Cursor: vivaui.PointerCursorHand})
						row.Children(func() {
							f.Row().Size(28, 28).Shrink(0).Radius(6).Background(projectBadgeColor(path)).Center().Children(func() {
								f.Text(projectInitials(path)).FontSize(12).Weight(vivaui.WeightSemibold).Color(vivaui.RGB24(0xffffff))
							})
							f.Column().Grow(1).Gap(4).Children(func() {
								f.Text(name).FontSize(14).Weight(vivaui.WeightSemibold).Ellipsis().Color(theme.AccentForeground)
								f.Text(path).FontSize(12).Ellipsis().Muted()
							})
						})
						if row.Clicked() {
							f.Canvas().AfterPaint(func() { m.startOpen(path) })
						}
					}
				})
			}
		})
	})
}
