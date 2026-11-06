# Release process

JarLens uses Semantic Versioning and annotated Git tags. Releases are built only
from reviewed commits on `main` through GitHub Actions.

## Distribution variants

Every platform and architecture has two variants:

- **Standard** embeds a pinned Adoptium Temurin JRE 21. It works offline without
  a system Java installation and is the default for most users.
- **Portable** uses Java 11+ from `PATH` or `JARLENS_JAVA` and is intended for
  users who manage their own runtime.

The packaging workflow verifies the JRE checksum and bundles it without changing
the application. Windows and Linux use `jre/bin` beside the executable; macOS uses
`JarLens.app/Contents/Resources/jre/bin`. Runtime downloads happen only in CI,
never at application startup.

## Prepare

1. Complete the milestone and confirm all required checks on `main`.
2. Update `CHANGELOG.md`, compatibility notes, third-party notices, and the
   version in `viva.toml`.
3. Add `docs/releases/<version>.md` with verified features, known limitations,
   engine and Viva versions, and upgrade notes.
4. Merge the release preparation PR. Do not tag an unreviewed working tree.

## Tag and build

Create an annotated tag whose message is the committed release note:

```sh
git tag -a v<version> -F docs/releases/<version>.md
git push origin v<version>
```

The release workflow rejects a tag/version mismatch or missing notes. It packages
Windows, macOS, and Linux artifacts, collects an SBOM and notices, creates a draft
GitHub Release, generates SHA-256 checksums, and verifies the artifact set before
publication.

## Signing and provenance

Unsigned builds must be labelled honestly. macOS signing/notarization and Windows
signing are enabled only when the appropriate protected credentials are
configured. Secrets belong in GitHub Environments with approval rules, not in
pull-request workflows.

The release records source and dependency versions. A published release is
immutable: fixes require a new version rather than replacing artifacts in place.

## After release

Verify downloads and native startup, update the website or documentation links
when they exist, close the milestone, and announce actual capabilities and known
limitations. A failed release remains a draft until its artifact set is repaired.
