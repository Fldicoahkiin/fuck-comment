package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spf13/cobra"
)

// 颜色常量
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
)

var (
	// 版本信息，在构建时通过 ldflags 注入
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
	
	// 命令行参数
	targetFile string
	forceMode  bool
	showVersion bool
	
	// 统计信息
	processedFiles []string
	skippedFiles   []string
	
	// 安全限制
	maxFileSize = 100 * 1024 * 1024 // 100MB
	maxLineLength = 50000           // 50K字符
	
	// 备份相关
	backupTimestamp = time.Now().Format("20060102_150405")
	backupRootDir   string // 备份根目录，格式：bak/dirname_timestamp

	// 支持的编程语言文件扩展名
	supportedExtensions = map[string]bool{
		// C/C++ family
		".c":     true, // C
		".cpp":   true, // C++
		".cc":    true, // C++
		".cxx":   true, // C++
		".h":     true, // C/C++ Header
		".hpp":   true, // C++ Header
		".cs":    true, // C#
		
		// Java family
		".java":  true, // Java
		".scala": true, // Scala
		".kt":    true, // Kotlin
		".groovy": true, // Groovy
		
		// JavaScript/TypeScript family
		".js":    true, // JavaScript
		".jsx":   true, // React JSX
		".ts":    true, // TypeScript
		".tsx":   true, // TypeScript JSX
		".mjs":   true, // ES6 Module
		".cjs":   true, // CommonJS
		".coffee": true, // CoffeeScript
		
		// Systems programming
		".go":    true, // Go
		".rs":    true, // Rust
		".swift": true, // Swift
		".dart":  true, // Dart
		".zig":   true, // Zig
		".d":     true, // D
		
		// Mobile development
		".m":     true, // Objective-C/MATLAB
		".mm":    true, // Objective-C++
		
		// Scripting languages
		".py":    true, // Python
		".rb":    true, // Ruby
		".php":   true, // PHP
		".pl":    true, // Perl
		".pm":    true, // Perl Module
		".lua":   true, // Lua
		".tcl":   true, // Tcl
		
		// Shell scripting
		".sh":    true, // Shell
		".bash":  true, // Bash
		".zsh":   true, // Zsh
		".fish":  true, // Fish
		".ps1":   true, // PowerShell
		".bat":   true, // Batch
		".cmd":   true, // Command
		
		// Functional languages
		".hs":    true, // Haskell
		".elm":   true, // Elm
		".ml":    true, // OCaml
		".fs":    true, // F#
		".fsx":   true, // F# Script
		".clj":   true, // Clojure
		".cljs":  true, // ClojureScript
		".scm":   true, // Scheme
		".lisp":  true, // Lisp
		".lsp":   true, // Lisp
		".el":    true, // Emacs Lisp
		
		// Data science & analysis
		".r":     true, // R
		".R":     true, // R
		".jl":    true, // Julia
		".nb":    true, // Mathematica
		
		// Web technologies
		".html":  true, // HTML
		".htm":   true, // HTML
		".xml":   true, // XML
		".svg":   true, // SVG
		".vue":   true, // Vue
		".svelte": true, // Svelte
		".astro": true, // Astro
		
		// CSS and preprocessors
		".css":   true, // CSS
		".scss":  true, // SCSS
		".sass":  true, // Sass
		".less":  true, // Less
		".styl":  true, // Stylus
		
		// Template engines
		".twig":  true, // Twig
		".erb":   true, // ERB
		".ejs":   true, // EJS
		".hbs":   true, // Handlebars
		".mustache": true, // Mustache
		".pug":   true, // Pug
		".jade":  true, // Jade
		".liquid": true, // Liquid
		
		// Configuration files
		".yaml":  true, // YAML
		".yml":   true, // YAML
		".toml":  true, // TOML
		".ini":   true, // INI
		".cfg":   true, // Config
		".conf":  true, // Config
		".json":  true, // JSON (with comments)
		".jsonc": true, // JSON with Comments
		".json5": true, // JSON5
		
		// Documentation
		".md":    true, // Markdown
		".markdown": true, // Markdown
		".mdx":   true, // MDX
		".tex":   true, // LaTeX
		".rst":   true, // reStructuredText
		".asciidoc": true, // AsciiDoc
		".adoc":  true, // AsciiDoc
		
		// Database
		".sql":   true, // SQL
		".plsql": true, // PL/SQL
		".psql":  true, // PostgreSQL
		
		// Assembly
		".asm":   true, // Assembly
		".s":     true, // Assembly
		".S":     true, // Assembly
		
		// Hardware description
		".v":     true, // Verilog
		".vh":    true, // Verilog Header
		".sv":    true, // SystemVerilog
		".vhd":   true, // VHDL
		".vhdl":  true, // VHDL
		
		// Game development
		".gd":    true, // GDScript
		".hlsl":  true, // HLSL
		".glsl":  true, // GLSL
		".shader": true, // Shader
		
		// Other languages
		".pas":   true, // Pascal
		".pp":    true, // Pascal
		".ada":   true, // Ada
		".adb":   true, // Ada
		".ads":   true, // Ada
		".f":     true, // Fortran
		".f90":   true, // Fortran 90
		".f95":   true, // Fortran 95
		".for":   true, // Fortran
		".cob":   true, // COBOL
		".cbl":   true, // COBOL
		".pro":   true, // Prolog
		".erl":   true, // Erlang
		".ex":    true, // Elixir
		".exs":   true, // Elixir Script
		".nim":   true, // Nim
		".cr":    true, // Crystal
		".odin":  true, // Odin
		".jai":   true, // Jai
		
		// Build systems & tools
		".mk":    true, // Makefile
		".cmake": true, // CMake
		".gradle": true, // Gradle
		".sbt":   true, // SBT
		".bazel": true, // Bazel
		".bzl":   true, // Bazel
		".dockerfile": true, // Dockerfile
		
		// DevOps & Infrastructure
		".tf":    true, // Terraform
		".hcl":   true, // HCL
		".nomad": true, // Nomad
		".consul": true, // Consul
		".vault": true, // Vault
	}
)

// isBinaryFile 检测是否为二进制文件
func isBinaryFile(content []byte) bool {
	if len(content) == 0 {
		return false
	}
	
	// 检查前512字节是否包含null字节
	checkSize := 512
	if len(content) < checkSize {
		checkSize = len(content)
	}
	
	for i := 0; i < checkSize; i++ {
		if content[i] == 0 {
			return true
		}
	}
	
	// 检查是否为有效UTF-8
	return !utf8.Valid(content)
}

// 颜色输出函数
func printSuccess(format string, args ...interface{}) {
	fmt.Printf(ColorGreen+"✓ "+format+ColorReset+"\n", args...)
}

func printError(format string, args ...interface{}) {
	fmt.Printf(ColorRed+"✗ "+format+ColorReset+"\n", args...)
}

func printWarning(format string, args ...interface{}) {
	fmt.Printf(ColorYellow+"⚠ "+format+ColorReset+"\n", args...)
}

