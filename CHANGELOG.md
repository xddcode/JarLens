# Changelog

All notable changes to JarLens are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-11-10

### Added

- First stable release with Simplified Chinese, English, Japanese, and Korean interface languages, immediate switching, localized settings search, and persistent preferences.
- Selectable built-in Vineflower 1.12.0, CFR 0.152, and Procyon 0.6.0 decompiler engines with persistent settings and accurate source provenance.
- Update checks and release history from the default official source or CNB, with optional automatic checks and no automatic installation.
- System-appearance following alongside explicit dark and light themes.
- Compact 2-column, default 4-column, and relaxed 6-column source indentation presets with independently configurable indentation guides.
- Native Java executable file picker with automatic runtime detection retained.
- CNB branch, tag, release-note, and asset synchronization through GitHub Actions.
- Usage and legal notice covering lawful analysis, copyright, and prohibited cracking or piracy use.

### Changed

- Upgraded the build requirement to Go 1.26.10 for current standard-library security fixes.
- Refreshed the application icon, OpenIDEA refresh glyph, and rounded inset decompiler banner.

### Fixed

- Dismiss settings selects on outside clicks and finish source drag selection before class-link navigation.
- Keep the Java executable field visible beside the native Browse button.
- Defer native appearance and system-theme work until the window event loop has started, preventing macOS startup hangs.

## [0.1.0-beta.3] - 2026-11-10

### Added

- Simplified Chinese, English, Japanese, and Korean interface languages with immediate switching, localized settings search, and persistent preferences.
- Selectable built-in Vineflower 1.12.0, CFR 0.152, and Procyon 0.6.0 decompiler engines with persistent settings and accurate source provenance.
- Update checks and release history from the default official source or CNB, with optional automatic checks and no automatic installation.
- System-appearance following alongside explicit dark and light themes.
- Compact 2-column, default 4-column, and relaxed 6-column source indentation presets with independently configurable indentation guides.
- CNB branch, tag, release-note, and asset synchronization through GitHub Actions.

### Changed

- Upgraded the build requirement to Go 1.26.10 for current standard-library security fixes.
- Refreshed the application icon, OpenIDEA refresh glyph, and rounded inset decompiler banner.

### Fixed

- Dismiss settings selects on outside clicks and finish source drag selection before class-link navigation.

## [0.1.0-beta.2] - 2026-11-10

### Added

- Portable installer downloads for Windows (`setup.exe`), macOS (`dmg`), and Linux (`deb`), alongside the existing archives and bundled-JRE standard packages.

## [0.1.0-beta.1] - 2026-11-09

### Added

- Initial Go and Viva project structure.
- OpenIDEA Islands-style workspace with an integrated custom titlebar, repository/theme icon actions, and working dark and light themes.
- Headless workspace, theme-switch, and window-boundary tests.
- Safe JAR/Class opening with lazy package and class navigation.
- Embedded Vineflower 1.12.0 decompilation with cancellation, diagnostics, and source export.
- Java syntax highlighting for dark and light themes.
- Real JDK engine integration tests and native open/decompile/read verification.
- Draggable project/editor splitter with keyboard adjustment and a neutral drag indicator.
- Neutral source focus surface without an accent outline.
- Preserved `.class` suffixes and rounded, inset tree selection that remains visible on hover.
- Multiple document tabs with independent editor state, direct switching, close action, and Ctrl/Cmd+W.
- Nested classes grouped under their outer class; selecting one locates its declaration in the outer document without creating another tab.
- XML, JSON, YAML, Manifest, and other text resources with encoding detection, language-aware icons, and original-name export.
- Project tree filtering, source find with match navigation, and go to line.
- Standard packages with managed JRE and portable packages using system Java.
- Embedded Inter and JetBrains Mono fonts with OFL notices.
- Editor line numbers, current-line feedback, bracket matching, selection, and dual scrollbars.
- Repository governance, design, architecture, testing, and release documentation.
- GitHub Actions quality, security, documentation, and release workflows.

### Changed

- Upgraded Viva to v0.5.1, migrated hover styling to `flex.Look`, and reworked settings into an OpenIDEA-style searchable page tree with staged edits, history, per-page reset, and stable select overlays.
- Refined the workbench against OpenIDEA with compact editor tabs, rounded selected-tab fills and inset borders, hover-aware close actions, collapsible project filtering, and theme-correct source-provenance banners.
- Tab activation now restores the selected node, editor state, and matching status-bar message.
