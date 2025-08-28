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
		backupRootDir = filepath.Join("bak", dirName+"_"+backupTimestamp)
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

// shouldProtectInContext 检查是否应该在特定上下文中保护注释
func shouldProtectInContext(line string, pos int, fileType string, commentStart string) bool {
	// 通用保护：字符串内的注释符号
	if isInString(line, pos) {
		return true
	}
	
	switch fileType {
	case "markdown":
		// 保护表格中的HTML注释示例
		if commentStart == "<!--" && strings.Contains(line, "|") {
			return true
		}
	case "yaml", "yml":
		// 保护URL中的锚点和Shell变量
		if commentStart == "#" {
			beforeComment := line[:pos]  // 不要trim，保持原始格式
			// 保护URL锚点
			if strings.Contains(beforeComment, "http") {
				return true
			}
			// 保护Shell变量如 ${GITHUB_REF#refs/tags/}
			if strings.Contains(beforeComment, "${") {
				return true
			}
			// 保护任何包含$的行中的#
			if strings.Contains(beforeComment, "$") {
				return true
			}
		}
	case "css", "scss", "less", "stylus":
		// CSS中保护URL和content属性中的注释符号
		if commentStart == "/*" || commentStart == "//" {
			// 检查是否在url()函数中
			if strings.Contains(line[:pos], "url(") && !strings.Contains(line[:pos], ")") {
				return true
			}
			// 检查是否在content属性中
			if strings.Contains(line[:pos], "content:") {
				return true
			}
		}
	case "html", "xml", "svg":
		// HTML/XML中保护属性值和CDATA中的注释符号
		if commentStart == "<!--" {
			// 检查是否在CDATA中
			if strings.Contains(line[:pos], "<![CDATA[") && !strings.Contains(line[:pos], "]]>") {
				return true
			}
		}
	case "javascript", "typescript", "jsx", "tsx", "js", "ts":
		// JavaScript中保护正则表达式和模板字符串
		if commentStart == "//" || commentStart == "/*" {
			beforeComment := line[:pos]
			// 保护正则表达式 /pattern/
			if strings.Contains(beforeComment, "= /") || strings.Contains(beforeComment, "(/") {
				return true
			}
			// 保护模板字符串 `template`
			backtickCount := strings.Count(beforeComment, "`")
			if backtickCount%2 == 1 {
				return true
			}
		}
	case "go", "c", "cpp", "java", "c#", "cs":
		// 保护条件语句和不完整的语句
		if commentStart == "//" || commentStart == "/*" {
			beforeComment := strings.TrimSpace(line[:pos])
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
			// 保护包含 != 的语句
			if strings.Contains(beforeComment, "!=") && !strings.Contains(beforeComment, "{") {
				return true
			}
		}
	case "python":
		// Python中保护docstring和f-string
		if commentStart == "#" {
			// 检查是否在f-string中
			if strings.Contains(line[:pos], "f\"") || strings.Contains(line[:pos], "f'") {
				return true
			}
		}
	case "shell", "bash", "zsh":
		// Shell脚本中保护shebang和特殊变量
		if commentStart == "#" {
			// 保护shebang
			if pos == 0 && strings.HasPrefix(line, "#!") {
				return true
			}
			// 保护变量替换中的#
			beforeComment := line[:pos]
			if strings.Contains(beforeComment, "${") && !strings.Contains(beforeComment, "}") {
				return true
			}
		}
	case "sql":
		// SQL中保护字符串和标识符
		if commentStart == "--" || commentStart == "/*" {
			// 已经通过通用字符串保护处理
		}
	}
	return false
}