func printInfo(format string, args ...interface{}) {
	fmt.Printf(ColorBlue+"ℹ "+format+ColorReset+"\n", args...)
}

func printProcessing(format string, args ...interface{}) {
	fmt.Printf(ColorCyan+"→ "+format+ColorReset+"\n", args...)
}

func printHeader(format string, args ...interface{}) {
	fmt.Printf(ColorBold+ColorPurple+"🚀 "+format+ColorReset+"\n", args...)
}

// printSummary 显示处理结果摘要
func printSummary() {
	totalFiles := len(processedFiles) + len(skippedFiles)
	
	if totalFiles == 0 {
		fmt.Printf(ColorYellow+"未找到需要处理的文件\n"+ColorReset)
		return
	}
	
	// 简洁的统计信息
	fmt.Printf("\n")
	fmt.Printf(ColorGreen+"%d"+ColorReset+" 处理", len(processedFiles))
	if len(skippedFiles) > 0 {
		fmt.Printf(" | "+ColorYellow+"%d"+ColorReset+" 跳过", len(skippedFiles))
	}
	if backupRootDir != "" {
		fmt.Printf(" | 备份: "+ColorBlue+"%s"+ColorReset+"\n", backupRootDir)
	}
}

// isFileSafe 检查文件是否安全处理
func isFileSafe(filePath string, content []byte, force bool) error {
	// 在强制模式下，只检查二进制文件，其他限制可以绕过
	if force {
		if isBinaryFile(content) {
			return fmt.Errorf("文件 %s 是二进制文件，跳过处理", filePath)
		}
		return nil
	}
	
	// 非强制模式下的完整安全检查
	// 检查文件大小
	if len(content) > maxFileSize {
		return fmt.Errorf("文件 %s 太大 (%d bytes), 超过限制 %d bytes", filePath, len(content), maxFileSize)
	}
	
	// 检查是否为二进制文件
	if isBinaryFile(content) {
		return fmt.Errorf("文件 %s 是二进制文件，跳过处理", filePath)
	}
	
	// 检查行长度
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if len(line) > maxLineLength {
			return fmt.Errorf("文件 %s 第 %d 行太长 (%d 字符), 超过限制 %d 字符", filePath, i+1, len(line), maxLineLength)
		}
	}
	
	return nil
}

// initBackupDir 初始化备份根目录
func initBackupDir(workingDir string) {
	if backupRootDir == "" {
		dirName := filepath.Base(workingDir)
		backupRootDir = filepath.Join(workingDir, "bak", dirName+"_"+backupTimestamp)
	}
}

// createBackup 创建文件备份，保持目录结构
func createBackup(filePath, workingDir string) error {
	// 初始化备份根目录
	initBackupDir(workingDir)
	
	// 计算相对路径
	relPath, err := filepath.Rel(workingDir, filePath)
	if err != nil {
		return fmt.Errorf("计算相对路径失败: %v", err)
	}
	
	// 生成备份文件路径，保持目录结构
	backupPath := filepath.Join(backupRootDir, relPath)
	
	// 创建备份文件的目录
	backupFileDir := filepath.Dir(backupPath)
	if err := os.MkdirAll(backupFileDir, 0755); err != nil {
		return fmt.Errorf("创建备份目录失败: %v", err)
	}
	
	// 读取原文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}
	
	// 写入备份文件
	err = ioutil.WriteFile(backupPath, content, 0644)
	if err != nil {
		return fmt.Errorf("创建备份失败: %v", err)
	}
	
	return nil
}

// detectFileType 检测文件的真实类型，处理歧义扩展名
func detectFileType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	
	switch ext {
	case ".m":
		return detectMFileType(filePath)
	case ".r":
		return detectRFileType(filePath)
	case ".s":
		return detectSFileType(filePath)
	case ".d":
		return detectDFileType(filePath)
	case ".f":
		return detectFFileType(filePath)
	case ".pro":
		return detectProFileType(filePath)
	case ".pl":
		return detectPlFileType(filePath)
	case ".pp":
		return detectPpFileType(filePath)
	case ".v":
		return detectVFileType(filePath)
	case ".md", ".markdown":
		return "markdown"
	case ".yml", ".yaml":
		return "yaml"
	case ".json", ".jsonc", ".json5":
		return "json"
	case ".xml", ".html", ".htm", ".svg":
		return "xml"
	case ".css", ".scss", ".sass", ".less":
		return "css"
	case ".rs":
		return "rust"
	case ".shader", ".hlsl", ".glsl":
		return "c"
	default:
		return ext[1:] // 去掉点号
	}
}

// detectMFileType 区分 .m 文件是 Objective-C 还是 MATLAB
func detectMFileType(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "unknown"
	}
	
	const maxReadSize = 8192
	if len(content) > maxReadSize {
		content = content[:maxReadSize]
	}
	
	contentStr := strings.ToLower(string(content))
	
	// Objective-C 特征
	objcKeywords := []string{"#import", "@interface", "@implementation", "nsstring", "@property", "@synthesize", "foundation/foundation.h"}
	for _, keyword := range objcKeywords {
		if strings.Contains(contentStr, keyword) {
			return "objc"
		}
	}
	
	// MATLAB 特征
	matlabKeywords := []string{"function", "end", "clear all", "clc", "matlab"}
	matlabCount := 0
	for _, keyword := range matlabKeywords {
		if strings.Contains(contentStr, keyword) {
			matlabCount++
		}
	}
	
	if strings.Contains(contentStr, "%") && matlabCount >= 1 {
		return "matlab"
	}
	
	if matlabCount >= 2 {
		return "matlab"
	}
	
	return "unknown"
}

// detectRFileType 检测 R 语言文件
func detectRFileType(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "unknown"
	}
	
	const maxReadSize = 4096
	if len(content) > maxReadSize {
		content = content[:maxReadSize]
	}
	
	contentStr := strings.ToLower(string(content))
	
	rKeywords := []string{"library(", "<-", "data.frame", "ggplot", "install.packages", "require("}
	for _, keyword := range rKeywords {
		if strings.Contains(contentStr, keyword) {
			return "r"
		}
	}
	return "unknown"
}

// detectSFileType 区分 .s 文件类型
func detectSFileType(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "unknown"
	}
	contentStr := string(content)
	if strings.Contains(contentStr, ".section") || strings.Contains(contentStr, ".global") {
		return "assembly"
	}
	if strings.Contains(contentStr, "(define") {
		return "scheme"
	}
	return "unknown"
}

// detectDFileType 检测 D 语言文件
func detectDFileType(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "unknown"
	}
	contentStr := string(content)
	if strings.Contains(contentStr, "import std") || strings.Contains(contentStr, "void main") {
		return "d"
	}
	return "unknown"
}

// detectFFileType 检测 Fortran 文件
func detectFFileType(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "unknown"
	}
	contentStr := string(content)
	if strings.Contains(contentStr, "program") || strings.Contains(contentStr, "subroutine") {
		return "fortran"
	}
	return "unknown"
}

