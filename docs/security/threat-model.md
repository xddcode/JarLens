# Security threat model

## Assets

- The user's files, source code, credentials, and filesystem.
- The integrity of the JarLens workspace and exported source.
- Availability of the desktop application and host resources.
- Release artifacts and update metadata.

## Untrusted inputs

JAR, Class, DEX, APK, nested archives, decompiler output, file names, metadata,
and dragged paths are untrusted. A file extension or icon is not proof of type or
safety.

## Primary threats and controls

| Threat | Required control |
| --- | --- |
| Archive path traversal | Never use archive entry names directly as output paths; validate and confine every destination |
| Decompression bomb | Bound compressed and expanded sizes, entry count, depth, time, and memory |
| Malformed bytecode | Isolate engine work, return structured errors, and fuzz parsers |
| Engine hang or excessive output | Context cancellation, process termination, timeouts, and output limits |
| Inspected code execution | Never load classes into the application JVM or run artifact entry points |
| Stale or cross-file cache results | Content digest, engine version, options, and workspace identity in cache keys |
| Sensitive data disclosure | Local-only processing by default; minimize logs and require explicit export |
| Release tampering | Protected tags/environments, checksums, SBOM, and optional platform signing |

Sandboxing an engine is defense in depth, not a substitute for safe archive
handling. Claims about isolation must match the actual platform implementation.

## Security review triggers

A PR receives explicit security review when it changes archive extraction,
process creation, temporary files, network behavior, cache identity, release
signing, update logic, or handling of credentials and private source.
