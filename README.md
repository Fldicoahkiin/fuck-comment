# fuck-comment

一键删除代码注释的命令行工具

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![CI](https://github.com/Fldicoahkiin/fuck-comment/workflows/Build%20and%20Release/badge.svg)](https://github.com/Fldicoahkiin/fuck-comment/actions)
[![Release](https://img.shields.io/github/v/release/Fldicoahkiin/fuck-comment?include_prereleases)](https://github.com/Fldicoahkiin/fuck-comment/releases)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## 简介

fuck-comment 是一个跨平台命令行工具，用于删除代码文件中的注释。支持8种注释格式，覆盖137个文件扩展名。

### 🔧 支持的编程语言

| 语言类别 | 语言 | 扩展名 | 注释格式 |
|----------|------|--------|----------|
| **C/C++家族** | C/C++ | `.c` `.cpp` `.cc` `.cxx` `.h` `.hpp` | `//` `/* */` |
| | C# | `.cs` | `//` `/* */` |
| **Java家族** | Java | `.java` | `//` `/* */` |
| | Scala | `.scala` | `//` `/* */` |
| | Kotlin | `.kt` | `//` `/* */` |
| | Groovy | `.groovy` | `//` `/* */` |
| **JavaScript家族** | JavaScript | `.js` `.jsx` `.mjs` `.cjs` | `//` `/* */` |
| | TypeScript | `.ts` `.tsx` | `//` `/* */` |
| | CoffeeScript | `.coffee` | `#` |
| **系统编程** | Go | `.go` | `//` `/* */` |
| | Rust | `.rs` | `//` `/* */` |
| | Swift | `.swift` | `//` `/* */` |
| | Dart | `.dart` | `//` `/* */` |
| | Zig | `.zig` | `//` |
| | D | `.d` | `//` `/* */` |
| **移动开发** | Objective-C | `.m` `.mm` | `//` `/* */` |
| **脚本语言** | Python | `.py` | `#` |
| | Ruby | `.rb` | `#` |
| | PHP | `.php` | `//` `/* */` `#` |
| | Perl | `.pl` `.pm` | `#` |
| | Lua | `.lua` | `--` |
| | Tcl | `.tcl` | `#` |
| **Shell脚本** | Bash/Shell | `.sh` `.bash` `.zsh` `.fish` | `#` |
| | PowerShell | `.ps1` | `#` |
| | Batch | `.bat` `.cmd` | `REM` |
| **函数式语言** | Haskell | `.hs` | `--` `{- -}` |
| | Elm | `.elm` | `--` `{- -}` |
| | OCaml | `.ml` | `(* *)` |
| | F# | `.fs` `.fsx` | `//` `(* *)` |
| | Clojure | `.clj` `.cljs` | `;` |
| | Scheme | `.scm` | `;` |
| | Lisp | `.lisp` `.lsp` | `;` |
| | Emacs Lisp | `.el` | `;` |
| **数据科学** | R | `.r` `.R` | `#` |
| | Julia | `.jl` | `#` |
| | MATLAB | `.m` | `%` |
| | Mathematica | `.nb` | `(* *)` |
| **Web技术** | HTML | `.html` `.htm` | `<!-- -->` |
| | XML | `.xml` `.svg` | `<!-- -->` |
| | Vue | `.vue` | `//` `/* */` `<!-- -->` |
| | Svelte | `.svelte` | `//` `/* */` `<!-- -->` |
| | Astro | `.astro` | `//` `/* */` `<!-- -->` |
| **CSS预处理器** | CSS | `.css` | `/* */` |
| | SCSS | `.scss` | `//` `/* */` |
| | Sass | `.sass` | `//` |
| | Less | `.less` | `//` `/* */` |
| | Stylus | `.styl` | `//` `/* */` |
| **模板引擎** | Twig | `.twig` | `{# #}` |
| | ERB | `.erb` | `<%# %>` |
| | EJS | `.ejs` | `<%# %>` |
| | Handlebars | `.hbs` | `{{! }}` |
| | Mustache | `.mustache` | `{{! }}` |
| | Pug | `.pug` | `//` |
| | Liquid | `.liquid` | `{% comment %}` |
| **配置文件** | YAML | `.yaml` `.yml` | `#` |
| | TOML | `.toml` | `#` |
| | INI | `.ini` `.cfg` `.conf` | `#` `;` |
| | JSON5 | `.json5` `.jsonc` | `//` `/* */` |
| **文档格式** | Markdown | `.md` `.markdown` `.mdx` | `<!-- -->` |
| | LaTeX | `.tex` | `%` |
| | reStructuredText | `.rst` | `..` |
| | AsciiDoc | `.asciidoc` `.adoc` | `//` |
| **数据库** | SQL | `.sql` `.plsql` `.psql` | `--` `/* */` |
| **汇编语言** | Assembly | `.asm` `.s` `.S` | `;` |
| **硬件描述** | Verilog | `.v` `.vh` `.sv` | `//` `/* */` |
| | VHDL | `.vhd` `.vhdl` | `--` |
| **游戏开发** | GDScript | `.gd` | `#` |
| | HLSL | `.hlsl` | `//` `/* */` |
| | GLSL | `.glsl` | `//` `/* */` |
| | Shader | `.shader` | `//` `/* */` |
| **其他语言** | Pascal | `.pas` `.pp` | `//` `(* *)` `{ }` |
| | Ada | `.ada` `.adb` `.ads` | `--` |
| | Fortran | `.f` `.f90` `.f95` `.for` | `!` |
| | COBOL | `.cob` `.cbl` | `*` |
| | Prolog | `.pro` | `%` `/* */` |
| | Erlang | `.erl` | `%` |
| | Elixir | `.ex` `.exs` | `#` |
| | Nim | `.nim` | `#` |
| | Crystal | `.cr` | `#` |
| | Odin | `.odin` | `//` `/* */` |
| | Jai | `.jai` | `//` `/* */` |
| **构建工具** | Makefile | `.mk` | `#` |
| | CMake | `.cmake` | `#` |
| | Gradle | `.gradle` | `//` `/* */` |
| | SBT | `.sbt` | `//` `/* */` |
| | Bazel | `.bazel` `.bzl` | `#` |
| | Dockerfile | `.dockerfile` | `#` |
| **DevOps** | Terraform | `.tf` | `#` `//` |
| | HCL | `.hcl` | `#` `//` |
| | Nomad | `.nomad` | `#` |
| | Consul | `.consul` | `#` |
| | Vault | `.vault` | `#` |

## 安装

### 下载预编译版本

从 [Releases](https://github.com/Fldicoahkiin/fuck-comment/releases) 下载对应平台的可执行文件：

### 源码编译

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

## 注释删除规则

### 支持的注释格式

- `//` 行注释 (C/C++, Go, Java, JavaScript等)
- `/* */` 块注释 (C/C++, Go, Java, JavaScript等) 
- `#` 井号注释 (Python, Shell, YAML等)
- `--` 双破折号注释 (SQL, Haskell等)
- `;` 分号注释 (Assembly, Lisp等)
- `%` 百分号注释 (LaTeX, MATLAB等)
- `!` 感叹号注释 (Fortran等)
- `<!-- -->` HTML注释 (HTML, XML等)

### 歧义扩展名智能检测

工具会自动检测以下歧义扩展名的真实文件类型：

| 扩展名 | 可能的语言 | 检测方法 |
|--------|------------|----------|
| `.m` | Objective-C / MATLAB | 检测关键字和语法特征 |
| `.r` | R语言 | 检测R语言特有函数和语法 |
| `.s` | Assembly / Scheme | 检测汇编指令或Scheme语法 |
| `.d` | D语言 | 检测D语言特有语法 |
| `.f` | Fortran | 检测Fortran语法特征 |
| `.pro` | Prolog / Qt Project | 检测语法特征 |
| `.pl` | Perl / Prolog | 检测语法特征 |
| `.pp` | Pascal / Puppet | 检测语法特征 |
| `.v` | Verilog / Vim Script | 检测硬件描述语法 |

### ⚠️ 重要说明

**Python Docstring 限制**: 工具不会删除Python的docstring（`"""文档字符串"""`），因为：
- Docstring在技术上是字符串字面量，不是注释
- 它们是Python API文档的重要组成部分
- 可通过`help()`函数和`__doc__`属性访问
- 许多文档生成工具依赖docstring

如需删除docstring，建议使用专门的Python代码格式化工具。

### 处理示例

处理前:
```go
package main
import "fmt" // 导入fmt包
/* 主函数 */
func main() {
    message := "Hello // World" // 字符串中的//不会被删除
    fmt.Println(message) /* 输出 */
}
```

处理后:
```go
package main
import "fmt"
func main() {
    message := "Hello // World"
    fmt.Println(message)
}
```

## 开发

### 环境要求

- Go 1.21+
- Make (可选)

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

### 测试

```bash
# 运行测试
go test -v

# 测试覆盖率
go test -cover
```


## 注意事项

- 使用前备份重要文件
- 建议在版本控制环境下使用
- 确保对目标文件有写入权限
- 文件需为UTF-8编码

## 贡献

欢迎提交Issue和Pull Request

## 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件。

## 致谢

- [Cobra](https://github.com/spf13/cobra)
- [Go](https://golang.org/)

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=Fldicoahkiin/fuck-comment&type=Date)](https://www.star-history.com/#Fldicoahkiin/fuck-comment&Date)
