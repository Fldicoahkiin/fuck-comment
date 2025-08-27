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

支持 Go、C/C++、Java、JavaScript、TypeScript、C#、PHP、Swift、Kotlin、Rust、Scala、Dart、Objective-C 等语言的 `//` 和 `/* */` 注释格式。

## ⚠️ 重要提醒

- 使用前请备份重要代码文件
- 建议在Git等版本控制系统下使用
- 确保对目标文件有写入权限

---

**完整文档**: https://github.com/Fldicoahkiin/fuck-comment#readme