// detectProFileType 区分 .pro 文件类型
func detectProFileType(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "unknown"
	}
	contentStr := string(content)
	if strings.Contains(contentStr, "QT") || strings.Contains(contentStr, "TARGET") {
		return "qt"
	}
	if strings.Contains(contentStr, ":-") {
		return "prolog"
	}
	return "unknown"
}

// detectPlFileType 区分 .pl 文件类型
func detectPlFileType(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "unknown"
	}
	contentStr := string(content)
	if strings.Contains(contentStr, "#!/usr/bin/perl") || strings.Contains(contentStr, "use strict") {
		return "perl"
	}
	if strings.Contains(contentStr, ":-") {
		return "prolog"
	}
	return "unknown"
}

// detectPpFileType 区分 .pp 文件类型
func detectPpFileType(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "unknown"
	}
	contentStr := string(content)
	if strings.Contains(contentStr, "program") || strings.Contains(contentStr, "begin") {
		return "pascal"
	}
	if strings.Contains(contentStr, "class") && strings.Contains(contentStr, "puppet") {
		return "puppet"
	}
	return "unknown"
}

// detectVFileType 检测 Verilog 文件
func detectVFileType(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "unknown"
	}
	contentStr := string(content)
	if strings.Contains(contentStr, "module") || strings.Contains(contentStr, "always") {
		return "verilog"
	}
	return "unknown"
}

// CommentRule 定义注释处理规则
type CommentRule struct {
	StartPattern string
	EndPattern   string
	IsLineComment bool
	ProtectFunc  func(line string, pos int) bool // 保护函数，返回true表示不删除
}

// ProtectionContext 保护上下文结构体
type ProtectionContext struct {
	Line        string
	Pos         int
	FileType    string
	CommentStart string
}

// shouldProtectInContext 检查是否应该在特定上下文中保护注释符号
func shouldProtectInContext(line string, pos int, fileType string, commentStart string) bool {
	ctx := ProtectionContext{
		Line:         line,
		Pos:          pos,
		FileType:     fileType,
		CommentStart: commentStart,
	}
	return checkProtectionRules(ctx)
}

// checkProtectionRules 检查保护规则
func checkProtectionRules(ctx ProtectionContext) bool {
	switch ctx.FileType {
	case "c", "cpp", "cc", "cxx", "h", "hpp", "java", "javascript", "js", "typescript", "ts", "go", "rust", "rs", "php", "swift", "kotlin", "scala", "dart", "cs":
		// C风格语言的通用保护已在通用规则中处理
		break
	case "yaml", "yml":
		if ctx.CommentStart == "#" {
			beforeComment := ctx.Line[:ctx.Pos]
			// 保护字符串内的#
			if isInAnyString(ctx.Line, ctx.Pos) {
				return true
			}
			
			// 保护Shell变量展开中的#（如${VAR#pattern}）
			if strings.Contains(beforeComment, "${") {
				// 检查整行的Shell变量语法
				fullLine := ctx.Line
				openBraces := strings.Count(fullLine[:ctx.Pos], "{")
				closeBraces := strings.Count(fullLine[:ctx.Pos], "}")
				if openBraces > closeBraces {
					// 检查#后面是否有}来确认这是Shell变量语法
					afterHash := fullLine[ctx.Pos+1:]
					if strings.Contains(afterHash, "}") {
						return true
					}
				}
			}
			
			// 保护URL中的锚点
			if strings.Contains(beforeComment, "http") && strings.Contains(ctx.Line[ctx.Pos:], "#") {
				return true
			}
			
			// 保护行首注释（仅保护结构性注释）
			if strings.TrimSpace(beforeComment) == "" {
				// 检查是否为结构性注释
				comment := strings.TrimSpace(ctx.Line[ctx.Pos:])
				
				// 保护markdown风格标题 (# ## ### 等)
				if strings.HasPrefix(comment, "# #") || strings.HasPrefix(comment, "# ##") || strings.HasPrefix(comment, "# ###") ||
				   strings.HasPrefix(comment, "## ") || strings.HasPrefix(comment, "### ") {
					return true
				}
				
				// 保护结构性注释的通用模式
				if isStructuralComment(comment) {
					return true
				}
				
				// 其他行首注释不保护（普通注释）
				return false
			}
			
			// 对于行尾注释，只保护字符串内和特殊URL情况，不保护普通注释
			return false
		}
	case "css", "scss", "sass", "less":
		// CSS中保护URL和content属性中的注释符号
		if ctx.CommentStart == "/*" || ctx.CommentStart == "//" {
			// 检查是否在url()函数中
			if strings.Contains(ctx.Line[:ctx.Pos], "url(") && !strings.Contains(ctx.Line[:ctx.Pos], ")") {
				return true
			}
			// 检查是否在content属性中
			if strings.Contains(ctx.Line[:ctx.Pos], "content:") {
				return true
			}
		}
	case "html", "xml", "svg":
		// HTML/XML中保护属性值和CDATA中的注释符号
		if ctx.CommentStart == "<!--" {
			// 检查是否在CDATA中
			if strings.Contains(ctx.Line[:ctx.Pos], "<![CDATA[") && !strings.Contains(ctx.Line[:ctx.Pos], "]]>") {
				return true
			}
		}
		// 保护条件语句和不完整的语句
		if ctx.CommentStart == "//" || ctx.CommentStart == "/*" {
			beforeComment := strings.TrimSpace(ctx.Line[:ctx.Pos])
			// 保护不完整的条件语句
			if strings.Contains(beforeComment, "if ") && !strings.Contains(beforeComment, "{") {
				return true
			}
			if strings.Contains(beforeComment, "for ") && !strings.Contains(beforeComment, "{") {
				return true
			}
			if strings.Contains(beforeComment, "while ") && !strings.Contains(beforeComment, "{") {
				return true
			}
			// 保护包含 != 的语句（但不包括Rust的情况）
			if strings.Contains(beforeComment, "!=") && !strings.Contains(beforeComment, "{") && ctx.FileType != "rust" {
				return true
			}
			// Rust特殊保护
			if ctx.FileType == "rust" || ctx.FileType == "rs" {
				// 保护println!宏调用
				if strings.Contains(beforeComment, "println!") && !strings.Contains(beforeComment, ";") {
					return true
				}
				if strings.Contains(beforeComment, "use ") && !strings.Contains(beforeComment, ";") {
					return true
				}
				// 不要过度保护原始字符串外的注释
				// 只有当注释确实在字符串内部时才保护
			}
		}
	case "python", "py":
		return checkPythonProtection(ctx)
	case "shell", "bash", "zsh", "sh":
		return checkShellProtection(ctx)
	}
	// Rust特殊保护
	if ctx.FileType == "rust" || ctx.FileType == "rs" {
		beforeComment := ctx.Line[:ctx.Pos]
		// 保护println!宏调用
		if strings.Contains(beforeComment, "println!") && !strings.Contains(beforeComment, ";") {
			return true
		}
		if strings.Contains(beforeComment, "use ") && !strings.Contains(beforeComment, ";") {
			return true
		}
	}
	return false
}

