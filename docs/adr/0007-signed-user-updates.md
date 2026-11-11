# ADR 0007: Sign and limit in-place updates

- Status: Superseded by ADR 0008
- Date: 2026-11-11

## Context

JarLens previously checked releases and opened a download page. Users expect a
complete update action, but replacing a running executable or silently installing
an unverified package would risk the existing installation and user trust. Motty
already demonstrates useful interaction boundaries: explicit acceptance, progress,
cancellation, installation-type detection, and a manual fallback.

## Decision

Reuse Viva's update client and native installation helpers rather than maintain a
second updater. Each release publishes `latest.json` and an Ed25519 detached
signature. The application embeds only the public key; the private key exists only
in the release secret. Downloads use HTTPS and must match the signed length and
SHA-256 both after download and immediately before installation.

Offer Update Now only for registered, writable per-user Windows installations.
After explicit acceptance, show progress and cancellation, run the NSIS installer
silently after the application quits, and restart it. Development, portable,
read-only, system-wide, and non-Windows installations retain the manual download
action with an explanation. A missing feed, invalid signature, changed package,
or failed launch does not quit the application or intentionally replace it.

The release workflow signs the feed only after every platform package arrives,
then includes the feed in the checksummed and remotely verified artifact set.
Windows setup names follow Viva's NSIS classifier.

## Alternatives

Downloading an installer without a signed feed would leave trust to transport and
redirects alone. Copying Motty's separate minisign updater would duplicate Viva's
existing verification and platform helpers and add another dependency. Updating
portable or development executables in place could replace unrelated files or
restart a different installation, so those cases remain manual.

## Consequences

Releases require `VIVA_UPDATE_KEY` and two additional feed assets. Existing
versions without the embedded public key continue to use manual downloads. Windows
per-user installs gain one-click updates; other platforms require their own later
signing and installation policy before in-place updates are enabled.
