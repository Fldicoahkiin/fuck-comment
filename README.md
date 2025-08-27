# fuck-comment

**一键删注释** - 代码注释删除工具

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Platform](https://img.shields.io/badge/Platform-Windows%20|%20macOS%20|%20Linux-lightgrey)](https://github.com/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## 📖 简介

`fuck-comment` 是一个高效的跨平台CLI工具，专门用于删除代码文件中的所有注释。支持 `//` 和 `/* */` 两种注释格式，适用于多种主流编程语言。

### 🔧 支持的编程语言

| 语言 | 扩展名 | 注释格式 |
|------|--------|----------|
| Go | `.go` | `//` `/* */` |
| C/C++ | `.c` `.cpp` `.cc` `.cxx` `.h` `.hpp` | `//` `/* */` |
| Java | `.java` | `//` `/* */` |
| JavaScript | `.js` `.jsx` | `//` `/* */` |
| TypeScript | `.ts` `.tsx` | `//` `/* */` |
| C# | `.cs` | `//` `/* */` |
| PHP | `.php` | `//` `/* */` |
| Swift | `.swift` | `//` `/* */` |
| Kotlin | `.kt` | `//` `/* */` |
| Rust | `.rs` | `//` `/* */` |
| Scala | `.scala` | `//` `/* */` |
| Dart | `.dart` | `//` `/* */` |
| Objective-C | `.m` `.mm` | `//` `/* */` |

## 🚀 快速开始

### 安装方式

#### 方式一：下载预编译版本

从 [Releases](https://github.com/Fldicoahkiin/fuck-comment/releases) 页面下载对应平台的可执行文件：

```bash
# macOS (Intel)
wget https://github.com/Fldicoahkiin/fuck-comment/releases/download/v1.0.0/fuck-comment-darwin-amd64

# macOS (Apple Silicon)
wget https://github.com/Fldicoahkiin/fuck-comment/releases/download/v1.0.0/fuck-comment-darwin-arm64

# Linux (x64)
wget https://github.com/Fldicoahkiin/fuck-comment/releases/download/v1.0.0/fuck-comment-linux-amd64

# Windows (x64)
# 下载 fuck-comment-windows-amd64.exe
```

#### 方式二：源码编译

```bash
# 克隆仓库
git clone https://github.com/Fldicoahkiin/fuck-comment.git
cd fuck-comment

# 编译
make build

# 或者直接使用go build
go build -o fuck-comment .
```

#### 方式三：Docker运行

```bash
# 构建Docker镜像
docker build -t fuck-comment .

# 使用Docker处理当前目录
docker run --rm -v $(pwd):/workspace fuck-comment

# 使用docker-compose
docker-compose run fuck-comment

# 处理指定目录
TARGET_DIR=/path/to/code docker-compose run fuck-comment-process
```

### 基本用法

```bash
# 删除当前目录及子目录所有支持文件的注释
./fuck-comment

# 显示详细处理信息
./fuck-comment -v

# 删除指定文件的注释
./fuck-comment -f main.go

# 强制模式：处理所有文件类型（不限扩展名）
./fuck-comment --force

# 查看帮助信息
./fuck-comment --help
```

## 📚 详细用法

### 命令行参数

| 参数 | 简写 | 描述 | 示例 |
|------|------|------|------|
| `--help` | `-h` | 显示帮助信息 | `fuck-comment -h` |
| `--file` | `-f` | 指定单个文件 | `fuck-comment -f main.go` |
| `--force` | | 强制模式，处理所有文件类型 | `fuck-comment --force` |
| `--verbose` | `-v` | 显示详细处理信息 | `fuck-comment -v` |

### 使用示例

#### 1. 处理整个项目

```bash
# 进入项目目录
cd /path/to/your/project

# 删除所有支持文件的注释
./fuck-comment -v
```

输出示例：
```
🚀 开始处理目录: /path/to/your/project
处理文件: ./main.go
✓ 已处理: ./main.go
处理文件: ./utils/helper.js
✓ 已处理: ./utils/helper.js
✅ 共处理了 15 个文件
```

#### 2. 处理单个文件

```bash
# 删除指定文件的注释
./fuck-comment -f src/main.cpp
```

#### 3. 强制模式处理

```bash
# 处理所有文件，不限文件类型
./fuck-comment --force -v
```

## 🔍 注释删除规则

### 支持的注释格式

1. **行注释**: `// 这是行注释`
2. **块注释**: `/* 这是块注释 */`
3. **多行块注释**:
   ```
   /*
    * 这是多行
    * 块注释
    */
   ```

### 处理示例

**处理前**:
```go
package main

import "fmt" // 导入fmt包

/*
 * 主函数
 * 程序入口点
 */
func main() {
    message := "Hello // World" // 这不是注释
    fmt.Println(message) /* 输出消息 */
}
```

**处理后**:
```go
package main

import "fmt"

func main() {
    message := "Hello // World"
    fmt.Println(message)
}
```

## 🛠️ 开发

### 环境要求

- Go 1.21 或更高版本
- Make (可选，用于构建)

### 本地开发

```bash
# 克隆仓库
git clone https://github.com/Fldicoahkiin/fuck-comment.git
cd fuck-comment

# 安装依赖
go mod tidy

# 运行
go run main.go --help

# 构建
make build

# 跨平台构建
make build-all
```

### 项目结构

```
fuck-comment/
├── main.go          # 主程序文件
├── go.mod           # Go模块文件
├── go.sum           # 依赖校验文件
├── Makefile         # 构建配置
├── build.sh         # 构建脚本
├── README.md        # 项目文档
└── dist/            # 构建输出目录
```

### 构建命令

```bash
# 本地构建
make build

# 跨平台构建
make build-all

# 清理构建文件
make clean

# 安装到系统
make install

# 运行测试
make test

# 运行所有测试（包括基准测试）
go test -v -bench=.
```

### 测试覆盖

项目包含完整的单元测试，覆盖核心功能：

- ✅ **注释删除逻辑测试** - 验证各种注释格式的正确处理
- ✅ **字符串检测测试** - 确保不会误删字符串内的注释符号
- ✅ **文件类型识别测试** - 验证支持的文件扩展名检测
- ✅ **文件处理测试** - 端到端的文件处理验证
- ✅ **性能基准测试** - 确保处理大文件时的性能表现

**性能表现**（Apple M1）：
- 注释删除：~1.8μs per operation
- 字符串检测：~81ns per operation

## ⚠️ 注意事项

1. **备份重要文件**: 使用前请备份重要代码文件
2. **版本控制**: 建议在Git等版本控制系统下使用
3. **测试环境**: 建议先在测试环境验证效果
4. **文件权限**: 确保对目标文件有写入权限
5. **字符编码**: 工具假设文件使用UTF-8编码
6. **大文件处理**: 对于超大文件，建议分批处理或使用`--verbose`监控进度

## 🤝 贡献

欢迎提交Issue和Pull Request！

### 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [Cobra](https://github.com/spf13/cobra) - 强大的CLI框架
- [Go](https://golang.org/) - 优秀的编程语言

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=Fldicoahkiin/fuck-comment&type=Date)](https://www.star-history.com/#Fldicoahkiin/fuck-comment&Date)
