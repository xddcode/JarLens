# JarLens roadmap

The roadmap describes direction, not release promises. GitHub Issues and
Milestones contain the actionable work and acceptance criteria.

## M0 — Engineering foundation

- Go/Viva application boundaries and a native workspace shell.
- Apache-2.0 licensing and repository health files.
- AI coding and contribution protocols.
- Quality, cross-platform, security, documentation, and release automation.
- Design tokens and UI verification standards.

## M1 — Minimum useful decompiler

- Open JAR and standalone Class files without executing their contents.
- Lazy package/class tree with clear loading and failure states.
- One pinned decompiler engine behind the engine interface.
- Read-only source viewer, diagnostics, cancellation, and source export.
- Contract fixtures and an end-to-end open/decompile/read test.

## M2 — Reading and navigation

- Class, method, and field navigation.
- Workspace and full-text search.
- Content-addressed decompilation cache.
- Multiple source tabs and persisted workspace state.
- Large archive benchmarks and memory limits.

## M3 — Comparison and release quality

- Additional engines and normalized result comparison.
- Side-by-side source diff and engine diagnostics.
- Complete Windows, macOS, and Linux packages.
- Verified release provenance, SBOM, and signing where credentials are available.

## M4 — Extended analysis

- JADX integration for Android artifacts.
- Bytecode/source correlation and richer metadata inspection.
- Accessibility, localization, and performance refinements based on user feedback.

## Explicit non-goals for the initial releases

- A new JVM decompiler implemented from scratch.
- A dynamic third-party plugin system.
- Cloud upload of source or archives.
- Executing code from inspected artifacts.
