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
- `internal/workspace` owns open artifacts, tabs, selection, and task state.
- `internal/archive` reads JAR/Class metadata and exposes lazy hierarchy models.
- `internal/decompile` defines engine-neutral requests, results, diagnostics,
  cancellation, and task state.
- `internal/engine` contains process adapters and shared engine contract tests.

Packages are introduced when their owning behavior is implemented; empty
architectural scaffolding is not a goal.

## Execution model

The Viva window thread owns UI state. Long-running archive and decompiler work
runs outside Paint, reports through an invalidator, and is copied to the UI model
on the window thread. Every task has a context, a terminal state, and bounded
output. Paint never starts a goroutine or blocks on a process.

## Engine model

The initial engine is built in, not dynamically loaded. An engine request carries
an immutable artifact reference, options, and a cancellation context. A result
contains source or a structured failure, diagnostics, engine identity, and engine
version. The UI never parses engine-specific console text to infer success.

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
