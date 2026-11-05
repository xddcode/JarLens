# Compatibility and support matrix

JarLens has not published a release. This page distinguishes the current build
target from verified product support.

| Platform | Build target | Current verification |
| --- | --- | --- |
| Windows amd64 | Viva Win32 / Direct3D 11 | Real JAR opening and Vineflower source display verified locally; GitHub Actions build |
| macOS amd64/arm64 | Viva Cocoa / Metal | CI build; native product verification required before release |
| Linux amd64/arm64 | Viva X11 / Vulkan | CI build; native product verification required before release |

## Runtime and file formats

- Go 1.26 or newer is required to build from source.
- Viva is pinned to commit `f7d46f622aacbe8d0dd7ad41c55f0e03a4cc762c` for the
  character-grid editor and process-local font registration.
- JAR and standalone Class opening are implemented with bounded entry reads.
- Vineflower 1.12.0 is embedded as an unmodified Apache-2.0 artifact.
- The current development build requires a Java 11+ runtime on `PATH`; JDK is only
  needed to compile integration fixtures.
- A pinned runtime bundled into installers is planned before the 1.0 release.
- DEX, APK, and AAB are not supported in the initial roadmap stages.

## Compatibility changes

A PR that changes supported operating systems, architectures, class-file
versions, Java runtimes, engines, or artifact formats updates this matrix and the
release notes. Unsupported combinations produce an explicit diagnostic rather
than an unexplained failure.
