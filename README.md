# JarLens

[![CI](https://github.com/xddcode/jarlens/actions/workflows/ci.yml/badge.svg)](https://github.com/xddcode/jarlens/actions/workflows/ci.yml)
[![Integration](https://github.com/xddcode/jarlens/actions/workflows/integration.yml/badge.svg)](https://github.com/xddcode/jarlens/actions/workflows/integration.yml)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26.1%2B-00ADD8.svg)](go.mod)

**English | [简体中文](README.zh-CN.md)**

A modern, GPU-native Java decompiler and code explorer built with Go and [Viva](https://github.com/vivaui/viva).

> JarLens is under active development and has not published a release. The current source can open JAR/Class files, browse their structure, decompile classes with the embedded Vineflower engine, and export Java source.

## Why JarLens?

- Native desktop rendering without Electron, a browser engine, or an embedded WebView.
- An OpenIDEA/IntelliJ-style workbench with working Islands Dark and Islands Light themes.
- A code-first workspace designed for reading, searching, and comparing decompiled Java.
- Replaceable decompiler engine adapters rather than a UI tied to one engine.
- Reproducible development, testing, and releases driven entirely through GitHub.

## Install

- **Standard** packages include a pinned JRE and work without Java installed.
- **Portable** packages use Java 11+ on `PATH` or the `JARLENS_JAVA` override.

Both variants use the same source, tests, and decompiler engine. Platform and
architecture-specific assets are published through GitHub Releases.

## Build from source

Requirements: Go 1.26.1 or newer, a platform supported by Viva, and a Java 11+ runtime available as `java` for decompilation.

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
| Vineflower 1.12.0 decompilation | Available |
| Multiple document tabs and current source export | Available |
| XML, JSON, YAML, Manifest, and other UTF-8/UTF-16 text resources | Available |
| Project filter, source find, and go to line | Available |
| Headless UI and real JDK integration tests | Available |
| Global index, cache, and multi-engine diff | Planned |
| Managed Java runtime in standard packages | Available |
| Installers and published releases | Not yet available |

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

JarLens 正在积极开发，尚未发布正式版本。当前源码已支持打开 JAR/Class、浏览包结构、使用内置 Vineflower 引擎反编译类，以及导出 Java 源码。

### 本地运行

环境要求：Go 1.26.1 或更高版本、Viva 支持的操作系统，以及可通过 `java` 命令访问的 Java 11+ 运行时。

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
