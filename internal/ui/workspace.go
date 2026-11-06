package ui

import (
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/components"
	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/text"
	vivaui "github.com/vivaui/viva/ui"
	"github.com/xddcode/jarlens/internal/archive"
)

var (
	tree                           = vivaui.Tree{}
	workbench                      = tree.Feature("workbench")
	themeToggle                    = tree.ID(workbench, "theme-toggle")
	githubOpen                     = tree.ID(workbench, "github-open")
	archiveOpen                    = tree.ID(workbench, "archive-open")
	archiveSave                    = tree.ID(workbench, "archive-save")
	workCancel                     = tree.ID(workbench, "work-cancel")
	mainMenuButton                 = tree.ID(workbench, "main-menu-button")
	mainMenuTitle                  = tree.ID(workbench, "main-menu-title")
	mainMenu                       = tree.ID(workbench, "main-menu")
	mainMenuDismiss                = tree.ID(workbench, "main-menu-dismiss")
	mainMenuOpenFile               = tree.ID(workbench, "main-menu-open-file")
	mainMenuHome                   = tree.ID(workbench, "main-menu-home")
	mainMenuCurrentProject         = tree.ID(workbench, "main-menu-current-project")
	homeOpen                       = tree.ID(workbench, "home-open")
	homeRecentList                 = tree.ID(workbench, "home-recent-list")
	homeRecentProjects             = tree.Sub(workbench, tree.Name("home-recent-projects"))
	projectMenuScrollbar           = tree.ID(workbench, "project-menu-scrollbar")
	recentProjects                 = tree.Sub(workbench, tree.Name("recent-projects"))
	projectTree                    = tree.ID(workbench, "project-tree")
	projectFilter                  = tree.ID(workbench, "project-filter")
	projectSearch                  = tree.ID(workbench, "project-search")
	projectTitle                   = tree.ID(workbench, "project-title")
	projectMenu                    = tree.ID(workbench, "project-menu")
	projectMenuDismiss             = tree.ID(workbench, "project-menu-dismiss")
	projectExpandAll               = tree.ID(workbench, "project-expand-all")
	projectCollapseAll             = tree.ID(workbench, "project-collapse-all")
	projectFilterMenu              = tree.ID(workbench, "project-filter-menu")
	projectSplitter                = tree.ID(workbench, "project-splitter")
	projectHorizontalScrollbar     = tree.ID(workbench, "project-horizontal-scrollbar")
	sourceEditor                   = tree.ID(workbench, "source-editor")
	sourceScrollbar                = tree.ID(workbench, "source-scrollbar")
	sourceHorizontalScrollbar      = tree.ID(workbench, "source-horizontal-scrollbar")
	splitSourceEditor              = tree.ID(workbench, "split-source-editor")
	splitSourceScrollbar           = tree.ID(workbench, "split-source-scrollbar")
	splitSourceHorizontalScrollbar = tree.ID(workbench, "split-source-horizontal-scrollbar")
	projectNodes                   = tree.Sub(workbench, tree.Name("project-nodes"))
	projectDisclosures             = tree.Sub(workbench, tree.Name("project-disclosures"))
	editorTabs                     = tree.Sub(workbench, tree.Name("editor-tabs"))
	editorTabClose                 = tree.Sub(workbench, tree.Name("editor-tab-close"))
	editorTabsMenu                 = tree.ID(workbench, "editor-tabs-menu")
	editorTabsMenuOpen             = tree.ID(workbench, "editor-tabs-menu-open")
	editorTabsMenuDismiss          = tree.ID(workbench, "editor-tabs-menu-dismiss")
	editorTabsCloseAll             = tree.ID(workbench, "editor-tabs-close-all")
	editorTabsCloseOthers          = tree.ID(workbench, "editor-tabs-close-others")
	editorTabsCloseLeft            = tree.ID(workbench, "editor-tabs-close-left")
	editorTabsCloseRight           = tree.ID(workbench, "editor-tabs-close-right")
	editorTabContextMenu           = tree.ID(workbench, "editor-tab-context-menu")
	editorTabContextDismiss        = tree.ID(workbench, "editor-tab-context-dismiss")
	editorTabCloseCurrent          = tree.ID(workbench, "editor-tab-close-current")
	editorTabSplitRight            = tree.ID(workbench, "editor-tab-split-right")
	editorTabSplitDown             = tree.ID(workbench, "editor-tab-split-down")
	sourceFind                     = tree.ID(workbench, "source-find")
	findPrevious                   = tree.ID(workbench, "find-previous")
	findNext                       = tree.ID(workbench, "find-next")
	findClose                      = tree.ID(workbench, "find-close")
	goLineDialog                   = tree.ID(workbench, "go-line-dialog")
	goLineInput                    = tree.ID(workbench, "go-line-input")
	goLineConfirm                  = tree.ID(workbench, "go-line-confirm")
	goLineClose                    = tree.ID(workbench, "go-line-close")
)

