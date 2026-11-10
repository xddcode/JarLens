# Third-party notices

JarLens uses the following third-party projects:

| Project | Use | License |
| --- | --- | --- |
| [Viva](https://github.com/vivaui/viva) | Native desktop application framework | Apache-2.0 |
| [OpenIDEA](https://github.com/l0ng-ai/openidea) | Visual and interaction reference; default file-icon mapping and SVG sources | Apache-2.0 |
| [JetBrains IntelliJ Platform](https://github.com/JetBrains/intellij-community) | Default package, class, and file-type icons compiled in `internal/ui/openideaicons` | Apache-2.0 |
| [Simple Icons](https://simpleicons.org) | GitHub mark in `internal/ui/icons.go` | CC0-1.0 |
| [Vineflower 1.12.0](https://github.com/Vineflower/vineflower) | Embedded Java decompiler engine | Apache-2.0 |
| [CFR 0.152](https://github.com/leibnitz27/cfr) | Embedded Java decompiler engine | MIT |
| [Procyon 0.6.0](https://github.com/mstrobel/procyon) | Embedded Java decompiler engine | Apache-2.0 |
| [Inter 4.1](https://github.com/rsms/inter) | Embedded interface font | SIL OFL 1.1 |
| [JetBrains Mono](https://github.com/JetBrains/JetBrainsMono) | Embedded source font | SIL OFL 1.1 |
| [Adoptium Temurin JRE 21](https://adoptium.net) | Managed Java runtime for standard packages | GPL-2.0 with Classpath Exception |

Vineflower is an unmodified Maven Central artifact with SHA-256
`1dfcfe974395734fa467ce620661c7623d05ba83670de0529b1fbd63ff548b9d`.
Its license is preserved in `internal/engine/vineflower/assets/LICENSE`, and its
source is available from the linked repository.

CFR 0.152 is an unmodified Maven Central artifact with SHA-256
`f686e8f3ded377d7bc87d216a90e9e9512df4156e75b06c655a16648ae8765b2`.
Procyon 0.6.0 is an unmodified GitHub release artifact with SHA-256
`821da96012fc69244fa1ea298c90455ee4e021434bc796d3b9546ab24601b779`.
Their licenses are preserved in `internal/engine/cfr/assets/LICENSE` and
`internal/engine/procyon/assets/LICENSE`.

The default icons are compiled from OpenIDEA's `src/ui/expui` package, class,
and file-type SVGs, including their dark variants. Their copyright and Apache-2.0
notice is preserved in `internal/ui/openideaicons/LICENSE.txt`.

JarLens is not affiliated with JetBrains. IntelliJ IDEA is a trademark of
JetBrains s.r.o. Open-source JetBrains icons, themes, and other assets retain
their copyright and license notices.

Before another engine, font, icon set, or other asset is distributed with
JarLens, its version, source, license, and required notices must be added here
and to the release's software bill of materials.
