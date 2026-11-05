# Required GitHub repository settings

These settings are part of the engineering baseline but live in GitHub rather
than Git. A maintainer applies and reviews them; workflow files alone cannot
enforce them.

## Main branch

- Require a pull request before merging.
- Require all CI, security, and documentation status checks.
- Require branches to be current before merging.
- Require conversation resolution.
- Restrict force pushes and deletion.
- Allow squash merging; disable or carefully control merge commits.
- Require one maintainer review once additional maintainers are active.

## Actions and security

- Default workflow token: read-only; grant write only in release jobs.
- Require approval for forked first-time contributors.
- Enable Dependabot alerts and updates.
- Enable secret scanning and push protection.
- Enable private vulnerability reporting.
- Use protected release Environments for signing credentials.

## Collaboration

- Use Issues for actionable defects and features.
- Use Discussions for questions and open-ended design discussion.
- Use Milestones for M0–M4 delivery and Projects for prioritization.
- Use Releases for versioned artifacts and immutable notes.

Changes to branch protection or release permissions are recorded in the project
operational history and reviewed by a maintainer.