func NewConfig(m *Model) app.Config {
	m.initialize()
	m.tooltip.ShowOnFocus()
	invalidator := &app.Invalidator{}
	m.invalidator = invalidator
	textFields := flex.Fields(&m.filter, &m.findField, &m.goLineField)
	config := app.Config{
		Tree: &tree, Title: "JarLens", Width: 1200, Height: 760,
		MinWidth: 900, MinHeight: 600,
		WindowMode: app.WindowCustomTitlebar, TitlebarHeight: 40,
		ClearColor:  app.Color{Red: 38.0 / 255, Green: 40.0 / 255, Blue: 44.0 / 255, Alpha: 1},
		Invalidator: invalidator, OnWake: m.wake, OnClosed: m.close,
		WindowControls: func() app.WindowControlsStyle {
			theme := WorkspaceTheme(m.Theme)
			return app.WindowControlsStyle{
				Foreground: theme.Foreground, Hover: theme.Hover, Pressed: theme.Muted,
				Close: vivaui.RGB24(0xc42b1c), CloseForeground: vivaui.RGB24(0xffffff),
				Minimize: "最小化", Maximize: "最大化", Restore: "还原", CloseName: "关闭",
			}
		},
		TextInput: func(context app.TextInputContext) *app.TextInput {
			if context.ID == sourceEditor && m.editor != nil {
				return &app.TextInput{Editor: m.editor}
			}
			if context.ID == splitSourceEditor && m.secondary != nil {
				return &app.TextInput{Editor: m.secondary.editor}
			}
			return textFields(context)
		},
		OnFileDrop: func(event vivaui.FileDropEvent) vivaui.EventResult {
			switch event.Phase {
			case vivaui.FileDropOver:
				return vivaui.EventResult{Handled: true}
			case vivaui.FileDropDrop:
				if len(event.Paths) > 0 {
					m.startOpen(event.Paths[0])
				}
				return vivaui.EventResult{Handled: true, Repaint: true}
			default:
				return vivaui.EventResult{}
			}
		},
		OnShortcut: func(event vivaui.KeyEvent) vivaui.EventResult {
			m.tooltip.Dismiss()
			primary := event.Modifiers & (vivaui.ModifierControl | vivaui.ModifierMeta)
			if primary != 0 && event.Key == vivaui.KeyEnter {
				delta := 1
				if event.Modifiers&vivaui.ModifierShift != 0 {
					delta = -1
				}
				m.findNext(delta)
				return vivaui.EventResult{Handled: true, Repaint: true}
			}
			if primary == 0 {
				return vivaui.EventResult{}
			}
			switch strings.ToLower(event.Text) {
			case "o":
				m.chooseArchive()
				return vivaui.EventResult{Handled: true}
			case "s":
				m.saveSource()
				return vivaui.EventResult{Handled: true}
			case "f":
				if event.Modifiers&vivaui.ModifierShift != 0 {
					m.filterOpen = true
				} else {
					m.findOpen = true
				}
				return vivaui.EventResult{Handled: true, Repaint: true}
			case "g":
				m.goLineOpen = true
				m.goLineField.SetText("")
				return vivaui.EventResult{Handled: true, Repaint: true}
			case "w":
				m.closeActiveDocument()
				return vivaui.EventResult{Handled: true}
			default:
				return vivaui.EventResult{}
			}
		},
		OnKey: func(event vivaui.KeyEvent) vivaui.EventResult {
			if event.Release {
				return vivaui.EventResult{}
			}
			m.tooltip.Dismiss()
			if event.Key == vivaui.KeyEscape && m.findOpen {
				m.findOpen = false
				return vivaui.EventResult{Handled: true, Repaint: true}
			}
			if event.Key == vivaui.KeyEscape && m.filterOpen && event.Target == projectFilter {
				m.filterOpen = false
				return vivaui.EventResult{Handled: true, Repaint: true}
			}
			if event.Target == projectSplitter {
				return m.split.Key(event)
			}
			if !event.Release && event.Target == splitSourceEditor && m.secondary != nil {
				movement, extend := components.TextareaMovementForKey(event.Key, event.Modifiers)
				if movement == components.TextareaMoveNone {
					return vivaui.EventResult{}
				}
				value := m.secondary.editor.String()
				offset := m.secondary.code.Move(value, m.secondary.editor.ByteOffset(m.secondary.editor.Caret()), movement)
				m.secondary.editor.MoveTo(utf8.RuneCountInString(value[:offset]), extend)
				return vivaui.EventResult{Handled: true, Repaint: true}
			}
			if event.Release || event.Target != sourceEditor || m.editor == nil {
				return vivaui.EventResult{}
			}
			movement, extend := components.TextareaMovementForKey(event.Key, event.Modifiers)
			if movement == components.TextareaMoveNone {
				return vivaui.EventResult{}
			}
			value := m.editor.String()
			offset := m.code.Move(value, m.editor.ByteOffset(m.editor.Caret()), movement)
			m.editor.MoveTo(utf8.RuneCountInString(value[:offset]), extend)
			return vivaui.EventResult{Handled: true, Repaint: true}
		},
		OnPointer: func(event vivaui.PointerEvent) vivaui.EventResult {
			if m.mainMenuOpen && m.projectMenuScroll.Drag(event) {
				return vivaui.EventResult{Handled: true, Repaint: true}
			}
			m.sourcePointer = vivaui.Point{X: event.X, Y: event.Y}
			m.sourcePointerInside = event.Target == sourceEditor && event.Kind != vivaui.PointerExit
			if event.Target == projectHorizontalScrollbar {
				return m.projectTreePointer(event)
			}
			if event.Target == sourceEditor {
				if result, handled := m.sourceLinkPointer(event); handled {
					return result
				}
			}
			if m.secondary != nil && (event.Target == splitSourceEditor || event.Target == splitSourceScrollbar || event.Target == splitSourceHorizontalScrollbar) {
				return m.secondary.code.Pointer(event)
			}
			if event.Target == sourceEditor || event.Target == sourceScrollbar || event.Target == sourceHorizontalScrollbar {
				return m.code.Pointer(event)
			}
			return vivaui.EventResult{}
		},
		OnScroll: func(event vivaui.ScrollEvent) vivaui.EventResult {
			m.tooltip.Dismiss()
			if m.mainMenuOpen && m.projectMenuScroll.Wheel(event) {
				return vivaui.EventResult{Handled: true, Repaint: true}
			}
			if m.secondary != nil && (event.Target == splitSourceEditor || event.Target == splitSourceScrollbar || event.Target == splitSourceHorizontalScrollbar) {
				return m.secondary.code.Scroll(event)
			}
			if event.Target == sourceEditor || event.Target == sourceScrollbar || event.Target == sourceHorizontalScrollbar {
				return m.code.Scroll(event)
			}
			if projectTreeEventTarget(event.Target) {
				delta := event.StepX(24)
				if delta == 0 && event.Modifiers&vivaui.ModifierShift != 0 {
					delta = event.StepY(24)
				}
				if delta != 0 {
					m.treeScrollX = max(0, m.treeScrollX+delta)
					return vivaui.EventResult{Handled: true, Repaint: true}
				}
			}
			return m.view.Wheel(event)
		},
	}

	config.Paint = func(canvas *vivaui.Canvas, size vivaui.Size) {
		if m.findOpen {
			canvas.RequestFocus(sourceFind)
		}
		if m.filterOpen {
			canvas.RequestFocus(projectFilter)
		}
		if m.focusSource {
			canvas.RequestFocus(sourceEditor)
			m.focusSource = false
		}
		if m.focusSplit {
			canvas.RequestFocus(splitSourceEditor)
			m.focusSplit = false
		}
		m.paint(canvas, size)
	}
	return config
}

