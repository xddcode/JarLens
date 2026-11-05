# ADR 0002: Integrate decompiler engines behind adapters

- Status: Accepted
- Date: 2026-11-05

## Context

Correct JVM decompilation requires sustained handling of bytecode and language
features. Product value comes from navigation, diagnostics, comparison, and
reading experience, not from immediately replacing mature engines.

## Decision

Integrate one pinned engine for M1 through an engine-neutral application
interface. Isolate external Java processes from Viva, with cancellation, timeout,
output limits, and structured results. Add more engines through the same contract
suite. Do not build a dynamic plugin API in the initial architecture.

## Alternatives

Direct UI-to-CLI integration is simpler initially but couples parsing, lifecycle,
and presentation. A new pure-Go decompiler gives control but creates a large,
unbounded correctness effort. Dynamic plugins are unnecessary before external
plugin requirements exist.

## Consequences

Distribution must include engine licenses, versions, runtime requirements, and
SBOM entries. Output differs between engines and versions, so golden tests are
versioned and normalized. Engine upgrades are explicit and release-visible.
