# JarLens

[![CI](https://github.com/xddcode/jarlens/actions/workflows/ci.yml/badge.svg)](https://github.com/xddcode/jarlens/actions/workflows/ci.yml)
[![Integration](https://github.com/xddcode/jarlens/actions/workflows/integration.yml/badge.svg)](https://github.com/xddcode/jarlens/actions/workflows/integration.yml)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26.11%2B-00ADD8.svg)](go.mod)

**English | [简体中文](README.zh-CN.md)**

A modern, GPU-native Java decompiler and code explorer built with Go and [Viva](https://github.com/vivaui/viva).

> JarLens is under active development, with [public releases](https://github.com/xddcode/JarLens/releases) available. The current source can open JAR/Class files, browse their structure, decompile classes with built-in Vineflower, CFR, or Procyon engines, and export Java source.

## Usage and Legal Notice

JarLens is provided solely for interoperability analysis, debugging, education, and lawful security research. It is not affiliated with, authorized by, or endorsed by the authors or copyright holders of any software analyzed with it.

JarLens does not provide, assist with, or endorse cracking, key generation, activation bypass, license circumvention, or distribution of pirated software. Do not use JarLens for these purposes.

Users are responsible for ensuring they have the legal right to inspect the relevant artifacts and for complying with copyright law, software licenses, and applicable regulations. Decompiled output may remain subject to the original software's copyright and license. Do not copy, modify, or redistribute it without the necessary rights or permission.

## Why JarLens?

- Native desktop rendering without Electron, a browser engine, or an embedded WebView.
- An OpenIDEA/IntelliJ-style workbench with working Islands Dark and Islands Light themes.
- Simplified Chinese, Traditional Chinese, English, Japanese, and Korean interface languages with a persistent language selector.
- A code-first workspace designed for reading, searching, and comparing decompiled Java.
- Replaceable decompiler engine adapters rather than a UI tied to one engine.
- Reproducible development, testing, and releases driven entirely through GitHub.

## Install

- **Standard** packages include a pinned JRE and work without Java installed.
- **Portable** packages use Java 11+ on `PATH` or the `JARLENS_JAVA` override.

Both variants use the same source, tests, and decompiler engine. Platform and
architecture-specific assets are published through GitHub Releases.

## Build from source

Requirements: Go 1.26.11 or newer, a platform supported by Viva, and a Java 11+ runtime available as `java` for decompilation.

```sh
git clone https://github.com/xddcode/jarlens.git
cd jarlens
go run ./cmd/jarlens
```

Open an artifact directly:

```sh
go run ./cmd/jarlens path/to/application.jar
```

Development and verification:

```sh
go tool viva dev .
make check
make race
make integration
go tool viva build
```

`make check` is the local quality gate. `make integration` needs a JDK with `java` and `javac`. The same checks run in GitHub Actions. See the [development guide](docs/development/setup.md) and [testing strategy](docs/testing/strategy.md).

## Project status

| Area | Status |
| --- | --- |
| Native OpenIDEA Islands workspace | Available |
| JAR/Class opening and lazy class tree | Available |
| Selectable Vineflower 1.12.0, CFR 0.152, and Procyon 0.6.0 engines | Available |
| Multiple document tabs and current source export | Available |
| XML, JSON, YAML, Manifest, and other UTF-8/UTF-16 text resources | Available |
| Project filter, source find, and go to line | Available |
| Startup restoration, close-to-tray, and per-user launch-at-login preferences | Available |
| Signed one-click updates for registered per-user Windows installations | Available |
| Headless UI and real JDK integration tests | Available |
| Global index, cache, and multi-engine diff | Planned |
| Managed Java runtime in standard packages | Available |
| Installers and published releases | Available |

The roadmap and compatibility policy are in [ROADMAP.md](ROADMAP.md) and [docs/compatibility.md](docs/compatibility.md).

## Documentation

- [Documentation index](docs/README.md)
- [Architecture](docs/architecture/overview.md)
- [Design system and real screenshots](docs/design/design-system.md)
- [AI coding protocol](docs/development/ai-coding.md)
- [Contributing](CONTRIBUTING.md)
- [Security policy](SECURITY.md)

## 简体中文

基于 Go 和 Viva 的现代化 GPU 原生 Java 反编译与代码浏览工具。

JarLens 已发布正式版本并持续开发，支持打开 JAR/Class、浏览包结构、使用内置 Vineflower、CFR 或 Procyon 引擎反编译类，以及导出 Java 源码。

### 本地运行

环境要求：Go 1.26.11 或更高版本、Viva 支持的操作系统，以及可通过 `java` 命令访问的 Java 11+ 运行时。

```sh
git clone https://github.com/xddcode/jarlens.git
cd jarlens
go run ./cmd/jarlens
```

直接打开文件：

```sh
go run ./cmd/jarlens path/to/application.jar
```

开发与质量检查：

```sh
go tool viva dev .
make check
make race
make integration
go tool viva build
```

`make integration` 需要本机 JDK 提供 `java` 和 `javac`。

### 项目原则

- 不使用 Electron、浏览器引擎或内嵌 WebView。
- 以源码阅读、搜索和引擎对比为核心体验。
- 通过适配器集成反编译引擎，避免 UI 与单一引擎绑定。
- 默认在本地处理文件，不执行被分析代码，不上传源码。
- 开发、测试和发布过程全部由 GitHub 仓库驱动。

参与贡献请阅读[贡献指南](CONTRIBUTING.md)，安全问题请遵循[安全政策](SECURITY.md)。

## License

JarLens is licensed under the [Apache License, Version 2.0](LICENSE). See [NOTICE](NOTICE) and [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) for attribution and distribution notices. The JarLens name and logo are not licensed under Apache-2.0. JarLens is not affiliated with JetBrains; IntelliJ IDEA is a trademark of JetBrains s.r.o.