func (m *Model) paint(canvas *vivaui.Canvas, size vivaui.Size) {
	theme := WorkspaceTheme(m.Theme)
	bounds := vivaui.Rect{Width: size.Width, Height: size.Height}
	canvas.Fill(bounds, theme.Card)

	themeLabel := "使用浅色主题"
	themeGlyph := sunGlyph
	if m.Theme == AppearanceLight {
		themeLabel = "使用深色主题"
		themeGlyph = moonGlyph
	}
	status := m.status
	if status == "" {
		status = "就绪"
	}
	leading, trailing := app.TitlebarInsets()
	m.updateFind()
	var sourceSlot flex.Box
	var secondarySourceSlot flex.Box
	var splitterSlot flex.Box
	var editorMenuButton flex.Box
	var projectTreeBox flex.Box
	var mainMenuTrigger flex.Box
	var mainMenuTitleBox flex.Box
	var projectTitleBox flex.Box
	treeContentWidth := m.projectTreeContentWidth(canvas)
	maxTreeScroll := max(float32(0), treeContentWidth-m.split.Pos)
	treeScrollX := min(max(m.treeScrollX, 0), maxTreeScroll)
	loading := m.busy && m.workspace == nil
	if doc := m.currentDocument(); doc != nil && doc.loading {
		loading = true
	}
	title := "JarLens"
	if m.workspace != nil && !m.homeVisible {
		title = filepathBase(m.workspace.Archive.Path)
	}

	m.view.Draw(canvas, theme, bounds, func(f *flex.Frame) {
		f.Column().Fill().Background(theme.Card).Children(func() {
			titlebar := f.Row().Height(40).PadSides(0, trailing+8, 0, leading+10).Gap(10).Items(flex.Center).Justify(flex.Between).MovesWindow()
			titlebar.Children(func() {
				f.Row().Height(34).Gap(10).Items(flex.Center).Children(func() {
					if m.appIcon.Width > 0 {
						f.Image(m.appIcon).Size(20, 20)
					} else {
						f.Row().Size(20, 20).Radius(4).Background(theme.Primary).Center().Children(func() {
							f.Text("J").FontSize(11).Bold().Color(theme.PrimaryForeground)
						})
					}
					mainMenuTrigger = f.Row().Size(30, 30).Center().Radius(8).HoverBackground(theme.Hover).
						HitWith(mainMenuButton, "主菜单", vivaui.HitOptions{Role: vivaui.RoleButton, Hint: "主菜单", Expanded: m.mainMenuOpen && !m.mainMenuFromTitle, Cursor: vivaui.PointerCursorHand})
					if m.mainMenuOpen && !m.mainMenuFromTitle {
						mainMenuTrigger = mainMenuTrigger.Background(theme.Hover)
					}
					mainMenuTrigger.Children(func() {
						f.Icon(menuGlyph).Size(18, 18).Color(theme.SecondaryForeground)
					})
					if mainMenuTrigger.Clicked() {
						if m.mainMenuOpen && !m.mainMenuFromTitle {
							m.mainMenuOpen = false
						} else {
							m.mainMenuOpen = true
							m.mainMenuFromTitle = false
						}
					}
					if m.workspace == nil || m.homeVisible {
						f.Row().Height(30).PadSides(0, 8, 0, 0).Gap(6).Items(flex.Center).Children(func() {
							f.Icon(homeGlyph).Size(16, 16).Color(theme.SecondaryForeground)
							f.Text("Home").FontSize(13).Weight(vivaui.WeightSemibold).Color(theme.AccentForeground)
						})
					} else {
						mainMenuTitleBox = f.Row().Height(30).MaxWidth(340).PadSides(0, 8, 0, 0).Gap(6).Items(flex.Center).Radius(8).
							HoverBackground(theme.Hover).
							HitWith(mainMenuTitle, title, vivaui.HitOptions{Role: vivaui.RoleButton, Hint: "归档菜单", Expanded: m.mainMenuOpen && m.mainMenuFromTitle, Cursor: vivaui.PointerCursorHand})
						if m.mainMenuOpen && m.mainMenuFromTitle {
							mainMenuTitleBox = mainMenuTitleBox.Background(theme.Hover)
						}
						mainMenuTitleBox.Children(func() {
							f.Row().Size(20, 20).Radius(4).Background(projectBadgeColor(title)).Center().Children(func() {
								f.Text(projectInitials(title)).FontSize(11).Weight(vivaui.WeightSemibold).Color(vivaui.RGB24(0xffffff))
							})
							f.Text(title).FontSize(13).Weight(vivaui.WeightSemibold).MaxWidth(280).Ellipsis().Color(theme.AccentForeground)
							f.Icon(chevronDownGlyph).Size(12, 12).Color(theme.MutedForeground)
						})
						if mainMenuTitleBox.Clicked() {
							if m.mainMenuOpen && m.mainMenuFromTitle {
								m.mainMenuOpen = false
							} else {
								m.mainMenuOpen = true
								m.mainMenuFromTitle = true
							}
						}
					}
				})
				f.Row().Gap(4).Items(flex.Center).Children(func() {
					if iconButton(f, archiveOpen, "打开 JAR 或 Class 文件", openGlyph, theme, true) {
						canvas.AfterPaint(m.chooseArchive)
					}
					if iconButton(f, archiveSave, "导出反编译源码", saveGlyph, theme, m.source != "") {
						canvas.AfterPaint(m.saveSource)
					}
					if m.busy && iconButton(f, workCancel, "取消当前操作", cancelGlyph, theme, true) {
						m.cancelWork()
					}
					if iconButton(f, githubOpen, "打开 GitHub 仓库", githubGlyph, theme, true) {
						canvas.AfterPaint(m.openRepository)
					}
					if iconButton(f, themeToggle, themeLabel, themeGlyph, theme, true) {
						if m.Theme == AppearanceDark {
							m.Theme = AppearanceLight
						} else {
							m.Theme = AppearanceDark
						}
					}
				})
			})
			if m.workspace == nil || m.homeVisible {
				m.paintHome(f, theme)
			} else {
				f.Row().Grow(1).Children(func() {
					f.Row().Width(6)
					f.Column().Width(m.split.Pos).MarginSides(6, 0, 6, 0).Radius(10).Background(theme.Background).Children(func() {
						f.Row().Height(40).PadX(8).Gap(8).Items(flex.Center).Justify(flex.Between).Children(func() {
							projectTitleBox = f.Row().Height(24).PadX(4).Gap(4).Items(flex.Center).Radius(4).HoverBackground(theme.Hover).
								HitWith(projectTitle, "项目", vivaui.HitOptions{Role: vivaui.RoleButton, Hint: "项目视图", Expanded: m.projectMenuOpen, Cursor: vivaui.PointerCursorHand})
							if m.projectMenuOpen {
								projectTitleBox = projectTitleBox.Background(theme.Hover)
							}
							projectTitleBox.Children(func() {
								f.Text("项目").FontSize(13).Color(theme.AccentForeground)
								f.Icon(chevronDownGlyph).Size(12, 12).Color(theme.MutedForeground)
							})
							if projectTitleBox.Clicked() {
								m.projectMenuOpen = !m.projectMenuOpen
							}
							if tinyIconButton(f, projectSearch, "显示筛选", searchGlyph, theme, true) {
								m.filterOpen = !m.filterOpen
							}
						})
						if m.filterOpen {
							f.Row().Height(34).PadSides(0, 8, 6, 8).Items(flex.Center).Children(func() {
								f.Field(&m.filter, projectFilter, "筛选类和资源", "筛选").Grow(1)
							})
						}
						f.Divider()
						projectTreeBox = m.paintProjectTree(f, theme, treeContentWidth, treeScrollX)
					})
					splitterSlot = f.Splitter(projectSplitter, &m.split).Width(6).Mins(180, 320)
					f.Column().Grow(1).MarginSides(6, 6, 6, 0).Radius(10).Background(theme.Background).Children(func() {
						editorMenuButton = m.paintEditorHeader(f, theme)
						if m.findOpen {
							m.paintFindBar(f, theme)
							f.Divider()
						}
						switch {
						case loading:
							f.Column().Grow(1).Center().Gap(12).Children(func() {
								f.Spinner()
								f.Text(status).FontSize(13).Muted()
							})
						case m.source != "":
							f.Row().Height(36).PadX(12).Gap(8).Items(flex.Center).Background(editorBannerBackground(m.Theme)).Children(func() {
								f.Icon(infoGlyph).Size(14, 14).Color(theme.MutedForeground)
								f.Text(m.sourceInfo).FontSize(12).Color(theme.AccentForeground).Ellipsis()
							})
							f.Divider()
							switch {
							case m.secondary == nil:
								sourceSlot = f.Slot().Grow(1)
							case m.splitAxis == splitRight:
								f.Row().Grow(1).Children(func() {
									sourceSlot = f.Slot().Grow(1)
									f.Divider()
									secondarySourceSlot = f.Slot().Grow(1)
								})
							default:
								f.Column().Grow(1).Children(func() {
									sourceSlot = f.Slot().Grow(1)
									f.Divider()
									secondarySourceSlot = f.Slot().Grow(1)
								})
							}
						case m.diagnostic != "":
							f.Column().Grow(1).Justify(flex.Center).Pad(48).Gap(8).Children(func() {
								f.Text("无法显示内容").FontSize(18).Bold()
								f.Paragraph(m.diagnostic, 4).Muted()
							})
						case m.workspace == nil || m.selected < 0:
							m.paintWelcome(f, theme)
						default:
							f.Column().Grow(1).Center().Children(func() {
								f.Text("没有可显示的源代码").FontSize(16).Muted()
							})
						}
					})
				})
			}
			f.Row().Height(34).PadSides(0, 10, 0, 12).Items(flex.Center).Justify(flex.Between).Children(func() {
				f.Text(status).FontSize(13).Muted().Ellipsis()
				f.Text("本地工作区").FontSize(13).Muted()
			})
			if m.goLineOpen {
				m.paintGoLineDialog(f, theme)
			}
		})
	})
	if m.workspace != nil && !m.homeVisible {
		if rect := splitterSlot.Rect(); rect.Width > 0 {
			canvas.Fill(rect, theme.Card)
			canvas.Fill(vivaui.Rect{X: rect.X + (rect.Width-1)/2, Y: rect.Y, Width: 1, Height: rect.Height}, theme.Border)
		}
	}
	if m.source != "" && !m.busy && !m.homeVisible && sourceSlot.Rect().Width > 0 {
		m.paintSource(canvas, theme, sourceSlot.Rect())
	}
	if m.secondary != nil && !m.homeVisible && secondarySourceSlot.Rect().Width > 0 {
		m.paintSecondarySource(canvas, theme, secondarySourceSlot.Rect())
	}
	if treeScrollX != m.treeScrollX {
		canvas.AfterPaint(func() { m.treeScrollX = treeScrollX })
	}
	if treeContentWidth > m.split.Pos && m.workspace != nil && !m.homeVisible {
		rect := projectTreeBox.Rect()
		area := components.BeginScrollArea(canvas, theme, rect, vivaui.Size{Width: treeContentWidth, Height: rect.Height}).
			Offset(treeScrollX, 0).HorizontalScrollbar(projectHorizontalScrollbar, "横向滚动项目树", "")
		m.treeScrollGeometry = area.EndContent()
	}
	if m.tabsMenuOpen && len(m.documents) > 0 {
		m.paintEditorTabMenu(canvas, theme, bounds, editorMenuButton.Rect())
	}
	if m.tabsContextOpen {
		m.paintTabContextMenu(canvas, theme, bounds)
	}
	if m.projectMenuOpen {
		m.paintProjectMenu(canvas, theme, bounds, projectTitleBox.Rect())
	}
	if m.mainMenuOpen {
		anchor := mainMenuTrigger.Rect()
		if m.mainMenuFromTitle {
			anchor = mainMenuTitleBox.Rect()
		}
		m.paintMainMenu(canvas, theme, bounds, anchor)
	}
	m.tooltip.Paint(canvas, m.Theme, bounds, workspaceTooltipDelay)
}

