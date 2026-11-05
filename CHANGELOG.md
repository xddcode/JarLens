# Changelog

All notable changes to JarLens are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
- Embedded Inter and JetBrains Mono fonts with OFL notices.
- Editor line numbers, current-line feedback, bracket matching, selection, and dual scrollbars.
- Repository governance, design, architecture, testing, and release documentation.
- GitHub Actions quality, security, documentation, and release workflows.
