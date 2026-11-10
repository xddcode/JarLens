# Development setup

## Requirements

- Go 1.26.10 or newer.
- Windows 10/11, macOS with Xcode command-line tools, or Linux/X11 with Viva's
  system graphics dependencies.
- Git. Make is recommended for the shared quality commands.
- Java 11+ on `PATH` for decompilation; a JDK with `javac` for integration tests.

The Go module pins Viva and its command. Vineflower 1.12.0 is embedded in the
program, so no separate engine download is needed. A separate Viva checkout is not needed.
The engine checks a managed `jre` beside the executable (or the macOS app
Resources directory), then `JARLENS_JAVA`, and finally `PATH`. Standard release
packages need no system Java; portable packages use Java 11+ on `PATH`. The
Settings dialog can override this with a specific Java executable and persists
that choice in the user configuration directory.

## Run and test

```sh
go run ./cmd/jarlens
go tool viva dev .
go test ./... -count=1
go tool viva build
```

Run the complete local gate before submitting a PR:

```sh
make check
make race
make integration
```

`make check` verifies formatting, module tidiness, vet, static analysis, and
ordinary tests. `make race` runs the race detector separately because it consumes
more time and memory. `make integration` compiles a real Java fixture and runs it
through the embedded Vineflower engine.

## Working with Viva changes

Use a local, untracked `go.work` when testing an unreleased Viva change. Do not
commit workstation-specific paths or a `replace` directive pointing beside the
repository. Record any required Viva version or commit in the PR and release
notes.

## Repository layout

- `cmd/jarlens`: executable entry point.
- `internal/ui`: Viva workspace, visual theme, and UI behavior tests.
- `docs`: architecture, design, development, testing, security, and releases.
- `.github`: issue and PR contracts, automation, and workflows.

Additional packages are created with their first real behavior, following the
boundaries in [the architecture overview](../architecture/overview.md).

## Definition of done

A change is complete when its issue acceptance criteria are met, relevant tests
and documentation are updated, required CI checks pass, and no unrelated files
or generated artifacts are included. UI changes additionally need real visual
evidence and behavior coverage.