// checkPythonProtection 检查Python的保护规则
func checkPythonProtection(ctx ProtectionContext) bool {
	if ctx.CommentStart == "#" {
		beforeComment := ctx.Line[:ctx.Pos]
		
		// 保护docstring中的#（仅在docstring内部）
		if strings.Contains(beforeComment, `"""`) {
			firstTriple := strings.Index(beforeComment, `"""`)
			afterFirst := beforeComment[firstTriple+3:]
			if !strings.Contains(afterFirst, `"""`) {
				return true // 在未闭合的docstring内部
			}
		}
		if strings.Contains(beforeComment, "'''") {
			firstTriple := strings.Index(beforeComment, "'''")
			afterFirst := beforeComment[firstTriple+3:]
			if !strings.Contains(afterFirst, "'''") {
				return true // 在未闭合的docstring内部
			}
		}
		
		// 保护字符串中的URL锚点（只有当#确实在字符串内部时才保护）
		if isInAnyString(ctx.Line, ctx.Pos) && strings.Contains(beforeComment, "http") {
			return true
		}
		
		// 保护Python原始字符串中的#
		if strings.Contains(beforeComment, "r\"") || strings.Contains(beforeComment, "r'") {
			quoteCount := strings.Count(beforeComment, "\"") + strings.Count(beforeComment, "'")
			if quoteCount%2 == 1 {
				return true
			}
		}
		
		// f-string处理：只保护{}内部的#，不保护字符串外的注释
		if strings.Contains(beforeComment, "f\"") || strings.Contains(beforeComment, "f'") {
			// 检查#是否在f-string的{}内部
			braceCount := 0
			inFString := false
			var stringChar byte
			
			for i := 0; i < len(beforeComment); i++ {
				char := beforeComment[i]
				if !inFString {
					if (char == '"' || char == '\'') && i > 0 && beforeComment[i-1] == 'f' {
						inFString = true
						stringChar = char
					}
				} else {
					if char == stringChar && (i == 0 || beforeComment[i-1] != '\\') {
						inFString = false
					} else if char == '{' {
						braceCount++
					} else if char == '}' {
						braceCount--
					}
				}
			}
			
			// 只有在f-string的{}内部才保护#
			return inFString && braceCount > 0
		}
	}
	return false
}

// checkShellProtection 检查Shell脚本的保护规则
func checkShellProtection(ctx ProtectionContext) bool {
	if ctx.CommentStart == "#" {
		// 保护shebang
		if ctx.Pos == 0 && strings.HasPrefix(ctx.Line, "#!") {
			return true
		}
		// 保护变量替换中的#，如 ${GITHUB_REF#refs/tags/}
		beforeComment := ctx.Line[:ctx.Pos]
		if strings.Contains(beforeComment, "${") {
			// 检查是否在变量替换的#操作符位置
			if strings.Count(beforeComment, "{") > strings.Count(beforeComment, "}") {
				return true
			}
		}
		// 保护条件语句中的#
		if strings.Contains(beforeComment, "[ ") && !strings.Contains(beforeComment, " ]") {
			return true
		}
		// 保护字符串中的URL锚点（只有当#确实在字符串内部时才保护）
		if isInAnyString(ctx.Line, ctx.Pos) && strings.Contains(beforeComment, "http") {
			return true
		}
		// 保护颜色代码（更精确的检查）
		if strings.Contains(beforeComment, "#") && len(beforeComment) >= 6 {
			// 检查是否是颜色代码格式
			lastHash := strings.LastIndex(beforeComment, "#")
			if lastHash >= 0 && lastHash < len(beforeComment)-1 {
				afterHash := beforeComment[lastHash+1:]
				if len(afterHash) >= 3 && len(afterHash) <= 6 {
					// 检查是否全为十六进制字符
					isHex := true
					for _, c := range afterHash {
						if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
							isHex = false
							break
						}
					}
					if isHex {
						return true
					}
				}
			}
		}
	}
	return false
}

// 继续原有的switch语句
func checkProtectionRulesLegacy(ctx ProtectionContext) bool {
	switch ctx.FileType {
	case "shell", "bash", "zsh", "sh":
		// Shell脚本中保护shebang和特殊变量
		if ctx.CommentStart == "#" {
			// 保护shebang
			if ctx.Pos == 0 && strings.HasPrefix(ctx.Line, "#!") {
				return true
			}
			// 保护变量替换中的#，如 ${GITHUB_REF#refs/tags/}
			beforeComment := ctx.Line[:ctx.Pos]
			if strings.Contains(beforeComment, "${") {
				// 检查是否在变量替换的#操作符位置
				if strings.Count(beforeComment, "{") > strings.Count(beforeComment, "}") {
					return true
				}
			}
			// 保护条件语句中的#
			if strings.Contains(beforeComment, "[ ") && !strings.Contains(beforeComment, " ]") {
				return true
			}
		}
	case "sql":
		// SQL中保护字符串和标识符
		if ctx.CommentStart == "--" || ctx.CommentStart == "/*" {
			// 已经通过通用字符串保护处理
		}
	case "php":
		// PHP中保护变量和URL
		if ctx.CommentStart == "//" || ctx.CommentStart == "/*" || ctx.CommentStart == "#" {
			beforeComment := ctx.Line[:ctx.Pos]
			// 保护PHP变量
			if strings.Contains(beforeComment, "$") {
				return true
			}
			// 保护URL
			if strings.Contains(beforeComment, "http") {
				return true
			}
		}
	case "ruby", "rb":
		// Ruby中保护符号和正则表达式
		if ctx.CommentStart == "#" {
			beforeComment := ctx.Line[:ctx.Pos]
			// 保护Ruby符号
			if strings.Contains(beforeComment, ":") {
				return true
			}
			// 保护正则表达式
			if strings.Contains(beforeComment, "/") && !strings.Contains(beforeComment, "\"") {
				return true
			}
		}
	case "perl", "pl":
		// Perl中保护变量和正则表达式
		if ctx.CommentStart == "#" {
			beforeComment := ctx.Line[:ctx.Pos]
			// 保护Perl变量
			if strings.Contains(beforeComment, "$") || strings.Contains(beforeComment, "@") || strings.Contains(beforeComment, "%") {
				return true
			}
			// 保护正则表达式
			if strings.Contains(beforeComment, "=~") || strings.Contains(beforeComment, "!~") {
				return true
			}
		}
	case "lua":
		// Lua中保护字符串和长注释
		if ctx.CommentStart == "--" {
			beforeComment := ctx.Line[:ctx.Pos]
			// 保护长字符串中的--
			if strings.Contains(beforeComment, "[[") && !strings.Contains(beforeComment, "]]") {
				return true
			}
		}
	case "r", "R":
		// R语言中保护赋值操作符和URL
		if ctx.CommentStart == "#" {
			beforeComment := ctx.Line[:ctx.Pos]
			// 保护赋值操作符 <-
			if strings.Contains(beforeComment, "<-") {
				return true
			}
			// 保护URL
			if strings.Contains(beforeComment, "http") {
				return true
			}
			// 保护颜色代码
			if strings.Contains(beforeComment, "#") && len(beforeComment) >= 7 {
				return true
			}
		}
	}
	return false
}

