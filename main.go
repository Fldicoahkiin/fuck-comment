package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// 版本信息，在构建时通过 ldflags 注入
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"

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

	// CLI 参数
	targetFile string
	forceMode  bool
	verbose    bool
	showVersion bool
)

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

// removeMarkdownComments 处理 Markdown 文件 - 不删除 # 标题
func removeMarkdownComments(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inCodeBlock := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// 检测代码块
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			result = append(result, line)
			continue
		}
		
		// 在代码块内，不处理注释
		if inCodeBlock {
			result = append(result, line)
			continue
		}
		
		// 只删除 HTML 注释，保留 # 标题
		processedLine := line
		for {
			startIdx := strings.Index(processedLine, "<!--")
			if startIdx == -1 {
				break
			}
			endIdx := strings.Index(processedLine[startIdx:], "-->")
			if endIdx != -1 {
				endIdx += startIdx + 3
				processedLine = processedLine[:startIdx] + processedLine[endIdx:]
			} else {
				processedLine = processedLine[:startIdx]
				break
			}
		}
		
		result = append(result, processedLine)
	}
	
	return strings.Join(result, "\n")
}

// removeYamlComments 处理 YAML 文件 - 智能删除注释，保护YAML结构
func removeYamlComments(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	
	for _, line := range lines {
		processedLine := line
		inDoubleQuote := false
		inSingleQuote := false
		escaped := false
		bracketDepth := 0
		
		for i := 0; i < len(line); i++ {
			char := line[i]
			
			if escaped {
				escaped = false
				continue
			}
			
			if char == '\\' && (inDoubleQuote || inSingleQuote) {
				escaped = true
				continue
			}
			
			// 跟踪引号状态
			if char == '"' && !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
				continue
			}
			if char == '\'' && !inDoubleQuote {
				inSingleQuote = !inSingleQuote
				continue
			}
			
			// 跟踪数组/对象括号
			if !inDoubleQuote && !inSingleQuote {
				if char == '[' || char == '{' {
					bracketDepth++
				} else if char == ']' || char == '}' {
					bracketDepth--
				}
			}
			
			// 只在字符串外且不在数组/对象内删除 # 注释
			if char == '#' && !inDoubleQuote && !inSingleQuote && bracketDepth == 0 {
				beforeHash := strings.TrimSpace(line[:i])
				
				// 检查是否是YAML键值对的一部分
				if beforeHash == "" {
					// 整行都是注释
					processedLine = ""
					break
				} else if strings.Contains(beforeHash, ":") {
					// 包含冒号，可能是键值对后的注释
					processedLine = strings.TrimRight(line[:i], " \t")
					break
				} else {
					// 可能是值的一部分，保留
					continue
				}
			}
		}
		
		result = append(result, processedLine)
	}
	
	return strings.Join(result, "\n")
}

// removeJsonComments 处理 JSON 文件 - 删除 // 和 /* */ 注释
func removeJsonComments(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inBlockComment := false
	
	for _, line := range lines {
		processedLine := line
		
		if inBlockComment {
			if endIndex := strings.Index(line, "*/"); endIndex != -1 {
				processedLine = line[endIndex+2:]
				inBlockComment = false
			} else {
				continue
			}
		}
		
		// 处理行注释 //
		if idx := strings.Index(processedLine, "//"); idx != -1 && !isInString(processedLine, idx) {
			processedLine = strings.TrimRight(processedLine[:idx], " \t")
		}
		
		// 处理块注释 /* */
		for {
			startIdx := strings.Index(processedLine, "/*")
			if startIdx == -1 || isInString(processedLine, startIdx) {
				break
			}
			
			endIdx := strings.Index(processedLine[startIdx:], "*/")
			if endIdx != -1 {
				endIdx += startIdx + 2
				processedLine = processedLine[:startIdx] + processedLine[endIdx:]
			} else {
				processedLine = processedLine[:startIdx]
				inBlockComment = true
				break
			}
		}
		
		result = append(result, processedLine)
	}
	
	return strings.Join(result, "\n")
}