func (m *Model) projectTreeContentWidth(canvas *vivaui.Canvas) float32 {
	if m.workspace == nil {
		return 0
	}
	query := m.filter.Text()
	expanded := treeExpandedSignature(m.workspace.Tree, m.expanded)
	if m.treeLayout.tree == m.workspace.Tree && m.treeLayout.query == query && m.treeLayout.expanded == expanded {
		return max(m.split.Pos, m.treeLayout.width)
	}
	visible := m.workspace.Tree.Visible(m.expanded)
	if query != "" {
		visible = m.workspace.Tree.VisibleQuery(query)
	}
	width := float32(0)
	for _, item := range visible {
		nodeWidth := 12 + 4 + float32(item.Depth)*20 + 16 + 4 + 18 + 4 + canvas.TextWidth(item.Node.Name, 13) + 8
		width = max(width, nodeWidth)
	}
	m.treeLayout = treeLayoutCache{tree: m.workspace.Tree, query: query, expanded: expanded, width: width}
	return max(m.split.Pos, width)
}

func treeExpandedSignature(tree *archive.Tree, expanded map[int]bool) uint64 {
	value := uint64(1469598103934665603)
	for _, node := range tree.Nodes {
		if expanded[node.ID] {
			value = (value ^ uint64(node.ID+1)) * 1099511628211
		}
	}
	return value
}