// removeCommentsByRules 通用注释删除函数
func removeCommentsByRules(content string, fileType string, rules []CommentRule) string {
	lines := strings.Split(content, "\n")
	var result []string
	inBlockComment := false
	inCodeBlock := false
	inTemplateString := false
	currentBlockEnd := ""
	
	for _, line := range lines {
		originalLine := line
		processedLine := line
		
		// Markdown特殊处理：检测代码块
		if fileType == "markdown" {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "```") {
				inCodeBlock = !inCodeBlock
				result = append(result, line)
				continue
			}
			if inCodeBlock {
				result = append(result, line)
				continue
			}
		}
		
		// JavaScript特殊处理：检测模板字符串
		if fileType == "javascript" || fileType == "typescript" || fileType == "js" || fileType == "ts" {
			backtickCount := strings.Count(line, "`")
			if backtickCount > 0 {
				for i := 0; i < backtickCount; i++ {
					inTemplateString = !inTemplateString
				}
			}
			if inTemplateString {
				result = append(result, line)
				continue
			}
		}
		
		// 处理跨行块注释
		if inBlockComment {
			if endIndex := strings.Index(line, currentBlockEnd); endIndex != -1 {
				processedLine = line[endIndex+len(currentBlockEnd):]
				inBlockComment = false
				currentBlockEnd = ""
			} else {
				continue
			}
		}
		
		// 应用注释规则
		for _, rule := range rules {
			if rule.IsLineComment {
				// 处理行注释
				for i := 0; i <= len(processedLine)-len(rule.StartPattern); i++ {
					if strings.HasPrefix(processedLine[i:], rule.StartPattern) {
						// 检查是否在字符串中
						if isInString(processedLine, i) {
							continue
						}
						// 检查是否需要保护
						if shouldProtectInContext(processedLine, i, fileType, rule.StartPattern) {
							continue
						}
						// 删除行注释
						processedLine = strings.TrimRight(processedLine[:i], " \t")
						break
					}
				}
			} else {
				// 处理块注释
				for {
					startIdx := strings.Index(processedLine, rule.StartPattern)
					if startIdx == -1 {
						break
					}
					// 检查是否在字符串中
					if isInString(processedLine, startIdx) {
						break
					}
					// 检查是否需要保护
					if shouldProtectInContext(processedLine, startIdx, fileType, rule.StartPattern) {
						break
					}
					
					endIdx := strings.Index(processedLine[startIdx:], rule.EndPattern)
					if endIdx != -1 {
						// 同一行内的块注释
						endIdx += startIdx + len(rule.EndPattern)
						processedLine = processedLine[:startIdx] + processedLine[endIdx:]
					} else {
						// 跨行块注释开始
						processedLine = processedLine[:startIdx]
						inBlockComment = true
						currentBlockEnd = rule.EndPattern
						break
					}
				}
			}
		}
		
		// 保留非空行或原本就是空行的行
		if strings.TrimSpace(processedLine) != "" || strings.TrimSpace(originalLine) == "" {
			result = append(result, processedLine)
		}
	}
	
	return strings.Join(result, "\n")
}

// removeMarkdownComments 处理 Markdown 文件
func removeMarkdownComments(content string) string {
	rules := []CommentRule{
		{StartPattern: "<!--", EndPattern: "-->", IsLineComment: false},
	}
	return removeCommentsByRules(content, "markdown", rules)
}

// removeYamlComments 处理 YAML 文件
func removeYamlComments(content string) string {
	rules := []CommentRule{
		{StartPattern: "#", EndPattern: "", IsLineComment: true},
	}
	return removeCommentsByRules(content, "yaml", rules)
}

// removeJsonComments 处理 JSON 文件
func removeJsonComments(content string) string {
	rules := []CommentRule{
		{StartPattern: "//", EndPattern: "", IsLineComment: true},
		{StartPattern: "/*", EndPattern: "*/", IsLineComment: false},
	}
	return removeCommentsByRules(content, "json", rules)
}

// removeXmlComments 处理 XML/HTML 文件
func removeXmlComments(content string) string {
	rules := []CommentRule{
		{StartPattern: "<!--", EndPattern: "-->", IsLineComment: false},
	}
	return removeCommentsByRules(content, "xml", rules)
}

// removeCssComments 处理 CSS 文件
func removeCssComments(content string) string {
	rules := []CommentRule{
		{StartPattern: "/*", EndPattern: "*/", IsLineComment: false},
	}
	return removeCommentsByRules(content, "css", rules)
}

// removeGoComments 处理 Go 文件
func removeGoComments(content string) string {
	rules := []CommentRule{
		{StartPattern: "//", EndPattern: "", IsLineComment: true},
		{StartPattern: "/*", EndPattern: "*/", IsLineComment: false},
	}
	return removeCommentsByRules(content, "go", rules)
}

