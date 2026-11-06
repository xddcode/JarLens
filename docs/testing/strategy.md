# Testing strategy

## Principles

Tests specify observable behavior and protect compatibility. A high coverage
percentage does not compensate for missing failure paths, weak assertions, or an
untested user journey.

## Layers

### Unit tests

Cover pure archive, workspace, cache, navigation, configuration, and task logic.
Include empty, malformed, boundary, cancellation, and cleanup behavior. Tests use
temporary directories and never write to a developer's real workspace.

### Engine contract tests

Every engine adapter runs the same suite once engines are introduced. Fixtures
cover supported class-file versions, generics, nested classes, lambdas, records,
malformed bytecode, timeouts, cancellation, and engine process failures.

Golden output is tied to an explicit engine version and normalized for line
endings and environment-specific paths. Assertions also check semantic markers
so formatting changes do not become the only tested behavior.

### UI behavior tests

Use `app.NewDriver` to drive the real Viva configuration without a native window.
Test visible labels, navigation, loading, empty, failure, cancellation, tabs,
search, and diff behavior. Do not assert incidental pixel coordinates.

### Scripted E2E and native smoke

The UI driver suite exercises the complete workflow with an instrumented engine:

```text
open archive -> expand package -> select class -> decompile -> read source -> export
```

The `integration` build tag additionally compiles a real Java fixture with JDK 21
and decompiles it with the embedded Vineflower engine. `tools/runtime_test.py`
verifies runtime selection and ZIP/tar.gz bundle contents. Standard-package smoke
verification launches without a system Java executable and confirms decompilation
uses the managed runtime. Native Windows verification uses UI Automation and a
real JAR; screenshots are stored with the design evidence.

Platform integration changes also require native startup and interaction checks
on the affected operating system. Screenshots in PRs must come from the real
application.

### Fuzz, regression, and performance

Archive and class parsers require fuzz targets when introduced. Every fixed
security or compatibility defect gains a regression fixture. Benchmarks track
large archive listing, source rendering, search, cache reuse, and memory use;
shared CI runners do not enforce fragile wall-clock limits.

## Fixtures

Prefer small Java sources compiled during integration setup. If compiled fixtures
are stored, document the compiler and generation command. Third-party archives
need known redistribution terms. Never commit proprietary or unknown-origin JARs.

## Quality gates

The initial project gate is formatting, module tidiness, vet, staticcheck,
ordinary tests, race tests, and release builds. Coverage reporting begins with
M1; target at least 80% overall and 90% for core domain packages while retaining
behavior and platform tests as separate requirements.