func (m *Model) paintProjectTree(f *flex.Frame, theme vivaui.Theme, contentWidth, scrollX float32) flex.Box {
	if m.workspace == nil {
		f.Column().Pad(12).Children(func() {
			f.Text("尚未打开归档").FontSize(13).Muted()
		})
		return flex.Box{}
	}
	query := m.filter.Text()
	visible := m.workspace.Tree.Visible(m.expanded)
	if query != "" {
		visible = m.workspace.Tree.VisibleQuery(query)
		if len(visible) == 0 {
			f.Column().Pad(12).Children(func() {
				f.Text("没有匹配的类或资源").FontSize(13).Muted()
			})
			return flex.Box{}
		}
	}
	start, end := 0, len(visible)
	state := m.view.ScrollState(projectTree)
	if state.Viewport > 0 {
		start = max(0, int(state.Offset/24)-2)
		end = min(len(visible), int((state.Offset+state.Viewport)/24)+3)
		start = min(start, end)
	}
	box := f.ScrollArea(projectTree).Grow(1)
	box.Children(func() {
		if start > 0 {
			f.Row().Height(float32(start) * 24).Shrink(0)
		}
		for index := start; index < end; index++ {
			m.paintProjectNode(f, theme, visible[index], contentWidth, scrollX)
		}
		if end < len(visible) {
			f.Row().Height(float32(len(visible)-end) * 24).Shrink(0)
		}
	})
	return box
}

func (m *Model) paintProjectNode(f *flex.Frame, theme vivaui.Theme, visible archive.VisibleNode, contentWidth, scrollX float32) {
	node := visible.Node
	label := node.Name
	id := projectNodes.At(node.ID)
	selected := node.Entry >= 0 && node.Entry == m.selected
	selectedBackground := theme.Accent
	hoverColor := treeHoverColor(m.Theme)
	if selected {
		hoverColor = selectedBackground
	}
	row := f.Row().Width(max(contentWidth-12, 0)).Height(24).Shrink(0).MarginSides(0, 6, 0, 6-scrollX).Radius(5).PadSides(0, 8, 0, 4+float32(visible.Depth)*20).Gap(4).Items(flex.Center).
		HoverBackground(hoverColor).
		HitWith(id, label, vivaui.HitOptions{
			Role: vivaui.RoleButton, Selected: selected,
			Expanded: m.filter.Text() != "" || m.expanded[node.ID], Expandable: m.filter.Text() == "" && len(node.Children) > 0,
			Cursor: vivaui.PointerCursorHand,
		})
	if selected {
		row = row.Background(selectedBackground)
	}
	row.Children(func() {
		if len(node.Children) > 0 && m.filter.Text() == "" {
			m.paintNodeDisclosure(f, node, theme)
		} else {
			f.Row().Size(16, 24).Shrink(0)
		}
		paintFileIcon(f, node.Path, node.Kind, m.expanded[node.ID], 16).Shrink(0).MarginSides(0, 2, 0, 0)
		f.Text(label).FontSize(13).Shrink(0).Color(theme.AccentForeground)
	})
	if row.Clicked() {
		f.Canvas().AfterPaint(func() { m.activateNode(node) })
	}
}

func (m *Model) paintNodeDisclosure(f *flex.Frame, node archive.Node, theme vivaui.Theme) {
	glyph := chevronRightGlyph
	if m.expanded[node.ID] {
		glyph = chevronDownGlyph
	}
	button := f.Row().Size(16, 24).Center().HitWith(projectDisclosures.At(node.ID), "展开或折叠 "+node.Name, vivaui.HitOptions{
		Role: vivaui.RoleButton, Expanded: m.expanded[node.ID], Expandable: true, Cursor: vivaui.PointerCursorHand,
	})
	button.Children(func() {
		f.Icon(glyph).Size(16, 16).Color(theme.MutedForeground)
	})
	if button.Clicked() {
		f.Canvas().AfterPaint(func() { m.expanded[node.ID] = !m.expanded[node.ID] })
	}
}

func (m *Model) activateNode(node archive.Node) {
	switch node.Kind {
	case archive.NodePackage:
		m.expanded[node.ID] = !m.expanded[node.ID]
	case archive.NodeClass:
		m.startDecompile(node.Entry)
	case archive.NodeResource:
		m.startResource(node.Entry)
	}
}

func projectTreeEventTarget(id vivaui.ID) bool {
	if id == projectTree || id == projectHorizontalScrollbar {
		return true
	}
	if _, ok := id.In(projectNodes); ok {
		return true
	}
	_, ok := id.In(projectDisclosures)
	return ok
}

func (m *Model) projectTreePointer(event vivaui.PointerEvent) vivaui.EventResult {
	geometry := m.treeScrollGeometry
	switch event.Kind {
	case vivaui.PointerDown:
		if event.Button != vivaui.PointerButtonPrimary || geometry.HorizontalThumb.Width == 0 {
			return vivaui.EventResult{}
		}
		m.treeScrollDragging = true
		m.treeScrollStartX = event.X
		m.treeScrollStartPos = m.treeScrollX
	case vivaui.PointerMove:
		if !m.treeScrollDragging {
			return vivaui.EventResult{}
		}
		thumbLeft := geometry.HorizontalThumb.X + event.X - m.treeScrollStartX
		m.treeScrollX = geometry.HorizontalOffsetForThumb(thumbLeft)
	case vivaui.PointerUp:
		if !m.treeScrollDragging {
			return vivaui.EventResult{}
		}
		m.treeScrollDragging = false
	default:
		return vivaui.EventResult{}
	}
	return vivaui.EventResult{Handled: true, Repaint: true}
}

