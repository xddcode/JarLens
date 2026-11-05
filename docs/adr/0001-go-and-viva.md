# ADR 0001: Build JarLens with Go and Viva

- Status: Accepted
- Date: 2026-11-05

## Context

JarLens needs a compact native desktop application with code-dense UI, direct
platform integration, and a contributor-friendly build. The team already develops
Viva and Motty and does not want Electron or an embedded WebView.

## Decision

Use Go for application and domain services and Viva for native UI. Keep the Viva
library and command pinned in `go.mod`. Use a local untracked `go.work` for
optional framework co-development rather than committing a sibling-path replace.

## Alternatives

Wails would require a WebView and frontend stack despite the existing native
framework. Fyne would duplicate the team's Viva investment. A pure-Go bytecode
decompiler is independent of GUI choice and is rejected as an initial scope.

## Consequences

The application shares Viva's platform support and maturity constraints. Native
verification is required on each release platform. A single-repository clone can
build against a pinned Viva release without machine-specific paths.