// removeComments 根据文件类型智能删除注释
func removeComments(content string, fileType string) string {
	// 对于特殊文件类型，不处理或特殊处理
	switch fileType {
	case "markdown":
		return removeMarkdownComments(content)
	case "yaml", "yml":
		return removeYamlComments(content)
	case "json", "jsonc", "json5":
		return removeJsonComments(content)
	case "xml", "html", "htm", "svg":
		return removeXmlComments(content)
	case "css", "scss", "sass", "less", "styl":
		return removeCssComments(content)
	case "go":
		return removeGoComments(content)
	case "javascript", "typescript", "js", "ts", "jsx", "tsx":
		rules := []CommentRule{
			{StartPattern: "//", EndPattern: "", IsLineComment: true},
			{StartPattern: "/*", EndPattern: "*/", IsLineComment: false},
		}
		return removeCommentsByRules(content, fileType, rules)
	case "c", "cpp", "cc", "cxx", "h", "hpp", "cs", "java", "scala", "kt", "groovy":
		rules := []CommentRule{
			{StartPattern: "//", EndPattern: "", IsLineComment: true},
			{StartPattern: "/*", EndPattern: "*/", IsLineComment: false},
		}
		return removeCommentsByRules(content, fileType, rules)
	}
	
	lines := strings.Split(content, "\n")
	var result []string
	inBlockComment := false
	inHTMLComment := false
	inBacktickString := false // 跟踪反引号字符串状态

	for _, line := range lines {
		originalLine := line
		processedLine := line
		
		// 处理HTML注释块
		if inHTMLComment {
			if endIndex := strings.Index(line, "-->"); endIndex != -1 {
				processedLine = line[endIndex+3:]
				inHTMLComment = false
			} else {
				// 整行都在HTML注释中，跳过这一行
				continue
			}
		}
		
		// 处理C风格块注释
		if inBlockComment {
			if endIndex := strings.Index(line, "*/"); endIndex != -1 {
				processedLine = line[endIndex+2:]
				inBlockComment = false
			} else {
				// 整行都在块注释中，跳过这一行
				continue
			}
		}
		
		// 更新反引号字符串状态
		for i, char := range processedLine {
			if char == '`' && !inBlockComment && !inHTMLComment {
				// 检查是否在其他类型的字符串中
				if !isInQuoteString(processedLine, i) {
					inBacktickString = !inBacktickString
				}
			}
		}
		
		if !inBlockComment && !inHTMLComment && !inBacktickString {
			// 找到最早的注释位置，避免冲突
			earliestCommentPos := len(processedLine)
			
			// 检查C风格行注释 //
			for i := 0; i < len(processedLine)-1; i++ {
				if processedLine[i] == '/' && processedLine[i+1] == '/' && !isInString(processedLine, i) {
					if i < earliestCommentPos {
						earliestCommentPos = i
					}
					break
				}
			}
			
			// 检查双破折号注释 -- (Haskell, Ada, SQL等)
			for i := 0; i < len(processedLine)-1; i++ {
				if processedLine[i] == '-' && processedLine[i+1] == '-' && !isInString(processedLine, i) {
					if i < earliestCommentPos {
						earliestCommentPos = i
					}
					break
				}
			}
			
			// 检查Python/Shell风格行注释 # (只有在非字符串且有实际内容时才处理)
			for i := 0; i < len(processedLine); i++ {
				if processedLine[i] == '#' && !isInString(processedLine, i) {
					// 确保不是单独的字符
					if len(strings.TrimSpace(processedLine)) > 1 || i > 0 {
						if i < earliestCommentPos {
							earliestCommentPos = i
						}
						break
					}
				}
			}
			
			// 检查分号注释 ; (Assembly, Lisp等)
			for i := 0; i < len(processedLine); i++ {
				if processedLine[i] == ';' && !isInString(processedLine, i) {
					// 确保不是单独的字符
					if len(strings.TrimSpace(processedLine)) > 1 || i > 0 {
						if i < earliestCommentPos {
							earliestCommentPos = i
						}
						break
					}
				}
			}
			
			// 检查百分号注释 % (LaTeX, MATLAB等)
			for i := 0; i < len(processedLine); i++ {
				if processedLine[i] == '%' && !isInString(processedLine, i) {
					// 确保不是单独的字符
					if len(strings.TrimSpace(processedLine)) > 1 || i > 0 {
						if i < earliestCommentPos {
							earliestCommentPos = i
						}
						break
					}
				}
			}
			
			// 检查感叹号注释 ! (Fortran等)
			for i := 0; i < len(processedLine); i++ {
				if processedLine[i] == '!' && !isInString(processedLine, i) {
					// 确保不是单独的字符
					if len(strings.TrimSpace(processedLine)) > 1 || i > 0 {
						if i < earliestCommentPos {
							earliestCommentPos = i
						}
						break
					}
				}
			}
			
			// 如果找到了注释，截断到该位置
			if earliestCommentPos < len(processedLine) {
				processedLine = processedLine[:earliestCommentPos]
			}
			
			// 处理HTML注释 <!-- -->
			for {
				startIdx := strings.Index(processedLine, "<!--")
				if startIdx == -1 || isInString(processedLine, startIdx) {
					break
				}
				
				endIdx := strings.Index(processedLine[startIdx:], "-->")
				if endIdx != -1 {
					// 同一行内的HTML注释
					endIdx += startIdx + 3
					processedLine = processedLine[:startIdx] + processedLine[endIdx:]
				} else {
					// 跨行HTML注释开始
					processedLine = processedLine[:startIdx]
					inHTMLComment = true
					break
				}
			}
			
			// 处理C风格块注释 /* */
			for {
				startIdx := strings.Index(processedLine, "/*")
				if startIdx == -1 || isInString(processedLine, startIdx) {
					break
				}
				
				endIdx := strings.Index(processedLine[startIdx:], "*/")
				if endIdx != -1 {
					// 同一行内的块注释
					endIdx += startIdx + 2
					processedLine = processedLine[:startIdx] + processedLine[endIdx:]
				} else {
					// 跨行块注释开始
					processedLine = processedLine[:startIdx]
					inBlockComment = true
					break
				}
			}
		}
		
		// 移除行尾空白
		processedLine = strings.TrimRight(processedLine, " \t")
		
		// 如果处理后的行只包含空白字符，且原行包含注释，则跳过该行
		if strings.TrimSpace(processedLine) == "" && strings.TrimSpace(originalLine) != "" {
			trimmedOriginal := strings.TrimSpace(originalLine)
			if strings.HasPrefix(trimmedOriginal, "//") || 
			   strings.HasPrefix(trimmedOriginal, "/*") || 
			   strings.HasPrefix(trimmedOriginal, "#") ||
			   strings.HasPrefix(trimmedOriginal, "<!--") ||
			   strings.HasPrefix(trimmedOriginal, "--") ||
			   strings.HasPrefix(trimmedOriginal, ";") ||
			   strings.HasPrefix(trimmedOriginal, "%") ||
			   strings.HasPrefix(trimmedOriginal, "!") ||
			   strings.Contains(trimmedOriginal, "*/") ||
			   strings.Contains(trimmedOriginal, "-->") {
				continue
			}
		}
		
		result = append(result, processedLine)
	}
	
	return strings.Join(result, "\n")
}

