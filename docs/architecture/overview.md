# Architecture overview

## Goals

JarLens is a native desktop application for inspecting Java artifacts and reading
decompiled source. The architecture optimizes for responsive interaction,
replaceable engines, testable behavior, and safe handling of untrusted files.

## Boundaries

```text
cmd/jarlens
    |
internal/ui ---- internal/workspace
    |                   |
    |              internal/archive
    |                   |
    +------------ internal/decompile
                        |
                   internal/engine
                        |
                 pinned Java engines
```

- `cmd/jarlens` is the thin executable entry point.
- `internal/ui` owns Viva presentation and user interaction. It does not invoke
  Java processes or parse archives directly.
- `internal/workspace` owns open artifacts, recent-workspace persistence, tabs,
  selection, and task state. The application entry injects the persistent recent
  store, so UI tests never write the user's real history.
- `internal/archive` reads JAR/Class metadata and exposes lazy hierarchy models.
- `internal/decompile` defines engine-neutral requests, results, diagnostics,
  cancellation, and task state.
- `internal/engine` contains process adapters and shared engine contract tests.

Packages are introduced when their owning behavior is implemented; empty
architectural scaffolding is not a goal.

## Localization

`internal/ui` owns the shared Simplified Chinese, Traditional Chinese, English,
Japanese, and Korean catalog and resolves interface text when painting.
`internal/workspace` persists only the selected language identifier. Source code,
archive contents, and release notes are not translated.

## Execution model

The Viva window thread owns UI state. Long-running archive and decompiler work
runs outside Paint, reports through an invalidator, and is copied to the UI model
on the window thread. Every task has a context, a terminal state, and bounded
output. Paint never starts a goroutine or blocks on a process.

## Update model

`internal/update` separates release discovery from signed-package installation.
The UI resolves a Viva Ed25519-signed feed on a background task and renders
Markdown notes. GitHub and CNB use the same signature key, but each has its own
feed and asset URLs so the selected source carries release discovery, download,
and verification without silently switching hosts. Only a registered, writable
per-user Windows installation can start a download. Progress and cancellation
cross into the model through the invalidator; Viva verifies the package length
and SHA-256 again before its helper runs the NSIS installer after the application
quits. Other installation types retain manual downloads. The release workflow
signs both feeds but never exposes its private key to the application repository.

## Engine model

Engines are built in, not dynamically loaded. An engine request carries an
immutable artifact reference, options, and a cancellation context. A result
contains source or a structured failure, diagnostics, engine identity, and engine
version. The UI never parses engine-specific console text to infer success.
Vineflower, CFR, and Procyon share the neutral contract; a persistent user setting
selects subsequent tasks, while existing documents retain their actual provenance.

A future cache key includes the class content digest, engine identity and version,
and normalized options. Cache correctness takes priority over cache cleverness.

## Error model

User-visible failures distinguish unsupported input, malformed data, unavailable
runtime, timeout, cancellation, engine failure, and output limit. Internal errors
retain causes for logs; UI messages remain actionable and do not expose private
paths unnecessarily.

## Security boundary

Archives and engine output are untrusted. JarLens does not execute inspected
code, does not extract archive paths verbatim, and does not send artifact content
to a network service by default. See [the threat model](../security/threat-model.md).
