# Contributing to JarLens

Thank you for contributing. JarLens uses GitHub Issues, Pull Requests, Actions,
Projects, and Releases as its complete collaboration workflow.

## Before making a change

1. Search existing issues and pull requests.
2. Open an issue for a feature, compatibility problem, or substantial fix.
3. Give the issue a user scenario and verifiable acceptance criteria.
4. Discuss an ADR before changing a public workflow, engine protocol, security
   boundary, or major architectural decision.

Small documentation and obvious defect fixes still need a pull request and CI.

## Development setup

Use Go 1.26 or newer. Dependencies, including the Viva command, are pinned by
the Go module. A sibling Viva checkout is not required. For local Viva
co-development, use an untracked `go.work`; do not commit a local `replace`
directive.

```sh
go run ./cmd/jarlens
go tool viva dev .
make check
make race
```

See [docs/development/setup.md](docs/development/setup.md) for platform notes.

## Pull requests

- Create a focused branch from an up-to-date `main`.
- Keep one PR scoped to one issue or coherent change.
- Use a Conventional Commit title, such as `feat: open JAR archives`.
- Link the issue and describe the actual behavior change.
- Add or update tests for success, failure, and boundary behavior.
- Update user documentation, architecture notes, and AGENTS.md when applicable.
- Attach real screenshots for visual changes and identify the platform.
- List new dependencies and their licenses.
- Do not include unrelated formatting or refactoring.

All required checks must pass. Maintainers use squash merges to keep a clear,
linear project history. Do not force-push to `main` or bypass branch protection.

## AI-assisted contributions

AI-generated changes follow exactly the same review standard as other changes.
Read [docs/development/ai-coding.md](docs/development/ai-coding.md). Never claim
a command passed without running it, weaken a test to hide a regression, or copy
code with an unknown provenance or incompatible license.

## Tests and quality

At minimum, run:

```sh
make check
make race
```

New parsers need malformed-input and fuzz coverage. Engine adapters must pass
the shared contract suite. UI changes need Viva Driver behavior tests; native
smoke or scripted E2E coverage is required when platform integration changes.

## License and attribution

By submitting a contribution, you agree that it is licensed under Apache-2.0.
Do not submit code, test archives, icons, fonts, or other assets unless their
origin and redistribution terms are known and compatible. Update
`THIRD_PARTY_NOTICES.md` when distributed third-party material changes.