// getCommentRulesForLanguage 获取指定语言的注释规则
func getCommentRulesForLanguage(fileType string) []CommentRule {
	// C风格语言 (// 和 /* */)
	cStyleRules := []CommentRule{
		{StartPattern: "//", EndPattern: "", IsLineComment: true},
		{StartPattern: "/*", EndPattern: "*/", IsLineComment: false},
	}
	
	// 井号注释语言 (#)
	hashStyleRules := []CommentRule{
		{StartPattern: "#", EndPattern: "", IsLineComment: true},
	}
	
	// 双破折号语言 (--)
	dashStyleRules := []CommentRule{
		{StartPattern: "--", EndPattern: "", IsLineComment: true},
	}
	_ = dashStyleRules // 避免未使用变量错误
	
	switch fileType {
	case "javascript", "js", "typescript", "ts", "go":
		return cStyleRules
	case "c", "cpp", "cc", "cxx", "h", "hpp", "cs", "java", "scala", "kt", "groovy":
		return cStyleRules
	case "rust", "rs":
		return cStyleRules
	case "swift", "dart", "zig", "d":
		return cStyleRules
	case "shell", "bash", "zsh", "sh":
		return hashStyleRules
	case "python", "py":
		return hashStyleRules
	case "ruby", "rb":
		return hashStyleRules
	case "perl", "pl", "pm":
		return hashStyleRules
	case "r", "R":
		return hashStyleRules
	case "tcl":
		return hashStyleRules
	case "php":
		return []CommentRule{
			{StartPattern: "//", EndPattern: "", IsLineComment: true},
			{StartPattern: "/*", EndPattern: "*/", IsLineComment: false},
			{StartPattern: "#", EndPattern: "", IsLineComment: true},
		}
	case "lua":
		return []CommentRule{
			{StartPattern: "--", EndPattern: "", IsLineComment: true},
			{StartPattern: "--[[", EndPattern: "]]", IsLineComment: false},
		}
	case "sql", "plsql", "psql":
		return []CommentRule{
			{StartPattern: "--", EndPattern: "", IsLineComment: true},
			{StartPattern: "/*", EndPattern: "*/", IsLineComment: false},
		}
	case "haskell", "hs":
		return []CommentRule{
			{StartPattern: "--", EndPattern: "", IsLineComment: true},
			{StartPattern: "{-", EndPattern: "-}", IsLineComment: false},
		}
	case "matlab", "m":
		return []CommentRule{
			{StartPattern: "%", EndPattern: "", IsLineComment: true},
			{StartPattern: "%{", EndPattern: "%}", IsLineComment: false},
		}
	case "latex", "tex":
		return []CommentRule{
			{StartPattern: "%", EndPattern: "", IsLineComment: true},
		}
	case "assembly", "asm", "s":
		return []CommentRule{
			{StartPattern: ";", EndPattern: "", IsLineComment: true},
			{StartPattern: "#", EndPattern: "", IsLineComment: true},
			{StartPattern: "//", EndPattern: "", IsLineComment: true},
		}
	case "fortran", "f", "f90", "f95":
		return []CommentRule{
			{StartPattern: "!", EndPattern: "", IsLineComment: true},
			{StartPattern: "C", EndPattern: "", IsLineComment: true},
			{StartPattern: "c", EndPattern: "", IsLineComment: true},
		}
	case "css", "scss", "sass", "less":
		return []CommentRule{
			{StartPattern: "/*", EndPattern: "*/", IsLineComment: false},
		}
	case "xml", "html", "htm", "svg", "markdown":
		return []CommentRule{
			{StartPattern: "<!--", EndPattern: "-->", IsLineComment: false},
		}
	case "yaml", "yml":
		return hashStyleRules
	case "json", "jsonc", "json5":
		return cStyleRules
	default:
		return hashStyleRules // 默认使用井号注释
	}
}
// removeCommentsByRules 根据注释规则删除注释
func removeCommentsByRules(content string, fileType string, rules []CommentRule) string {
	lines := strings.Split(content, "\n")
	var result []string
	var inBlockComment bool
	var inMultiLineString bool
	var inBacktickString bool
	inYAMLMultiLineBlock := false
	yamlBlockIndent := 0
	var blockEndPattern string

	for _, line := range lines {
		originalLine := line
		processedLine := line
		
		// 如果是空行，直接保留
		if strings.TrimSpace(line) == "" {
			result = append(result, line)
			continue
		}
		
		// YAML多行字符串块检测
		if fileType == "yaml" || fileType == "yml" {
			trimmedLine := strings.TrimSpace(line)
			currentIndent := len(line) - len(strings.TrimLeft(line, " \t"))
			
			// 检测多行字符串块开始 (|, >, |-, >-)
			if strings.Contains(line, ": |") || strings.Contains(line, ": >") || 
			   strings.Contains(line, ": |-") || strings.Contains(line, ": >-") {
				inYAMLMultiLineBlock = true
				yamlBlockIndent = currentIndent
			} else if inYAMLMultiLineBlock {
				// 检查是否退出多行字符串块
				if trimmedLine != "" && currentIndent <= yamlBlockIndent {
					inYAMLMultiLineBlock = false
				}
			}
			
			// 如果在YAML多行字符串块中，保护所有内容
			if inYAMLMultiLineBlock {
				result = append(result, originalLine)
				continue
			}
		}
		
		// 检查多行字符串状态 - 在处理注释之前更新状态
		oldMultiLineState := inMultiLineString
		oldBacktickState := inBacktickString
		
		// 跟踪反引号字符串状态（用于Go/JS/TS模板字符串）
		if fileType == "go" || fileType == "js" || fileType == "ts" || fileType == "jsx" || fileType == "tsx" || fileType == "javascript" {
			backtickCount := 0
			for i := 0; i < len(line); i++ {
				if line[i] == '`' && !isEscaped(line, i) {
					backtickCount++
				}
			}
			if backtickCount%2 == 1 {
				inBacktickString = !inBacktickString
			}
		}
		
		// Python docstring 处理
		if fileType == "python" || fileType == "py" {
			tempInMultiLine := inMultiLineString
			singleLineDocstring := false
			
			// 检查是否有三引号
			if strings.Contains(line, `"""`) || strings.Contains(line, "'''") {
				// 检查单行docstring
				if strings.Count(line, `"""`) >= 2 || strings.Count(line, "'''") >= 2 {
					// 可能是单行docstring
					startPos := -1
					endPos := -1
					quote := ""
					
					if pos := strings.Index(line, `"""`); pos != -1 {
						startPos = pos
						quote = `"""`
					} else if pos := strings.Index(line, "'''"); pos != -1 {
						startPos = pos
						quote = "'''"
					}
					
					if startPos != -1 {
						// 查找结束位置
						endPos = strings.Index(line[startPos+3:], quote)
						if endPos != -1 {
							endPos += startPos + 3 + 3 // 加上开始位置和三引号长度
						}
						
						if endPos < len(line) {
							beforeEnd := line[:endPos]
							afterEnd := line[endPos:]
							// 删除docstring后的注释
							if pos := strings.Index(afterEnd, "#"); pos != -1 {
								afterEnd = strings.TrimRight(afterEnd[:pos], " \t")
							}
							processedLine = beforeEnd + afterEnd
							singleLineDocstring = true
						} else {
							// 单行docstring占据整行，不影响多行状态
							singleLineDocstring = true
						}
					}
				}
				
				if !singleLineDocstring {
					// 计算不在字符串内的三引号数量
					count := 0
					for i := 0; i <= len(line)-3; i++ {
						if line[i:i+3] == "'''" && !isInQuoteString(line, i) {
							count++
							if count%2 == 1 {
								tempInMultiLine = !tempInMultiLine
							}
							i += 2 // 跳过这个三引号
						}
					}
				}
			}
			
			// 如果这一行开始时在多行字符串中，整行都应该被保护
			// 如果这一行结束了多行字符串，需要处理字符串结束后的注释
			if !singleLineDocstring && inMultiLineString && !tempInMultiLine {
				// 多行字符串在这一行结束，需要找到结束位置并处理后面的注释
				var endPos int = -1
				if strings.Contains(line, `"""`) {
					endPos = strings.Index(line, `"""`) + 3
				} else if strings.Contains(line, "'''") {
					endPos = strings.Index(line, "'''") + 3
				}
				
				if endPos > 0 && endPos < len(line) {
					// 多行字符串结束后还有内容，需要处理注释
					beforeEnd := line[:endPos]
					afterEnd := line[endPos:]
					
					// 处理字符串结束后的部分
					processedAfter := afterEnd
					// 删除Python行注释
					if pos := strings.Index(processedAfter, "#"); pos != -1 {
						processedAfter = strings.TrimRight(processedAfter[:pos], " \t")
					}
					
					processedLine = beforeEnd + processedAfter
				}
			}
			
			inMultiLineString = tempInMultiLine
		}
		
		// 如果之前在多行字符串中，跳过注释处理
		if oldMultiLineState {
			result = append(result, processedLine)
			continue
		}
		
		// 如果之前在反引号字符串中但现在不在，说明模板字符串结束了，需要处理外部注释
		if oldBacktickState && !inBacktickString {
			// 这行包含了模板字符串的结束，检查外部注释
			if strings.Contains(line, "`") {
				lastBacktick := -1
				for i := len(line) - 1; i >= 0; i-- {
					if line[i] == '`' && !isEscaped(line, i) {
						lastBacktick = i
						break
					}
				}
				
				if lastBacktick != -1 && lastBacktick < len(line)-1 {
					afterBacktick := line[lastBacktick+1:]
					
					// 检查是否有注释符号
					for _, rule := range rules {
						if rule.IsLineComment {
							if pos := strings.Index(afterBacktick, rule.StartPattern); pos != -1 {
								// 找到外部注释，删除它
								beforeEnd := line[:lastBacktick+1]
								afterEnd := afterBacktick[:pos]
								processedLine = beforeEnd + strings.TrimRight(afterEnd, " \t")
								break
							}
						}
					}
				}
			}
		}
		
		// 如果当前在反引号字符串中，跳过注释处理
		if inBacktickString {
			result = append(result, originalLine)
			continue
		}
		
		
		// 如果在块注释中
		if inBlockComment {
			if pos := strings.Index(processedLine, blockEndPattern); pos != -1 {
				// 找到块注释结束，保留结束后的内容
				afterComment := processedLine[pos+len(blockEndPattern):]
				inBlockComment = false
				
				// 如果结束后还有内容，继续处理这部分内容
				if strings.TrimSpace(afterComment) != "" {
					// 递归处理剩余内容
					remaining := removeCommentsByRules(afterComment, fileType, rules)
					result = append(result, remaining)
				} else {
					// 块注释结束后没有内容，这一行变成空行
					result = append(result, "")
				}
			} else {
				// 整行都在块注释中，这一行变成空行
				result = append(result, "")
			}
			continue
		}
		
		// 处理行注释和块注释
		for _, rule := range rules {
			if rule.IsLineComment {
				// 处理行注释：需要找到第一个不在字符串内的注释符号
				pos := -1
				// YAML特殊处理：区分结构性注释和普通注释
				if fileType == "yaml" || fileType == "yml" {
					// 遍历所有可能的#位置
					for i := 0; i <= len(processedLine)-len(rule.StartPattern); i++ {
						if strings.HasPrefix(processedLine[i:], rule.StartPattern) {
							// 检查是否在字符串内
							if isInAnyString(processedLine, i) {
								continue
							}
							
							beforeComment := processedLine[:i]
							// 如果#前只有空白字符，这是行首注释，检查是否为结构性注释
							if strings.TrimSpace(beforeComment) == "" {
								// 行首注释，检查是否需要保护（只保护结构性注释）
								if shouldProtectInContext(originalLine, i, fileType, rule.StartPattern) {
									pos = -1 // 保护结构性注释，不删除
									break
								} else {
									pos = i // 删除普通注释
									break
								}
							} else {
								// 行尾注释，检查是否需要保护（Shell变量等）
								if !shouldProtectInContext(originalLine, i, fileType, rule.StartPattern) {
									pos = i
									break
								}
							}
						}
					}
				} else {
					// 其他语言的原有逻辑
					for i := 0; i <= len(processedLine)-len(rule.StartPattern); i++ {
						if strings.HasPrefix(processedLine[i:], rule.StartPattern) {
							// 检查是否在字符串内（包括原始字符串和正则表达式）
							if !isInAnyString(originalLine, i) && !isInRegex(originalLine, i) {
								// 检查是否需要保护
								protected := shouldProtectInContext(originalLine, i, fileType, rule.StartPattern)
								if !protected {
									pos = i
									break
								}
							}
						}
					}
				}
				// 如果找到了注释位置，处理注释删除
				if pos != -1 {
					beforeComment := processedLine[:pos]
					// 如果注释前只有空白字符，则整行都是注释
					if strings.TrimSpace(beforeComment) == "" {
						processedLine = "" // 整行注释，变成空行
					} else {
						// 删除注释但去除尾部空格
						processedLine = strings.TrimRight(beforeComment, " \t")
					}
					break
				}
			} else {
				// 处理块注释
				if pos := strings.Index(processedLine, rule.StartPattern); pos != -1 {
					if !shouldProtectInContext(originalLine, pos, fileType, rule.StartPattern) && 
					   !isInAnyString(originalLine, pos) && !isInBacktickString(originalLine, pos) && !isInRegex(originalLine, pos) {
						beforeComment := processedLine[:pos]
						
						// 检查同一行是否有结束标记
						if endPos := strings.Index(processedLine[pos:], rule.EndPattern); endPos != -1 {
							// 同一行内的块注释
							actualEndPos := pos + endPos + len(rule.EndPattern)
							afterComment := processedLine[actualEndPos:]
							
							// 合并注释前后的内容，保留原有空格
							if strings.TrimSpace(beforeComment) == "" && strings.TrimSpace(afterComment) == "" {
								// 整行都是注释，变成空行
								processedLine = ""
							} else {
								// 保留注释前后的内容和原有空格
								processedLine = beforeComment + afterComment
							}
						} else {
							// 块注释跨行
							inBlockComment = true
							blockEndPattern = rule.EndPattern
							if strings.TrimSpace(beforeComment) == "" {
								// 注释前只有空白，整行变成空行
								processedLine = ""
							} else {
								processedLine = beforeComment
							}
						}
						break
					}
				}
			}
		}
		
		result = append(result, processedLine)
	}
	
	// 清理结果：移除由注释产生的空行，但保留原有的空行
	var finalResult []string
	originalLines := strings.Split(content, "\n")
	
	for i, line := range result {
		// 如果是空行
		if strings.TrimSpace(line) == "" {
			// 检查原始行是否也是空行
			if i < len(originalLines) && strings.TrimSpace(originalLines[i]) == "" {
				// 原始行就是空行，保留
				finalResult = append(finalResult, line)
			}
		} else {
			// 非空行，直接保留
			finalResult = append(finalResult, line)
		}
	}
	
	return strings.Join(finalResult, "\n")
}