func (m *Model) paintEditorHeader(f *flex.Frame, theme vivaui.Theme) flex.Box {
	if len(m.documents) == 0 {
		return flex.Box{}
	}
	selectedBackground, selectedBorder, hoverBackground := editorTabColors(m.Theme)
	var more flex.Box
	f.Row().Height(40).PadX(4).Gap(2).Items(flex.Center).Children(func() {
		for index, doc := range m.documents {
			selected := index == m.active
			hovered := f.Canvas().Hovered() == editorTabs.At(index) || f.Canvas().Hovered() == editorTabClose.At(index)
			tab := f.Row().Height(28).MinWidth(40).MaxWidth(300).Radius(6).PadSides(0, 0, 0, 10).Gap(6).Items(flex.Center).
				HoverBackground(hoverBackground).
				HitWith(editorTabs.At(index), doc.name, vivaui.HitOptions{Role: vivaui.RoleTab, Selected: selected, Hint: doc.location, Cursor: vivaui.PointerCursorHand})
			if selected {
				tab = tab.Background(selectedBackground).Stroke(1, selectedBorder)
			} else if hovered {
				tab = tab.Background(hoverBackground)
			}
			tab.Children(func() {
				kind := archive.NodeResource
				if doc.language == "java" {
					kind = archive.NodeClass
				}
				paintFileIcon(f, doc.path, kind, false, 14)
				f.Text(doc.name).FontSize(13).MaxWidth(260).Ellipsis().Color(theme.AccentForeground)
				closeID := editorTabClose.At(index)
				closeHovered := f.Canvas().Hovered() == closeID
				closePressed := f.Canvas().Pressed() == closeID
				close := f.Row().Size(20, 20).MarginSides(0, 4, 0, 0).Center().Radius(4).
					HitWith(closeID, "关闭 "+doc.name, vivaui.HitOptions{Role: vivaui.RoleButton, Hint: "关闭 " + doc.name + " (Ctrl+W)", Cursor: vivaui.PointerCursorHand})
				if f.Canvas().FocusVisible() == closeID {
					close = close.Stroke(2, theme.Ring)
				}
				close.Children(func() {
					if selected || hovered || f.Canvas().Focused() == closeID {
						closeColor := theme.MutedForeground
						if closeHovered || closePressed {
							closeColor = theme.PrimaryForeground
						}
						button := f.Row().Size(16, 16).Radius(8).Center()
						if closeHovered {
							button = button.Background(theme.Primary)
						}
						if closePressed {
							button = button.Background(theme.Primary.Mix(theme.Background, 0.16))
						}
						button.Children(func() {
							f.Icon(cancelGlyph).Size(11, 11).Color(closeColor)
						})
					}
				})
				if close.Clicked() {
					f.Canvas().AfterPaint(func() { m.closeDocument(index) })
				}
			})
			if event, ok := f.Canvas().PointerOn(editorTabs.At(index)); ok && event.Kind == vivaui.PointerDown && event.Button == vivaui.PointerButtonSecondary {
				m.activateDocument(index)
				m.tabsContextTab = index
				m.tabsContextPoint = vivaui.Point{X: event.X, Y: event.Y}
				m.tabsContextOpen = true
				m.tabsMenuOpen = false
				f.Canvas().Consume()
			}
			if tab.Clicked() {
				m.activateDocument(index)
			}
		}
		f.Spacer()
		moreHover := theme.Hover
		moreColor := tabActionsColor(m.Theme)
		if m.tabsMenuOpen {
			moreHover = toolbarPressedColor(m.Theme)
		}
		more = f.Row().Size(28, 28).Center().Radius(6).HoverBackground(moreHover).
			HitWith(editorTabsMenuOpen, "标签页操作", vivaui.HitOptions{Role: vivaui.RoleButton, Hint: "标签页操作", Expanded: m.tabsMenuOpen, Cursor: vivaui.PointerCursorHand})
		if m.tabsMenuOpen {
			more = more.Background(moreHover)
		}
		more.Children(func() {
			f.Icon(moreVerticalGlyph).Size(16, 16).Color(moreColor)
		})
		if more.Clicked() {
			m.tabsMenuOpen = !m.tabsMenuOpen
		}
	})
	f.Divider()
	return more
}

func (m *Model) paintEditorTabMenu(canvas *vivaui.Canvas, theme vivaui.Theme, viewport, anchor vivaui.Rect) {
	const rowHeight = float32(24)
	menuTheme := contextMenuTheme(theme, m.Theme)
	items := []components.ContextMenuItem{
		{ID: editorTabsCloseAll, Label: "关闭全部标签页"},
		{ID: editorTabsCloseOthers, Label: "关闭其他标签页", Disabled: len(m.documents) < 2},
		{ID: editorTabsCloseLeft, Label: "关闭左侧标签页", Disabled: m.active <= 0},
		{ID: editorTabsCloseRight, Label: "关闭右侧标签页", Disabled: m.active < 0 || m.active >= len(m.documents)-1},
	}
	width := float32(180)
	for _, item := range items {
		width = max(width, components.ContextMenuItemWidth(canvas, menuTheme, item))
	}
	size := vivaui.Size{Width: width, Height: float32(len(items))*rowHeight + 12}
	content, _ := components.PlaceOverlay(anchor, size, viewport, components.OverlayPlacement{
		Side: components.OverlayBottom, Align: components.OverlayAlignEnd, SideOffset: 6, Margin: 6,
	})
	menu := components.BeginContextMenu(canvas, menuTheme, content, components.ContextMenuOptions{
		Layer: 4, Scope: editorTabsMenu, DismissID: editorTabsMenuDismiss, Viewport: viewport, RowHeight: rowHeight,
	})
	canvas.SetInitialFocus(editorTabsCloseAll)
	current := editorTabsCloseAll
	for _, item := range items {
		if !item.Disabled && (item.ID == canvas.Focused() || item.ID == canvas.Hovered()) {
			current = item.ID
		}
	}
	menu.BeginContent()
	for _, item := range items {
		if item.ID == current {
			canvas.Rounded(menu.NextItemBounds(), 5, menuTheme.Accent)
		}
		menu.AddItem(item)
	}
	menu.EndContent()
	menu.End()
	switch {
	case canvas.Clicked(editorTabsMenuDismiss):
		m.tabsMenuOpen = false
	case canvas.Clicked(editorTabsCloseAll):
		m.closeAllDocuments()
	case canvas.Clicked(editorTabsCloseOthers):
		m.closeOtherDocuments()
	case canvas.Clicked(editorTabsCloseLeft):
		m.closeDocumentsToLeft()
	case canvas.Clicked(editorTabsCloseRight):
		m.closeDocumentsToRight()
	}
}