// removeXmlComments 处理 XML/HTML 文件 - 只删除 <!-- --> 注释
func removeXmlComments(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inComment := false
	
	for _, line := range lines {
		processedLine := line
		
		if inComment {
			if endIndex := strings.Index(line, "-->"); endIndex != -1 {
				processedLine = line[endIndex+3:]
				inComment = false
			} else {
				continue
			}
		}
		
		// 处理 HTML 注释
		for {
			startIdx := strings.Index(processedLine, "<!--")
			if startIdx == -1 {
				break
			}
			
			endIdx := strings.Index(processedLine[startIdx:], "-->")
			if endIdx != -1 {
				endIdx += startIdx + 3
				processedLine = processedLine[:startIdx] + processedLine[endIdx:]
			} else {
				processedLine = processedLine[:startIdx]
				inComment = true
				break
			}
		}
		
		result = append(result, processedLine)
	}
	
	return strings.Join(result, "\n")
}

// removeCssComments 处理 CSS 文件 - 只删除 /* */ 注释
func removeCssComments(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inComment := false
	
	for _, line := range lines {
		processedLine := line
		
		if inComment {
			if endIndex := strings.Index(line, "*/"); endIndex != -1 {
				processedLine = line[endIndex+2:]
				inComment = false
			} else {
				continue
			}
		}
		
		// 处理块注释 /* */
		for {
			startIdx := strings.Index(processedLine, "/*")
			if startIdx == -1 {
				break
			}
			
			endIdx := strings.Index(processedLine[startIdx:], "*/")
			if endIdx != -1 {
				endIdx += startIdx + 2
				processedLine = processedLine[:startIdx] + processedLine[endIdx:]
			} else {
				processedLine = processedLine[:startIdx]
				inComment = true
				break
			}
		}
		
		result = append(result, processedLine)
	}
	
	return strings.Join(result, "\n")
}

