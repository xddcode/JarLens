# ADR 0003: Use GitHub-only AI-assisted delivery

- Status: Accepted
- Date: 2026-11-05

## Context

JarLens is built with AI assistance and needs durable requirements, reviewable
changes, reproducible proof, and a low-friction open-source contribution path.

## Decision

Use GitHub Issues for acceptance criteria, ADRs for decisions, Pull Requests for
implementation and evidence, Actions for quality and release, and Releases for
immutable artifacts. `main` is protected and receives reviewed squash merges.
AI-generated changes follow the same tests, security, documentation, and license
rules as all other changes.

## Alternatives

External chat, documents, or project trackers would separate decisions from the
code. Direct-to-main AI changes would be faster locally but would remove
repeatable review and CI evidence.

## Consequences

Some repository settings must be applied by a maintainer because they are not
represented in Git. Every behavior change carries a documentation and testing
responsibility. Release automation is separated from untrusted PR execution.