func (m *Model) paintTabContextMenu(canvas *vivaui.Canvas, theme vivaui.Theme, viewport vivaui.Rect) {
	if m.tabsContextTab < 0 || m.tabsContextTab >= len(m.documents) {
		m.tabsContextOpen = false
		return
	}
	const rowHeight = float32(24)
	menuTheme := contextMenuTheme(theme, m.Theme)
	items := []components.ContextMenuItem{
		{ID: editorTabCloseCurrent, Label: "关闭", Shortcut: "Ctrl+W"},
		{ID: editorTabsCloseOthers, Label: "关闭其他标签页", Disabled: len(m.documents) < 2},
		{ID: editorTabsCloseAll, Label: "关闭全部标签页"},
		{ID: editorTabsCloseLeft, Label: "关闭左侧标签页", Disabled: m.tabsContextTab <= 0},
		{ID: editorTabsCloseRight, Label: "关闭右侧标签页", Disabled: m.tabsContextTab >= len(m.documents)-1},
		{ID: editorTabSplitRight, Label: "向右拆分", Icon: splitRightGlyph, Disabled: m.documents[m.tabsContextTab].source == ""},
		{ID: editorTabSplitDown, Label: "向下拆分", Icon: splitDownGlyph, Disabled: m.documents[m.tabsContextTab].source == ""},
	}
	width := float32(220)
	for _, item := range items {
		width = max(width, components.ContextMenuItemWidth(canvas, menuTheme, item))
	}
	size := vivaui.Size{Width: width, Height: float32(len(items))*rowHeight + 9 + 12}
	anchor := vivaui.Rect{X: m.tabsContextPoint.X, Y: m.tabsContextPoint.Y}
	content, _ := components.PlaceOverlay(anchor, size, viewport, components.OverlayPlacement{
		Side: components.OverlayBottom, Align: components.OverlayAlignStart, SideOffset: 4, Margin: 6,
	})
	menu := components.BeginContextMenu(canvas, menuTheme, content, components.ContextMenuOptions{
		Layer: 4, Scope: editorTabContextMenu, DismissID: editorTabContextDismiss, Viewport: viewport, RowHeight: rowHeight,
	})
	canvas.SetInitialFocus(editorTabCloseCurrent)
	for index, item := range items {
		if index == 5 {
			menu.AddSeparator()
		}
		menu.AddItem(item)
	}
	menu.End()
	switch {
	case canvas.Clicked(editorTabContextDismiss):
		m.tabsContextOpen = false
	case canvas.Clicked(editorTabCloseCurrent):
		index := m.tabsContextTab
		m.tabsContextOpen = false
		m.closeDocument(index)
	case canvas.Clicked(editorTabsCloseAll):
		m.closeAllDocuments()
		m.tabsContextOpen = false
	case canvas.Clicked(editorTabsCloseOthers):
		m.closeOtherDocuments()
		m.tabsContextOpen = false
	case canvas.Clicked(editorTabsCloseLeft):
		m.closeDocumentsToLeft()
		m.tabsContextOpen = false
	case canvas.Clicked(editorTabsCloseRight):
		m.closeDocumentsToRight()
		m.tabsContextOpen = false
	case canvas.Clicked(editorTabSplitRight):
		m.splitDocument(m.tabsContextTab, splitRight)
	case canvas.Clicked(editorTabSplitDown):
		m.splitDocument(m.tabsContextTab, splitDown)
	}
}

func (m *Model) splitDocument(index int, direction splitDirection) {
	if index < 0 || index >= len(m.documents) || m.documents[index].source == "" {
		return
	}
	doc := m.documents[index]
	editor := text.NewEditor(doc.source)
	if doc.editor != nil {
		editor.SetSelection(doc.editor.Selection())
	}
	m.secondary = &editorSplit{doc: doc, editor: editor}
	m.splitAxis = direction
	m.focusSplit = true
	m.tabsContextOpen = false
	if direction == splitDown {
		m.status = "已向下拆分 " + doc.name
	} else {
		m.status = "已向右拆分 " + doc.name
	}
}

func (m *Model) paintProjectMenu(canvas *vivaui.Canvas, theme vivaui.Theme, viewport, anchor vivaui.Rect) {
	menuTheme := contextMenuTheme(theme, m.Theme)
	items := []components.ContextMenuItem{
		{ID: projectExpandAll, Label: "展开全部", Disabled: m.workspace == nil},
		{ID: projectCollapseAll, Label: "折叠全部", Disabled: m.workspace == nil},
		{ID: projectFilterMenu, Label: "筛选类和资源"},
	}
	size := vivaui.Size{Width: 180, Height: float32(len(items))*24 + 12}
	content, _ := components.PlaceOverlay(anchor, size, viewport, components.OverlayPlacement{
		Side: components.OverlayBottom, Align: components.OverlayAlignStart, SideOffset: 4, Margin: 6,
	})
	menu := components.BeginContextMenu(canvas, menuTheme, content, components.ContextMenuOptions{
		Layer: 4, Scope: projectMenu, DismissID: projectMenuDismiss, Viewport: viewport, RowHeight: 24,
	})
	canvas.SetInitialFocus(projectExpandAll)
	for _, item := range items {
		menu.AddItem(item)
	}
	menu.End()
	switch {
	case canvas.Clicked(projectMenuDismiss):
		m.projectMenuOpen = false
	case canvas.Clicked(projectExpandAll):
		for _, node := range m.workspace.Tree.Nodes {
			if len(node.Children) > 0 {
				m.expanded[node.ID] = true
			}
		}
		m.projectMenuOpen = false
	case canvas.Clicked(projectCollapseAll):
		m.expanded = make(map[int]bool)
		m.projectMenuOpen = false
	case canvas.Clicked(projectFilterMenu):
		m.filterOpen = true
		m.projectMenuOpen = false
	}
}

