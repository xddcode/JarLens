# ADR 0006: Offer selectable built-in decompiler engines

- Status: Accepted
- Date: 2026-11-10

## Context

ADR 0002 established one pinned engine behind a neutral interface. Users now need
alternative output for classes whose bytecode or language constructs are handled
differently by mature engines. Switching must not invalidate existing source
provenance or require a dynamic plugin system.

## Decision

Ship Vineflower 1.12.0, CFR 0.152, and Procyon 0.6.0 as built-in engines. New
process adapters share a checked-artifact Java runner with cancellation, bounded
logs, output limits, and temporary-directory cleanup. Settings selects the engine
used by subsequent decompilation and persists its stable ID. Existing documents
retain the engine name and version that actually produced their source.

Engine artifacts are unmodified, pinned by SHA-256, and shipped with their
licenses and SBOM entries. Engine comparison, dynamic plugins, and on-demand
engine downloads remain outside this change.

## Alternatives

Downloading engines at runtime reduces package size but introduces network,
availability, and provenance failures. Dynamic plugins increase the trusted code
surface. Automatic parallel comparison is deferred because it multiplies Java
process memory use before users ask for a comparison.

## Consequences

Packages grow by the pinned engine artifacts. New engines must satisfy the shared
contract and real Class/JAR fixtures. Selection changes apply to later tasks;
already-open documents are not silently rewritten. A future cache key must include
engine identity, version, and options.