// 保留原有函数名作为兼容性包装
func isInBacktickString(line string, pos int) bool {
	return isInStringWithType(line, pos, StringTypeBacktick)
}

// removeCommentsByFileType 根据文件类型删除注释的统一函数
func removeCommentsByFileType(content, fileType string) string {
	rules := getCommentRulesForLanguage(fileType)
	return removeCommentsByRules(content, fileType, rules)
}

// removeComments 移除指定文件类型的注释
func removeComments(content, fileType string) string {
	// 统一使用规则处理所有文件类型
	return removeCommentsByFileType(content, fileType)
}

// StringType 字符串类型枚举
type StringType int

const (
	StringTypeAll StringType = iota // 所有类型字符串
	StringTypeQuote                 // 仅单双引号字符串
	StringTypeBacktick              // 仅反引号字符串
)

// isInStringWithType 统一的字符串检测函数
func isInStringWithType(line string, pos int, stringType StringType) bool {
	if pos >= len(line) {
		return false
	}
	
	var inSingleQuote, inDoubleQuote, inBacktick bool
	
	// 检查到pos位置之前的所有字符（不包括pos位置本身）
	for i := 0; i < pos; i++ {
		char := line[i]
		
		switch char {
		case '\'':
			if !inDoubleQuote && !inBacktick && (stringType == StringTypeAll || stringType == StringTypeQuote) {
				if !isEscaped(line, i) {
					inSingleQuote = !inSingleQuote
				}
			}
		case '"':
			if !inSingleQuote && !inBacktick && (stringType == StringTypeAll || stringType == StringTypeQuote) {
				if !isEscaped(line, i) {
					inDoubleQuote = !inDoubleQuote
				}
			}
		case '`':
			if !inSingleQuote && !inDoubleQuote && (stringType == StringTypeAll || stringType == StringTypeBacktick) {
				inBacktick = !inBacktick
			}
		}
	}
	
	switch stringType {
	case StringTypeQuote:
		return inSingleQuote || inDoubleQuote
	case StringTypeBacktick:
		return inBacktick
	default: // StringTypeAll
		return inSingleQuote || inDoubleQuote || inBacktick
	}
}

