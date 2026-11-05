# AI coding protocol

JarLens is developed with AI assistance. The repository, not a chat transcript,
is the durable source of requirements, decisions, implementation, and proof.

## Required workflow

1. Read the applicable AGENTS.md, nearby code, tests, and relevant ADRs.
2. Start from a GitHub Issue with a user scenario and acceptance criteria.
3. For substantial work, record the approach and alternatives in an ADR before
   implementation.
4. Keep changes focused. Do not mix in unrelated cleanup or formatting.
5. Add tests that fail for the original defect or missing behavior.
6. Run the same meaningful checks used by CI and report their actual results.
7. Update code, tests, documentation, and AGENTS.md together.
8. Submit evidence through a PR: tests, screenshots for UI changes, and any
   remaining limitations.

## Non-negotiable rules

- Never fabricate test, build, screenshot, review, or runtime results.
- Never delete or weaken a test merely to make a change pass.
- Never hide an unresolved failure behind a temporary workaround.
- Do not claim an unimplemented roadmap item as an existing feature.
- Do not introduce dependencies without checking availability, version,
  maintenance, security, and license.
- Do not copy code or assets with unknown provenance.
- Do not commit credentials, private archives, local paths, machine configuration,
  or generated build output.
- Prefer the smallest design that satisfies the acceptance criteria. No plugin
  systems, abstraction layers, or configuration surfaces without a current need.

## Review standard

AI-generated code receives the same human and automated review as any other
contribution. Reviewers check behavior, architecture, security, failure paths,
test quality, licensing, and documentation—not merely compilation.

Large changes should be split into independently reviewable PRs. Important design
context belongs in an ADR; implementation details belong in code and tests;
transient reasoning does not need to be committed.
