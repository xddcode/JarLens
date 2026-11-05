package ui

import (
	"strings"
	"unicode/utf8"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/components"
	"github.com/vivaui/viva/flex"
	vivaui "github.com/vivaui/viva/ui"
	"github.com/xddcode/jarlens/internal/archive"
)

var (
	tree                      = vivaui.Tree{}
	workbench                 = tree.Feature("workbench")
	themeToggle               = tree.ID(workbench, "theme-toggle")
	githubOpen                = tree.ID(workbench, "github-open")
	archiveOpen               = tree.ID(workbench, "archive-open")
	archiveSave               = tree.ID(workbench, "archive-save")
	workCancel                = tree.ID(workbench, "work-cancel")
	welcomeOpen               = tree.ID(workbench, "welcome-open")
	projectTree               = tree.ID(workbench, "project-tree")
	projectSplitter           = tree.ID(workbench, "project-splitter")
	sourceEditor              = tree.ID(workbench, "source-editor")
	sourceScrollbar           = tree.ID(workbench, "source-scrollbar")
	sourceHorizontalScrollbar = tree.ID(workbench, "source-horizontal-scrollbar")
	projectNodes              = tree.Sub(workbench, tree.Name("project-nodes"))
)

func NewConfig(m *Model) app.Config {
	m.initialize()
	m.tooltip.ShowOnFocus()
	invalidator := &app.Invalidator{}
	m.invalidator = invalidator
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
				Minimize: "Minimize", Maximize: "Maximize", Restore: "Restore", CloseName: "Close",
			}
		},
		TextInput: func(context app.TextInputContext) *app.TextInput {
			if context.ID == sourceEditor && m.editor != nil {
				return &app.TextInput{Editor: m.editor}
			}
			return nil
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
			primary := event.Modifiers & (vivaui.ModifierControl | vivaui.ModifierMeta)
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
			default:
				return vivaui.EventResult{}
			}
		},
		OnKey: func(event vivaui.KeyEvent) vivaui.EventResult {
			if event.Target == projectSplitter {
				return m.split.Key(event)
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
			if event.Target == sourceEditor || event.Target == sourceScrollbar || event.Target == sourceHorizontalScrollbar {
				return m.code.Pointer(event)
			}
			return vivaui.EventResult{}
		},
		OnScroll: func(event vivaui.ScrollEvent) vivaui.EventResult {
			if event.Target == sourceEditor || event.Target == sourceScrollbar || event.Target == sourceHorizontalScrollbar {
				return m.code.Scroll(event)
			}
			return m.view.Wheel(event)
		},
	}

	config.Paint = func(canvas *vivaui.Canvas, size vivaui.Size) {
		m.paint(canvas, size)
	}
	return config
}