func (m *Model) paintMainMenu(canvas *vivaui.Canvas, theme vivaui.Theme, viewport, anchor vivaui.Rect) {
	menuTheme := contextMenuTheme(theme, m.Theme)
	currentPath := ""
	if m.workspace != nil {
		currentPath = m.workspace.Archive.Path
	}
	recent := make([]string, 0, len(m.recentProjects))
	for _, path := range m.recentProjects {
		if path != currentPath {
			recent = append(recent, path)
		}
	}
	height := float32(24 + 9 + 24 + 9 + 12)
	if currentPath != "" {
		item := components.ContextMenuItem{Label: filepath.Base(currentPath), Description: currentPath}
		height += 24 + components.ContextMenuItemHeight(menuTheme, item, 24)
	}
	if len(recent) > 0 {
		height += 24
		for _, path := range recent {
			item := components.ContextMenuItem{Label: filepath.Base(path), Description: path}
			height += components.ContextMenuItemHeight(menuTheme, item, 24)
		}
	}
	size := vivaui.Size{Width: 420, Height: height}
	content, _ := components.PlaceOverlay(anchor, size, viewport, components.OverlayPlacement{
		Side: components.OverlayBottom, Align: components.OverlayAlignStart, SideOffset: 4, Margin: 6,
	})
	m.projectMenuScroll.BarID = projectMenuScrollbar
	menu := components.BeginContextMenu(canvas, menuTheme, content, components.ContextMenuOptions{
		Layer: 4, Scope: mainMenu, DismissID: mainMenuDismiss, Viewport: viewport, RowHeight: 24, Scroll: &m.projectMenuScroll,
	})
	menu.BeginContent()
	menu.AddItem(components.ContextMenuItem{ID: mainMenuOpenFile, Label: "打开...", Shortcut: "Ctrl+O", Icon: openGlyph})
	menu.AddSeparator()
	menu.AddItem(components.ContextMenuItem{ID: mainMenuHome, Label: "JarLens Home", Icon: homeGlyph})
	menu.AddSeparator()
	if currentPath != "" {
		menu.AddLabel("当前项目", false)
		item := components.ContextMenuItem{ID: mainMenuCurrentProject, Label: filepath.Base(currentPath), Description: currentPath, Inset: true}
		paintProjectBadge(canvas, menu.AddItem(item), currentPath)
	}
	if len(recent) > 0 {
		menu.AddLabel("最近项目", false)
		for index, path := range recent {
			item := components.ContextMenuItem{ID: recentProjects.At(index), Label: filepath.Base(path), Description: path, Inset: true}
			paintProjectBadge(canvas, menu.AddItem(item), path)
		}
	}
	menu.EndContent()
	menu.End()
	switch {
	case canvas.Clicked(mainMenuDismiss):
		m.mainMenuOpen = false
	case canvas.Clicked(mainMenuOpenFile):
		m.mainMenuOpen = false
		canvas.AfterPaint(m.chooseArchive)
	case canvas.Clicked(mainMenuHome):
		m.mainMenuOpen = false
		if m.RestoreWindow != nil {
			m.RestoreWindow()
		}
		m.homeVisible = true
		m.projectMenuOpen = false
		m.tabsMenuOpen = false
		m.tabsContextOpen = false
		m.goLineOpen = false
	case canvas.Clicked(mainMenuCurrentProject):
		m.mainMenuOpen = false
		if m.homeVisible && m.MaximizeWindow != nil {
			m.MaximizeWindow()
		}
		m.homeVisible = false
	default:
		for index, path := range recent {
			if canvas.Clicked(recentProjects.At(index)) {
				m.mainMenuOpen = false
				m.startOpen(path)
				return
			}
		}
	}
}

func filepathBase(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	if index := strings.LastIndex(path, "/"); index >= 0 {
		return path[index+1:]
	}
	return path
}

func iconButton(f *flex.Frame, id vivaui.ID, label string, glyph vivaui.Icon, theme vivaui.Theme, enabled bool) bool {
	color := theme.SecondaryForeground
	if !enabled {
		color = theme.MutedForeground
		color.Alpha *= 0.55
	}
	button := f.Row().Size(30, 30).Center().Radius(8).HoverBackground(theme.Hover).
		HitWith(id, label, vivaui.HitOptions{Role: vivaui.RoleButton, Hint: label, Cursor: vivaui.PointerCursorHand, Disabled: !enabled})
	button.Children(func() {
		f.Icon(glyph).Size(18, 18).Color(color)
	})
	return enabled && button.Clicked()
}

func smallIconButton(f *flex.Frame, id vivaui.ID, label string, glyph vivaui.Icon, theme vivaui.Theme, enabled bool) bool {
	color := theme.SecondaryForeground
	if !enabled {
		color = theme.MutedForeground
		color.Alpha *= 0.55
	}
	button := f.Row().Size(24, 24).Center().Radius(4).HoverBackground(theme.Hover).
		HitWith(id, label, vivaui.HitOptions{Role: vivaui.RoleButton, Hint: label, Cursor: vivaui.PointerCursorHand, Disabled: !enabled})
	button.Children(func() {
		f.Icon(glyph).Size(16, 16).Color(color)
	})
	return enabled && button.Clicked()
}

func tinyIconButton(f *flex.Frame, id vivaui.ID, label string, glyph vivaui.Icon, theme vivaui.Theme, enabled bool) bool {
	color := theme.SecondaryForeground
	if !enabled {
		color = theme.MutedForeground
		color.Alpha *= 0.55
	}
	button := f.Row().Size(20, 20).Center().Radius(4).HoverBackground(theme.Hover).
		HitWith(id, label, vivaui.HitOptions{Role: vivaui.RoleButton, Hint: label, Cursor: vivaui.PointerCursorHand, Disabled: !enabled})
	button.Children(func() {
		f.Icon(glyph).Size(16, 16).Color(color)
	})
	return enabled && button.Clicked()
}
