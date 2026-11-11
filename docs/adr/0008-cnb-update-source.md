# ADR 0008: Support signed updates through GitHub and CNB

- Status: Accepted
- Date: 2026-11-11

## Context

ADR 0007 established signed in-place Windows updates through GitHub. Users can
also select CNB as their download source, but CNB's release list uses different
field names and its mirrors must not be bypassed during installation. A GitHub
feed cannot simply be fetched from CNB because its signed asset URLs still point
to GitHub.

## Decision

Parse both GitHub's `tag_name`, `draft`, and `prerelease` fields and CNB's
`tag_ref`, `is_draft`, and `is_prerelease` fields. Treat CNB's release ordering
as authoritative when `is_latest` is absent.

Publish two independently signed feeds per release: `latest.json` for GitHub and
`cnb-latest.json` for CNB. Each feed names package and notes URLs on its own
host. Pass the user's selected source into feed resolution so release discovery,
feed verification, and installer download all remain on that host. An unavailable
or unsigned mirror retains manual download rather than silently switching hosts.

The release includes 27 assets and verifies both feeds with the application's
embedded public key before publication. Existing signature verification,
per-user installation checks, progress, cancellation, and restart behavior from
ADR 0007 remain unchanged.

## Alternatives

Routing every install through GitHub would not honor the CNB choice and would
still fail for users who cannot reliably reach GitHub. Editing asset URLs inside
one feed after signing would invalidate its signature. A second signing identity
would add key-management complexity without improving mirror availability.

## Consequences

New releases provide real end-to-end CNB updates once both release assets and the
CNB mirror synchronization complete. Older releases without `cnb-latest.json`
remain manual-download only on CNB. CNB's schema is now covered by tests and a
live release-channel check, while GitHub behavior remains compatible.
