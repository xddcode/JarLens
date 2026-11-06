# JarLens

[![CI](https://github.com/xddcode/jarlens/actions/workflows/ci.yml/badge.svg)](https://github.com/xddcode/jarlens/actions/workflows/ci.yml)
[![集成测试](https://github.com/xddcode/jarlens/actions/workflows/integration.yml/badge.svg)](https://github.com/xddcode/jarlens/actions/workflows/integration.yml)
[![许可证](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26.1%2B-00ADD8.svg)](go.mod)

**[English](README.md) | 简体中文**

JarLens 是一款基于 Go 和 [Viva](https://github.com/vivaui/viva) 构建的现代化 GPU 原生 Java 反编译与代码浏览工具。

## 为什么选择 JarLens？

- 使用原生桌面渲染，不依赖 Electron、浏览器引擎或内嵌 WebView。
- 提供 OpenIDEA/IntelliJ 风格的工作区，支持 Islands Dark 和 Islands Light 主题。
- 以反编译 Java 源码的阅读、搜索和对比为核心设计工作区。
- 通过可替换的引擎适配器集成反编译器，避免 UI 与单一引擎绑定。
- 默认在本地处理文件，不执行被分析的代码，也不上传源码。
- 开发、测试和发布流程均可复现，并由仓库自动化流程驱动。

## 安装

- **标准版**包含固定版本的 JRE，无需本机安装 Java。
- **便携版**使用 `PATH` 中的 Java 11+，也可通过 `JARLENS_JAVA` 指定 Java。

两个版本使用相同的源码、测试和反编译引擎。不同平台和架构的安装资源将通过 GitHub Releases 发布；目前尚未发布正式安装包。

## 从源码构建

环境要求：Go 1.26.1 或更高版本、Viva 支持的操作系统，以及可通过 `java` 命令访问的 Java 11+ 运行时。

```sh
git clone https://gitee.com/xddcode_admin/jar-lens.git
cd jar-lens
go run ./cmd/jarlens
```

直接打开归档或类文件：

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

`make check` 是本地质量检查入口。`make integration` 需要本机 JDK 提供 `java` 和 `javac`。相同检查也会在 GitHub Actions 中运行。详见[开发指南](docs/development/setup.md)和[测试策略](docs/testing/strategy.md)。

## 当前功能

- Home 主页支持打开文件、拖入 JAR/Class，以及快速进入最近项目。
- 从 Home 进入项目时自动展开窗口，返回 Home 时恢复之前的窗口大小和位置。
- 自动记录窗口位置、尺寸和最大化状态；下次启动可直接恢复最近打开的项目。
- 项目下拉菜单支持当前项目、最近项目和 Home 之间的切换。
- 支持 JAR/Class 打开、延迟加载的包与类树、项目筛选和横向滚动。
- 使用 Vineflower 1.12.0 反编译 Java 类，支持任务取消、错误诊断和源码导出。
- 多文档标签页保留独立的编辑器状态，支持关闭、左右拆分和上下拆分。
- 支持 Java 语法高亮、类链接跳转、源码查找和转到行。
- 支持 XML、JSON、YAML、Manifest 及其他 UTF-8/UTF-16 文本资源。
- 提供无窗口 UI 行为测试、竞态测试和真实 JDK 集成测试。

全局索引、缓存和多引擎差异对比仍在规划中，不应视为已完成功能。

路线图和兼容性政策请参阅 [ROADMAP.md](ROADMAP.md) 与[兼容性说明](docs/compatibility.md)。

## 文档

- [文档索引](docs/README.md)
- [架构概览](docs/architecture/overview.md)
- [设计规范与真实截图](docs/design/design-system.md)
- [AI 辅助开发规范](docs/development/ai-coding.md)
- [贡献指南](CONTRIBUTING.md)
- [安全政策](SECURITY.md)

## 许可证

JarLens 使用 [Apache License 2.0](LICENSE) 许可。署名和分发声明请参阅 [NOTICE](NOTICE) 与 [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)。

JarLens 名称和 Logo 不以 Apache-2.0 许可授权。JarLens 与 JetBrains 没有隶属或关联关系；IntelliJ IDEA 是 JetBrains s.r.o. 的商标。
