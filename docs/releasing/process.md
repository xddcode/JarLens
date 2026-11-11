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

Portable downloads additionally provide Windows `setup.exe`, macOS `dmg`, and
Linux `deb` packages beside their zip or tar.gz archives. These installers use the
same system-Java rules as the portable archives. Standard offline packages remain
zip or tar.gz so the current bundler cannot accidentally omit the managed JRE from
an installer-specific format.

## Prepare

1. Complete the milestone and confirm all required checks on `main`.
2. Update `CHANGELOG.md`, compatibility notes, third-party notices, the version
   in `viva.toml`, and `internal/update.Current`.
3. Add `docs/releases/<version>.md` with concise Chinese and English sections,
   clear Markdown headings, verified highlights, known limitations, engine and
   Viva versions, and upgrade notes. Both languages must cover the same important
   changes without repeating implementation details.
4. Merge the release preparation PR. Do not tag an unreviewed working tree.

## Tag and build

Create an annotated tag whose message is the committed release note:

```sh
git tag -a v<version> --cleanup=verbatim -F docs/releases/<version>.md
git push origin v<version>
```

`--cleanup=verbatim` preserves Markdown headings in the tag message; Git's default
comment stripping would otherwise remove every line that starts with `#`.

The release workflow rejects a tag/version mismatch or missing notes. It packages
Windows, macOS, and Linux artifacts, collects an SBOM and notices, creates a draft
GitHub Release, generates SHA-256 checksums, and verifies the artifact set before
publication.

## Automatic updates

`VIVA_UPDATE_KEY` is a required repository Actions secret containing the base64
Ed25519 private key. Only its public key is committed, in
`internal/update/install.go`. The publish job runs `viva feed` after all packages
arrive. GitHub uses `latest.json` and `latest.json.sig` with the stable
`releases/latest/download` asset URLs. CNB uses `cnb-latest.json` and
`cnb-latest.json.sig`, with assets and notes pinned to the CNB release tag. Both
feeds are verified with the application's embedded public key before upload. A
stable release contains 27 assets, including these four feed files; prereleases
do not publish stable update feeds. Windows
setup files use the Viva-compatible `jarlens-<version>-windows-<arch>-setup.exe`
name so each feed selects NSIS rather than a portable archive.

The app offers in-place updates only for registered, writable per-user Windows
installations. It verifies the signed feed and exact package bytes before running
the NSIS installer silently and restarting. Portable, development, system-wide,
and non-Windows installations keep the manual download action. Never print or
commit the private key; rotate it through the configured public-key list if it is
ever exposed.

For a prerelease, keep `[app] version` numeric because Viva's package manifest does
not accept a suffix. Tag `v<version>-beta.<n>` (or another prerelease suffix) and
commit notes under that full tag version. The workflow validates the numeric base,
uses the full tag version for notes and artifact names, and marks the GitHub
Release as a prerelease instead of latest stable.

## CNB mirror

The `Mirror to CNB` workflow pushes `main` and tags to `xddcode/JarLens` on CNB.
After the GitHub Release is published, the release workflow dispatches CNB
synchronization automatically. The job uses the pinned `mpgrm` Linux amd64 release
to synchronize metadata, notes, and all assets, including the CNB-specific feed.
The workflow verifies the tool archive with SHA-256 before running it.

Configure `CNB_USERNAME` and `CNB_TOKEN` in GitHub Actions secrets. The token needs
repository write access and Release read/write access on CNB. Release
synchronization runs on GitHub's Ubuntu runners; the Linux synchronization binary
is not a local Windows development dependency. To synchronize or retry one release
manually, run `gh workflow run sync-to-cnb.yaml -f tag=v<version>-beta.<n>`. The
manual job accepts the same pinned tag and verifies the tool checksum before use.

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
