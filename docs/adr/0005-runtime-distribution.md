# ADR 0005: Ship standard and portable runtime distributions

- Status: Accepted
- Date: 2026-11-05

## Context

JarLens uses a Java decompiler engine. Requiring users to install Java would make
the core workflow unreliable and inconsistent across environments. Downloading a
runtime on first launch would also break offline use.

## Decision

Publish two distribution variants for every supported OS and CPU architecture:

- **Standard**: includes a pinned Adoptium Temurin JRE 21 and works without a
  system Java installation.
- **Portable**: uses a user-provided Java 11+ runtime or `JARLENS_JAVA`.

Windows and Linux packages place `jre` beside the executable. macOS packages place
it in `JarLens.app/Contents/Resources/jre`. The engine checks an explicit override
first, then those managed paths, then `PATH`.

The release workflow downloads JRE artifacts only during packaging and verifies
SHA-256 before inclusion. It does not download anything at application startup.

## Alternatives

Requiring a system JRE is smaller but hostile to new users and inconsistent.
First-launch downloads work for optional plugins but not for a core offline
decompiler. A native-image engine remains a separate feasibility investigation,
not a promise for 1.0.

## Consequences

Standard archives are larger (about 50 MB per platform) but deterministic and
usable offline. Both variants retain the same source and tests. Runtime licenses,
checksums, and platform verification are part of each release.
