# Security policy

## Reporting a vulnerability

Use GitHub's **private vulnerability reporting** for this repository. Do not
open a public issue with sensitive reproduction details, credentials, proprietary
archives, or exploitable crash dumps.

Include the affected commit or release, operating system, a minimal reproduction,
expected and actual behavior, and the potential impact. A maintainer will
acknowledge the report and coordinate a fix and disclosure timeline.

## Supported versions

JarLens has not published a release yet. Security fixes are applied to `main`.
Once releases exist, this document will list the supported release lines.

## Security model

JarLens treats opened JAR, Class, DEX, and APK files as untrusted data.

- Opening an archive must never execute code contained in that archive.
- Archive extraction must prevent path traversal, absolute-path writes, and
  decompression-based resource exhaustion.
- External decompiler processes need cancellation, timeouts, output limits, and
  clear ownership of temporary files.
- Network access, if introduced, must be explicit and documented; source code and
  archive contents must not be uploaded by default.
- Logs and bug reports must not contain secrets or private source content unless
  the user deliberately includes them.

Only inspect files you are legally authorized to analyze. Decompiled output is an
approximation and must not be treated as the original source or as a security
proof.