// removeComments 根据文件类型智能删除注释
func removeComments(content string, fileType string) string {
	// 对于特殊文件类型，不处理或特殊处理
	switch fileType {
	case "markdown":
		return removeMarkdownComments(content)
	case "yaml":
		return removeYamlComments(content)
	case "json":
		return removeJsonComments(content)
	case "xml", "html":
		return removeXmlComments(content)
	case "css":
		return removeCssComments(content)
	}
	
	lines := strings.Split(content, "\n")
	var result []string
	inBlockComment := false
	inHTMLComment := false

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
		
		if !inBlockComment && !inHTMLComment {
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
			
			// 检查Python/Shell风格行注释 #
			for i := 0; i < len(processedLine); i++ {
				if processedLine[i] == '#' && !isInString(processedLine, i) {
					if i < earliestCommentPos {
						earliestCommentPos = i
					}
					break
				}
			}
			
			// 检查分号注释 ; (Assembly, Lisp等)
			for i := 0; i < len(processedLine); i++ {
				if processedLine[i] == ';' && !isInString(processedLine, i) {
					if i < earliestCommentPos {
						earliestCommentPos = i
					}
					break
				}
			}
			
			// 检查百分号注释 % (LaTeX, MATLAB等)
			for i := 0; i < len(processedLine); i++ {
				if processedLine[i] == '%' && !isInString(processedLine, i) {
					if i < earliestCommentPos {
						earliestCommentPos = i
					}
					break
				}
			}
			
			// 检查感叹号注释 ! (Fortran等)
			for i := 0; i < len(processedLine); i++ {
				if processedLine[i] == '!' && !isInString(processedLine, i) {
					if i < earliestCommentPos {
						earliestCommentPos = i
					}
					break
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

// isInString 检查指定位置是否在字符串字面量内
func isInString(line string, pos int) bool {
	if pos >= len(line) {
		return false
	}
	
	var inSingleQuote, inDoubleQuote, inBacktick bool
	
	for i := 0; i <= pos && i < len(line); i++ {
		char := line[i]
		
		switch char {
		case '\'':
			if !inDoubleQuote && !inBacktick {
				// 检查前面连续反斜杠的数量
				backslashCount := 0
				for j := i - 1; j >= 0 && line[j] == '\\'; j-- {
					backslashCount++
				}
				// 如果反斜杠数量为偶数，引号未被转义
				if backslashCount%2 == 0 {
					inSingleQuote = !inSingleQuote
				}
			}
		case '"':
			if !inSingleQuote && !inBacktick {
				// 检查前面连续反斜杠的数量
				backslashCount := 0
				for j := i - 1; j >= 0 && line[j] == '\\'; j-- {
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
func processFile(filePath string) error {
	if verbose {
		fmt.Printf("处理文件: %s\n", filePath)
	}
	
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}
	
	// 检测文件类型
	fileType := detectFileType(filePath)
	if verbose {
		fmt.Printf("检测到文件类型: %s\n", fileType)
	}
	
	// 删除注释
	newContent := removeComments(string(content), fileType)
	
	// 写回文件
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}
	
	if verbose {
		fmt.Printf("✓ 已处理: %s\n", filePath)
	}
	
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
		
		// 跳过目录和隐藏文件
		if d.IsDir() || strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		
		// 检查是否为支持的文件类型
		if !isSupportedFile(path, forceMode) {
			return nil
		}
		
		// 处理文件
		if err := processFile(path); err != nil {
			fmt.Printf("❌ 处理文件失败 %s: %v\n", path, err)
			return nil // 继续处理其他文件
		}
		
		processedCount++
		return nil
	})
	
	if err != nil {
		return err
	}
	
	fmt.Printf("✅ 共处理了 %d 个文件\n", processedCount)
	return nil
}

var rootCmd = &cobra.Command{
	Use:   "fuck-comment",
	Short: "一键删注释 - 删除代码文件中的所有注释",
	Long: `fuck-comment 是一个跨平台的CLI工具，用于删除代码文件中的注释。

支持的注释格式：
  // 行注释 (C/C++, Go, Java, JavaScript等)
  /* 块注释 */ (C/C++, Go, Java, JavaScript等)
  # 井号注释 (Python, Shell, YAML等)
  -- 双破折号注释 (SQL, Haskell等)
  ; 分号注释 (Assembly, Lisp等)
  % 百分号注释 (LaTeX, MATLAB等)
  ! 感叹号注释 (Fortran等)
  <!-- HTML注释 --> (HTML, XML等)

支持100+种编程语言和文件类型

使用示例：
  fuck-comment                    # 删除当前目录及子目录所有支持文件的注释
  fuck-comment -f main.go         # 删除指定文件的注释
  fuck-comment --force            # 强制删除所有文件的注释（不限文件类型）
  fuck-comment -v                 # 显示详细处理信息
  fuck-comment --version          # 显示版本信息`,
	Run: func(cmd *cobra.Command, args []string) {
		// 显示版本信息
		if showVersion {
			fmt.Printf("fuck-comment %s\n", Version)
			fmt.Printf("构建时间: %s\n", BuildTime)
			fmt.Printf("Git提交: %s\n", GitCommit)
			return
		}
		if targetFile != "" {
			// 处理单个文件
			if !isSupportedFile(targetFile, forceMode) && !forceMode {
				fmt.Printf("❌ 不支持的文件类型: %s\n", targetFile)
				fmt.Println("使用 -force 参数可强制处理所有文件类型")
				os.Exit(1)
			}
			
			if err := processFile(targetFile); err != nil {
				fmt.Printf("❌ 处理文件失败: %v\n", err)
				os.Exit(1)
			}
			
			fmt.Printf("✅ 文件处理完成: %s\n", targetFile)
		} else {
			// 处理当前目录
			currentDir, err := os.Getwd()
			if err != nil {
				fmt.Printf("❌ 获取当前目录失败: %v\n", err)
				os.Exit(1)
			}
			
			fmt.Printf("🚀 开始处理目录: %s\n", currentDir)
			if err := processDirectory(currentDir); err != nil {
				fmt.Printf("❌ 处理目录失败: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.Flags().StringVarP(&targetFile, "file", "f", "", "指定要处理的单个文件")
	rootCmd.Flags().BoolVar(&forceMode, "force", false, "强制模式：处理所有文件类型，不限扩展名")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "显示详细处理信息")
	rootCmd.Flags().BoolVar(&showVersion, "version", false, "显示版本信息")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("❌ 执行失败: %v\n", err)
		os.Exit(1)
	}
}