func (m *Model) paint(canvas *vivaui.Canvas, size vivaui.Size) {
	theme := WorkspaceTheme(m.Theme)
	bounds := vivaui.Rect{Width: size.Width, Height: size.Height}
	canvas.Fill(bounds, theme.Card)

	themeLabel := "Use light theme"
	themeGlyph := sunGlyph
	if m.Theme == AppearanceLight {
		themeLabel = "Use dark theme"
		themeGlyph = moonGlyph
	}
	status := m.status
	if status == "" {
		status = "Ready"
	}
	leading, trailing := app.TitlebarInsets()
	var sourceSlot flex.Box
	var splitterSlot flex.Box

	m.view.Draw(canvas, theme, bounds, func(f *flex.Frame) {
		f.Column().Fill().Background(theme.Card).Children(func() {
			titlebar := f.Row().Height(40).PadSides(0, trailing+8, 0, leading+12).Gap(10).Items(flex.Center).Justify(flex.Between).MovesWindow()
			titlebar.Children(func() {
				f.Row().Gap(10).Items(flex.Center).Children(func() {
					f.Text("JarLens").FontSize(14).Bold()
					f.Text("Java Decompiler").FontSize(12).Muted()
				})
				f.Row().Gap(4).Items(flex.Center).Children(func() {
					if iconButton(f, archiveOpen, "Open JAR or Class", openGlyph, theme, true) {
						canvas.AfterPaint(m.chooseArchive)
					}
					if iconButton(f, archiveSave, "Export decompiled source", saveGlyph, theme, m.source != "") {
						canvas.AfterPaint(m.saveSource)
					}
					if m.busy && iconButton(f, workCancel, "Cancel current operation", cancelGlyph, theme, true) {
						m.cancelWork()
					}
					if iconButton(f, githubOpen, "Open GitHub repository", githubGlyph, theme, true) {
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
			f.Row().Grow(1).Children(func() {
				f.Column().Width(44).PadY(6).Children(func() {
					f.Row().Height(34).MarginX(6).Radius(6).Background(theme.Accent).Center().Children(func() {
						f.Icon(projectGlyph).Size(18, 18).Color(theme.AccentForeground)
					})
				})
				f.Column().Width(m.split.Pos).MarginSides(6, 0, 6, 0).Radius(10).Background(theme.Background).Children(func() {
					f.Row().Height(40).PadX(12).Items(flex.Center).Children(func() {
						f.Text("Project").FontSize(13).Bold()
					})
					f.Divider()
					m.paintProjectTree(f, theme)
				})
				splitterSlot = f.Splitter(projectSplitter, &m.split).Mins(180, 320)
				f.Column().Grow(1).Margin(6).Radius(10).Background(theme.Background).Children(func() {
					m.paintEditorHeader(f, theme)
					f.Divider()
					switch {
					case m.busy:
						f.Column().Grow(1).Center().Gap(12).Children(func() {
							f.Spinner()
							f.Text(status).FontSize(13).Muted()
						})
					case m.source != "":
						f.Row().Height(30).PadX(12).Items(flex.Center).Background(theme.Accent).Children(func() {
							f.Text(m.sourceInfo).FontSize(12).Color(theme.AccentForeground).Ellipsis()
						})
						f.Divider()
						sourceSlot = f.Slot().Grow(1)
					case m.diagnostic != "":
						f.Column().Grow(1).Justify(flex.Center).Pad(48).Gap(8).Children(func() {
							f.Text("Unable to display content").FontSize(18).Bold()
							f.Paragraph(m.diagnostic, 4).Muted()
						})
					case m.workspace == nil:
						f.Column().Grow(1).Center().Gap(12).Children(func() {
							f.Text("Explore Java bytecode.").FontSize(20).Bold()
							f.Text("Open a JAR or Class file to start reading decompiled source.").FontSize(13).Muted()
							if f.Button(welcomeOpen, "Open JAR or Class").Clicked() {
								canvas.AfterPaint(m.chooseArchive)
							}
						})
					case m.selected < 0:
						f.Column().Grow(1).Center().Gap(8).Children(func() {
							f.Text("No file open").FontSize(20).Bold()
							f.Text("Select a class in the Project window.").FontSize(13).Muted()
						})
					default:
						f.Column().Grow(1).Center().Children(func() {
							f.Text("No source available").FontSize(16).Muted()
						})
					}
				})
			})
			f.Row().Height(30).PadX(12).Items(flex.Center).Justify(flex.Between).Children(func() {
				f.Text(status).FontSize(12).Muted().Ellipsis()
				f.Text("Local workspace").FontSize(12).Muted()
			})
		})
	})
	if rect := splitterSlot.Rect(); rect.Width > 0 {
		canvas.Fill(rect, theme.Card)
		canvas.Fill(vivaui.Rect{X: rect.X + (rect.Width-1)/2, Y: rect.Y, Width: 1, Height: rect.Height}, theme.Border)
	}
	if m.source != "" && !m.busy && sourceSlot.Rect().Width > 0 {
		m.paintSource(canvas, theme, sourceSlot.Rect())
	}
	m.tooltip.Paint(canvas, theme, bounds, components.TooltipOptions{SideOffset: 10}, components.HoverTooltipDelay)
}

func (m *Model) paintProjectTree(f *flex.Frame, theme vivaui.Theme) {
	if m.workspace == nil {
		f.Column().Pad(12).Children(func() {
			f.Text("No archive open").FontSize(13).Muted()
		})
		return
	}
	f.ScrollArea(projectTree).Grow(1).Children(func() {
		for _, visible := range m.workspace.Tree.Visible(m.expanded) {
			m.paintProjectNode(f, theme, visible)
		}
	})
}

func (m *Model) paintProjectNode(f *flex.Frame, theme vivaui.Theme, visible archive.VisibleNode) {
	node := visible.Node
	label := node.Name
	id := projectNodes.At(node.ID)
	row := f.Row().Height(24).PadSides(0, 4, 0, 8+float32(visible.Depth)*12).Gap(5).Items(flex.Center).
		HoverBackground(theme.Hover).
		HitWith(id, label, vivaui.HitOptions{
			Role: vivaui.RoleButton, Selected: node.Entry >= 0 && node.Entry == m.selected,
			Expanded: m.expanded[node.ID], Expandable: node.Kind == archive.NodePackage,
			Disabled: node.Kind == archive.NodeResource, Cursor: vivaui.PointerCursorHand,
		})
	if node.Entry >= 0 && node.Entry == m.selected {
		row = row.Background(theme.Accent)
	}
	row.Children(func() {
		switch node.Kind {
		case archive.NodePackage:
			glyph := chevronRightGlyph
			if m.expanded[node.ID] {
				glyph = chevronDownGlyph
			}
			f.Icon(glyph).Size(13, 13).Color(theme.MutedForeground)
			f.Icon(projectGlyph).Size(16, 16).Color(vivaui.RGB24(0xd5b778))
		case archive.NodeClass:
			f.Icon(classGlyph).Size(16, 16).Color(vivaui.RGB24(0x6db083))
		case archive.NodeResource:
			f.Icon(resourceGlyph).Size(16, 16).Color(theme.MutedForeground)
		}
		f.Text(label).FontSize(13).Ellipsis()
	})
	if row.Clicked() {
		f.Canvas().AfterPaint(func() { m.activateNode(node) })
	}
}

func (m *Model) activateNode(node archive.Node) {
	switch node.Kind {
	case archive.NodePackage:
		m.expanded[node.ID] = !m.expanded[node.ID]
	case archive.NodeClass:
		m.startDecompile(node.Entry)
	}
}

func (m *Model) paintEditorHeader(f *flex.Frame, _ vivaui.Theme) {
	tab := "No open files"
	breadcrumb := "Workspace"
	if m.workspace != nil {
		breadcrumb = filepathBase(m.workspace.Archive.Path)
		if m.selected >= 0 {
			entry := m.workspace.Archive.Entries[m.selected]
			tab = strings.TrimSuffix(entry.Name, ".class")
			breadcrumb = entry.Path
		}
	}
	f.Row().Height(38).PadX(12).Items(flex.Center).Children(func() {
		f.Text(tab).FontSize(13).Ellipsis()
	})
	f.Divider()
	f.Row().Height(30).PadX(12).Items(flex.Center).Children(func() {
		f.Text(breadcrumb).FontSize(12).Muted().Ellipsis()
	})
}

func filepathBase(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	if index := strings.LastIndex(path, "/"); index >= 0 {
		return path[index+1:]
	}
	return path
}

func iconButton(f *flex.Frame, id vivaui.ID, label string, glyph vivaui.Icon, theme vivaui.Theme, enabled bool) bool {
	color := theme.Foreground
	if !enabled {
		color = theme.MutedForeground
	}
	button := f.Row().Size(28, 28).Center().Radius(6).HoverBackground(theme.Hover).
		HitWith(id, label, vivaui.HitOptions{Role: vivaui.RoleButton, Hint: label, Cursor: vivaui.PointerCursorHand, Disabled: !enabled})
	button.Children(func() {
		f.Icon(glyph).Size(16, 16).Color(color)
	})
	return enabled && button.Clicked()
}