// isInQuoteString 检查指定位置是否在单引号或双引号字符串内（不包括反引号）
func isInQuoteString(line string, pos int) bool {
	if pos >= len(line) {
		return false
	}
	
	var inSingleQuote, inDoubleQuote bool
	lineBytes := []byte(line)
	
	for i := 0; i <= pos && i < len(lineBytes); i++ {
		char := lineBytes[i]
		
		switch char {
		case '\'':
			if !inDoubleQuote {
				backslashCount := 0
				for j := i - 1; j >= 0 && lineBytes[j] == '\\'; j-- {
					backslashCount++
				}
				if backslashCount%2 == 0 {
					inSingleQuote = !inSingleQuote
				}
			}
		case '"':
			if !inSingleQuote {
				backslashCount := 0
				for j := i - 1; j >= 0 && lineBytes[j] == '\\'; j-- {
					backslashCount++
				}
				if backslashCount%2 == 0 {
					inDoubleQuote = !inDoubleQuote
				}
			}
		}
		
		if i == pos {
			return inSingleQuote || inDoubleQuote
		}
	}
	
	return inSingleQuote || inDoubleQuote
}

// isInString 检查指定位置是否在字符串字面量内（优化版本）
func isInString(line string, pos int) bool {
	if pos >= len(line) {
		return false
	}
	
	var inSingleQuote, inDoubleQuote, inBacktick bool
	
	// 优化：使用字节切片避免重复的字符串索引
	lineBytes := []byte(line)
	
	for i := 0; i <= pos && i < len(lineBytes); i++ {
		char := lineBytes[i]
		
		switch char {
		case '\'':
			if !inDoubleQuote && !inBacktick {
				// 优化：直接计算反斜杠数量，避免重复循环
				backslashCount := 0
				for j := i - 1; j >= 0 && lineBytes[j] == '\\'; j-- {
					backslashCount++
				}
				// 如果反斜杠数量为偶数，引号未被转义
				if backslashCount%2 == 0 {
					inSingleQuote = !inSingleQuote
				}
			}
		case '"':
			if !inSingleQuote && !inBacktick {
				// 优化：直接计算反斜杠数量，避免重复循环
				backslashCount := 0
				for j := i - 1; j >= 0 && lineBytes[j] == '\\'; j-- {
					backslashCount++
				}
				// 如果反斜杠数量为偶数，引号未被转义
				if backslashCount%2 == 0 {
					inDoubleQuote = !inDoubleQuote
				}
			}
		case '`':
			if !inSingleQuote && !inDoubleQuote {
				inBacktick = !inBacktick
			}
		}
		
		// 如果我们已经到达目标位置，返回当前状态
		if i == pos {
			return inSingleQuote || inDoubleQuote || inBacktick
		}
	}
	
	return inSingleQuote || inDoubleQuote || inBacktick
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
	
	// 创建备份
	if err := createBackup(filePath, workingDir); err != nil {
		return fmt.Errorf("创建备份失败: %v", err)
	}
	
	// 删除注释
	newContent := removeComments(string(content), fileType)
	
	// 检查是否有变化
	if newContent == string(content) {
		// 删除不必要的备份
		os.Remove(filePath + ".backup")
		fmt.Printf(ColorBlue+"%-40s"+ColorReset+" |%s| "+ColorYellow+"无变化\n"+ColorReset, filePath, strings.ToUpper(fileType))
		return nil
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