// isEscaped 检查字符是否被转义
func isEscaped(line string, pos int) bool {
	if pos == 0 {
		return false
	}
	
	backslashCount := 0
	for i := pos - 1; i >= 0 && line[i] == '\\'; i-- {
		backslashCount++
	}
	// 奇数个反斜杠表示当前字符被转义
	return backslashCount%2 == 1
}

// isStructuralComment 检查是否为结构性注释（通用模式）
func isStructuralComment(comment string) bool {
	// 去掉注释符号，获取纯内容
	content := strings.TrimSpace(strings.TrimPrefix(comment, "#"))
	
	// 空注释或只有符号的注释不是结构性的
	if len(content) == 0 {
		return false
	}
	
	// 排除明显的普通注释模式
	commonPhrases := []string{"这是", "这个", "用于", "表示", "注释", "说明"}
	for _, phrase := range commonPhrases {
		if strings.Contains(content, phrase) {
			return false
		}
	}
	
	// 1. 包含emoji的注释通常是结构性的
	if containsEmoji(content) {
		return true
	}
	
	// 2. 包含分隔符的注释通常是结构性的
	separators := []string{"===", "---", "***", "###", "+++", "~~~"}
	for _, sep := range separators {
		if strings.Contains(content, sep) {
			return true
		}
	}
	
	// 3. 以数字开头的注释通常是步骤或列表项
	if len(content) > 0 && (content[0] >= '0' && content[0] <= '9') {
		return true
	}
	
	// 4. 短且包含特殊字符的通常是结构性的
	if len(content) <= 15 {
		specialChars := []string{"→", "•", "★", "▶", "◆", "■", "▲", "►"}
		for _, char := range specialChars {
			if strings.Contains(content, char) {
				return true
			}
		}
	}
	
	// 5. 全大写且较短的注释通常是标题
	if strings.ToUpper(content) == content && len(content) > 2 && len(content) <= 20 {
		// 排除常见的普通注释词汇
		commonWords := []string{"TODO", "FIXME", "HACK", "NOTE", "WARNING"}
		for _, word := range commonWords {
			if strings.Contains(content, word) {
				return false
			}
		}
		return true
	}
	
	return false
}

// containsEmoji 检查字符串是否包含emoji
func containsEmoji(s string) bool {
	for _, r := range s {
		// 检查常见的emoji范围
		if (r >= 0x1F600 && r <= 0x1F64F) || // 表情符号
		   (r >= 0x1F300 && r <= 0x1F5FF) || // 杂项符号
		   (r >= 0x1F680 && r <= 0x1F6FF) || // 交通和地图符号
		   (r >= 0x2600 && r <= 0x26FF) ||   // 杂项符号
		   (r >= 0x2700 && r <= 0x27BF) ||   // 装饰符号
		   (r >= 0x1F900 && r <= 0x1F9FF) {  // 补充符号
			return true
		}
	}
	return false
}

// 保留原有函数名作为兼容性包装
func isInQuoteString(line string, pos int) bool {
	return isInStringWithType(line, pos, StringTypeQuote)
}

func isInAnyString(line string, pos int) bool {
	return isInStringWithType(line, pos, StringTypeAll)
}

func isInString(line string, pos int) bool {
	return isInStringWithType(line, pos, StringTypeAll)
}

