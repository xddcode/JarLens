# Compatibility and support matrix

JarLens has not published a release. This page distinguishes the current build
target from verified product support.

| Platform | Build target | Current verification |
| --- | --- | --- |
| Windows amd64 | Viva Win32 / Direct3D 11 | Real JAR opening and Vineflower source display verified locally; GitHub Actions build |
| macOS amd64/arm64 | Viva Cocoa / Metal | CI build; native product verification required before release |
| Linux amd64/arm64 | Viva X11 / Vulkan | CI build; native product verification required before release |

## Runtime and file formats

- Go 1.26.11 or newer is required to build from source.
- Viva is pinned to v0.5.1 for current native rendering, text measurement,
  input handling, and component behavior.
- JAR and standalone Class opening are implemented with bounded entry reads.
- Text resources support UTF-8, UTF-8 BOM, UTF-16LE, and UTF-16BE. XML, JSON,
  YAML, Manifest, Properties, and arbitrary valid text open without an engine.
- Vineflower 1.12.0 is embedded as an unmodified Apache-2.0 artifact.
- Standard packages include Adoptium Temurin JRE 21 and work without a system
  Java installation.
- Portable packages use Java 11+ on `PATH` or an explicit `JARLENS_JAVA` override.
- JDK is only needed to compile integration fixtures.
- DEX, APK, and AAB are not supported in the initial roadmap stages.

## Compatibility changes

A PR that changes supported operating systems, architectures, class-file
versions, Java runtimes, engines, or artifact formats updates this matrix and the
release notes. Unsupported combinations produce an explicit diagnostic rather
than an unexplained failure.
