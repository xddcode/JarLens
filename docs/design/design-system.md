# JarLens design system

## Direction

JarLens uses the IntelliJ IDEA way of working as its interaction model and
[OpenIDEA](https://github.com/l0ng-ai/openidea) as its primary visual reference:
compact tool windows, a project hierarchy, editor tabs, breadcrumbs, command-driven
navigation, and an always-available status bar. The current palette follows
OpenIDEA's Islands theme family.

JarLens is not affiliated with JetBrains. IntelliJ IDEA is a trademark of
JetBrains s.r.o. Layout and interaction patterns may be reproduced; proprietary
resources may not. Open-source icons or other assets require their original
license and attribution when introduced.

## Current native workspace

These images come from the real Windows amd64 application. They show the initial
shell and theme switch, not a completed decompiler workflow.

| Islands Dark | Islands Light |
| --- | --- |
| ![JarLens dark workspace on Windows](assets/workspace-dark.png) | ![JarLens light workspace on Windows](assets/workspace-light.png) |

A real JAR decompiled by the embedded Vineflower engine in the native Windows
application:

![JarLens showing Vineflower-decompiled Java source](assets/workspace-decompiled.png)

## Workbench structure

```text
Titlebar: product / workspace actions / appearance
Activity stripe + Project tool window + Editor island
Status bar: task state / workspace and engine state
```

The titlebar is a single integrated Viva `WindowCustomTitlebar`, not a native
system titlebar above an application toolbar. Its empty area moves the window.
`app.TitlebarInsets` reserves platform controls, and themed minimize, maximize,
restore, and close controls retain their native window behavior and accessible
names.

Titlebar icon actions use a 28×28 hit area, a 16-pixel glyph, a 6-pixel hover
radius, and delayed tooltips. The GitHub action opens the official repository
through the operating system; the sun/moon action switches appearance. Both expose
stable accessible names and keyboard focus rather than relying on icon recognition.

Tool actions stay monochrome for consistency. Package, class, resource, status, and
brand icons may use restrained semantic colors—for example amber packages and green
Java classes—to make the tree scannable. Color supports recognition but never
replaces labels, selection state, or accessible text.

- The activity stripe exposes real tool windows only; no decorative dead buttons.
- The project window contains the lazy archive/package/class hierarchy.
- The editor owns tabs, breadcrumbs, source, diagnostics, search, and diff.
- Bottom and right tool windows remain hidden until they have useful content.
- Panels are resizable and keyboard accessible when those interactions land.

## Islands metrics

| Element | Baseline |
| --- | --- |
| Titlebar | 40 logical pixels |
| Tool-window header | 40 |
| Editor tab bar | 38 |
| Breadcrumb bar | 30 |
| Status bar | 30 |
| Project window | 240 wide |
| Tool/list row | 24 |
| Control | 28 high |
| Island gap | 6 |
| Island radius | 10 |
| Control/row radius | 4–6 |

Views use Viva metrics and theme tokens. A new isolated constant needs a design
reason and must not become a second unofficial theme system.

## Color roles

Both built-in themes use blue `#3871E1` for primary actions and focus. Dark uses
`#191A1C` editor surfaces, `#26282C` frame surfaces, and `#2A4371` selection.
Light uses white editor surfaces, `#E9EAEE` frame surfaces, and `#D0DFFE`
selection. Foreground, muted text, borders, hover, input, success, and danger are
defined per appearance in `internal/ui/theme.go`.

The previous teal accent is superseded. Brand artwork may retain a distinct
identity, but workbench selection and focus follow the semantic blue accent.

## Themes

- **Islands Dark** is the default.
- **Islands Light** is available through the working titlebar theme switch.
- High Contrast, Solarized, and Atom One are possible future additions, not
  advertised as available.
- Every new theme must define every semantic token, use the same metrics, pass
  behavior and bounds tests, and include populated-workspace screenshots.
- Editor syntax colors will have a theme-linked but separately identifiable
  scheme when the source viewer is implemented.

Theme persistence and operating-system synchronization are planned; neither is
currently claimed as implemented.

## Typography and source

UI text uses the embedded Inter 4.1 variable font at 13 logical pixels by default,
with 12 for secondary labels and 11 for compact metadata. Source uses embedded
JetBrains Mono at 13 pixels on a character grid. Both are registered only for the
JarLens process, retain their OFL notices, and use system fallback for scripts the
fonts do not cover.

The source view is read-only and includes line numbers, current-line feedback,
selection, bracket matching, vertical and horizontal scrolling, and keyboard
navigation. A source-provenance banner names the decompiler and version. Search
marks and folding arrive with their corresponding workflows. Decompiled code is
identified as generated output; minimap and decorative effects are not required.

Splitter dragging uses a neutral one-pixel indicator rather than an accent line.
The source surface does not acquire an accent border on click; caret, current line,
and selection communicate focus. Other controls retain their keyboard focus rings.

## Motion and accessibility

Motion is brief and limited to meaningful transitions. Core reading and scrolling
do not animate gratuitously. Every action is keyboard reachable, focus is visible,
contrast targets WCAG AA, and state is expressed with text or iconography in
addition to color. Screen-reader names and high-DPI behavior are acceptance
criteria, not post-release polish.

## Design change evidence

A visual PR includes real before/after screenshots when applicable, platform and
scale, and headless behavior tests. Token changes require both built-in themes and
all affected states to be reviewed, not only an empty welcome page.
