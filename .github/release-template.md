# 🚀 fuck-comment v{{VERSION}}

**一键删注释** - 高效的跨平台代码注释删除工具

## 📦 下载

| 平台 | 架构 | 文件名 | SHA256 |
|------|------|--------|--------|
| **Windows** | x64 | `fuck-comment-windows-amd64.exe` | `{{SHA256_WINDOWS_AMD64}}` |
| Windows | x86 | `fuck-comment-windows-386.exe` | `{{SHA256_WINDOWS_386}}` |
| Windows | ARM64 | `fuck-comment-windows-arm64.exe` | `{{SHA256_WINDOWS_ARM64}}` |
| **macOS** | Intel | `fuck-comment-darwin-amd64` | `{{SHA256_DARWIN_AMD64}}` |
| **macOS** | Apple Silicon | `fuck-comment-darwin-arm64` | `{{SHA256_DARWIN_ARM64}}` |
| **Linux** | x64 | `fuck-comment-linux-amd64` | `{{SHA256_LINUX_AMD64}}` |
| Linux | x86 | `fuck-comment-linux-386` | `{{SHA256_LINUX_386}}` |
| Linux | ARM64 | `fuck-comment-linux-arm64` | `{{SHA256_LINUX_ARM64}}` |
| Linux | ARM | `fuck-comment-linux-arm` | `{{SHA256_LINUX_ARM}}` |

## 🔍 文件校验

下载后请验证文件完整性：

```bash
# macOS/Linux
sha256sum fuck-comment-*
# 或者
shasum -a 256 fuck-comment-*

# Windows (PowerShell)
Get-FileHash fuck-comment-*.exe -Algorithm SHA256
```

## 🔧 支持的语言

支持 **137** 个文件扩展名，覆盖主流编程语言和配置文件：

### 编程语言
- **C/C++ 系列**: C, C++, C#, Objective-C
- **Java 系列**: Java, Scala, Kotlin, Groovy  
- **JavaScript 系列**: JavaScript, TypeScript, JSX, TSX, CoffeeScript
- **系统编程**: Go, Rust, Swift, Dart, Zig, D
- **脚本语言**: Python, Ruby, PHP, Perl, Lua, Tcl
- **Shell**: Bash, Zsh, Fish, PowerShell, Batch
- **函数式**: Haskell, Elm, OCaml, F#, Clojure, Lisp, Scheme
- **数据科学**: R, Julia, MATLAB
- **其他**: Fortran, Pascal, Ada, Erlang, Elixir, Nim, Crystal, Odin, Jai

### 标记语言
- **Web**: HTML, XML, SVG
- **样式**: CSS, SCSS, Sass, Less, Stylus  
- **文档**: Markdown, LaTeX, reStructuredText, AsciiDoc
- **配置**: YAML, TOML, INI, JSON
- **构建工具**: Makefile, CMake, Gradle, SBT, Bazel, Dockerfile
- **DevOps**: Terraform, HCL, Nomad, Consul, Vault

### 注释格式
- `//` 行注释 (C/C++, Go, Java, JavaScript等)
- `/* */` 块注释 (C/C++, Go, Java, JavaScript等)  
- `#` 井号注释 (Python, Shell, YAML等)
- `--` 双破折号注释 (SQL, Haskell等)
- `;` 分号注释 (Assembly, Lisp等)
- `%` 百分号注释 (LaTeX, MATLAB等)
- `!` 感叹号注释 (Fortran等)
- `<!-- -->` HTML注释 (HTML, XML等)

## ⚠️ 重要提醒

- 使用前请备份重要代码文件
- 建议在Git等版本控制系统下使用
- 确保对目标文件有写入权限

---

**完整文档**: https://github.com/Fldicoahkiin/fuck-comment#readme