// isInRegex 检查指定位置是否在正则表达式内
func isInRegex(line string, pos int) bool {
	if pos >= len(line) {
		return false
	}
	
	var inSingleQuote, inDoubleQuote, inBacktick bool
	var inRegex bool
	
	for i := 0; i < pos; i++ {
		char := line[i]
		
		switch char {
		case '\'':
			if !inDoubleQuote && !inBacktick && !inRegex {
				if !isEscaped(line, i) {
					inSingleQuote = !inSingleQuote
				}
			}
		case '"':
			if !inSingleQuote && !inBacktick && !inRegex {
				if !isEscaped(line, i) {
					inDoubleQuote = !inDoubleQuote
				}
			}
		case '`':
			if !inSingleQuote && !inDoubleQuote && !inRegex {
				inBacktick = !inBacktick
			}
		case '/':
			if !inSingleQuote && !inDoubleQuote && !inBacktick {
				if inRegex {
					if !isEscaped(line, i) {
						inRegex = false
					}
				} else {
					// 检查是否是正则表达式开始
					if i > 0 {
						j := i - 1
						for j >= 0 && (line[j] == ' ' || line[j] == '\t') {
							j--
						}
						if j >= 0 {
							prevChar := line[j]
							if prevChar == '=' || prevChar == '(' || prevChar == ',' || prevChar == ':' || 
							   prevChar == '[' || prevChar == '{' || prevChar == ';' {
								inRegex = true
							}
						}
					} else {
						inRegex = true
					}
				}
			}
		}
	}
	
	return inRegex
}

// processFile 处理单个文件，删除其中的注释
func processFile(filePath, workingDir string) error {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}
	
	// 安全检查
	if err := isFileSafe(filePath, content, forceMode); err != nil {
		printWarning("跳过 %s (二进制文件)", filePath)
		skippedFiles = append(skippedFiles, filePath)
		return nil // 跳过
	}
	
	// 检测文件类型
	fileType := detectFileType(filePath)
	
	// 删除注释
	newContent := removeComments(string(content), fileType)
	
	// 检查是否有变化
	if newContent == string(content) {
		fmt.Printf(ColorBlue+"%-40s"+ColorReset+" |%s| "+ColorYellow+"无变化\n"+ColorReset, filePath, strings.ToUpper(fileType))
		return nil
	}
	
	// 只有在有变化时才创建备份
	if err := createBackup(filePath, workingDir); err != nil {
		return fmt.Errorf("创建备份失败: %v", err)
	}
	
	// 写回文件
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}
	
	fmt.Printf(ColorGreen+"%-40s"+ColorReset+" |%s| "+ColorGreen+"✓\n"+ColorReset, filePath, strings.ToUpper(fileType))
	processedFiles = append(processedFiles, filePath)
	
	return nil
}

// isSupportedFile 检查文件是否为支持的类型
// 如果 force 为 true，则支持所有文件类型
func isSupportedFile(filePath string, force bool) bool {
	if force {
		return true
	}
	
	ext := strings.ToLower(filepath.Ext(filePath))
	return supportedExtensions[ext]
}

// processDirectory 递归处理目录中的所有支持文件
func processDirectory(dirPath string) error {
	var processedCount int
	
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// 跳过备份目录
		if d.IsDir() && d.Name() == "bak" {
			return filepath.SkipDir
		}
		
		// 跳过目录和隐藏文件
		if d.IsDir() || strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		
		// 检查是否为支持的文件类型
		if !isSupportedFile(path, forceMode) {
			return nil
		}
		
		// 处理文件
		if err := processFile(path, dirPath); err != nil {
			printError("处理文件失败 %s: %v", path, err)
			return nil // 继续处理其他文件
		}
		
		processedCount++
		return nil
	})
	
	if err != nil {
		return err
	}
	
	// 显示处理结果摘要
	printSummary()
	return nil
}

var rootCmd = &cobra.Command{
	Use:   "fuck-comment [directory]",
	Short: "删除代码注释的命令行工具",
	Long: `删除代码文件中的注释，支持137种文件扩展名。

支持的注释格式：
  //           行注释 (C/C++, Go, Java, JavaScript等)
  /* */        块注释 (C/C++, Go, Java, JavaScript等)
  #            井号注释 (Python, Shell, YAML等)
  --           双破折号注释 (SQL, Haskell等)
  ;            分号注释 (Assembly, Lisp等)
  %            百分号注释 (LaTeX, MATLAB等)
  !            感叹号注释 (Fortran等)
  <!-- -->     HTML注释 (HTML, XML等)

安全特性：
  • 自动备份到 bak/ 目录
  • 跳过二进制文件
  • 保护字符串中的注释符号
  • 保护URL锚点和Shell变量

参数说明：
  -f, --file string    指定要处理的单个文件
      --force          强制处理所有文件类型（包括二进制文件）
      --version        显示版本信息

使用示例:
  fuck-comment              删除当前目录所有支持文件的注释
  fuck-comment /path/to/dir 删除指定目录及其子目录的注释
  fuck-comment -f main.go   删除指定文件的注释
  fuck-comment --force      强制处理所有文件类型

注意事项：
  • 处理前会自动创建备份，备份文件保存在 bak/ 目录
  • 默认跳过二进制文件和隐藏文件
  • 使用 --force 参数可强制处理所有文件类型`,
	Run: func(cmd *cobra.Command, args []string) {
		// 显示版本信息
		if showVersion {
			fmt.Printf(ColorBold+ColorCyan+"fuck-comment %s\n"+ColorReset, Version)
			fmt.Printf("构建时间: %s\n", BuildTime)
			fmt.Printf("Git提交: %s\n", GitCommit)
			return
		}
		if targetFile != "" {
			// 处理单个文件
			if !isSupportedFile(targetFile, forceMode) && !forceMode {
				printError("不支持的文件类型: %s", targetFile)
				fmt.Println("使用 --force 参数可强制处理所有文件类型")
				os.Exit(1)
			}
			
			// 获取文件所在目录作为工作目录
			fileDir := filepath.Dir(targetFile)
			if err := processFile(targetFile, fileDir); err != nil {
				printError("处理文件失败: %v", err)
				os.Exit(1)
			}
			
			printSummary()
		} else {
			// 处理目录
			var targetDir string
			if len(args) > 0 {
				// 使用命令行参数指定的目录
				targetDir = args[0]
				// 检查目录是否存在
				if _, err := os.Stat(targetDir); os.IsNotExist(err) {
					printError("目录不存在: %s", targetDir)
					os.Exit(1)
				}
			} else {
				// 使用当前目录
				var err error
				targetDir, err = os.Getwd()
				if err != nil {
					printError("获取当前目录失败: %v", err)
					os.Exit(1)
				}
			}
			
			fmt.Printf(ColorPurple+"扫描目录: %s\n"+ColorReset, targetDir)
			if err := processDirectory(targetDir); err != nil {
				printError("处理目录失败: %v", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.Flags().StringVarP(&targetFile, "file", "f", "", "指定要处理的单个文件")
	rootCmd.Flags().BoolVar(&forceMode, "force", false, "强制处理所有文件类型（包括二进制文件）")
	rootCmd.Flags().BoolVar(&showVersion, "version", false, "显示版本信息")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		printError("执行失败: %v", err)
		os.Exit(1)
	}
}
