package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// 支持的编程语言文件扩展名
var supportedExtensions = map[string]bool{
	".c":     true,
	".cpp":   true,
	".cc":    true,
	".cxx":   true,
	".h":     true,
	".hpp":   true,
	".cs":    true,
	
	".java":  true,
	".scala": true,
	".kt":    true,
	".groovy": true,
	
	".js":    true,
	".jsx":   true,
	".ts":    true,
	".tsx":   true,
	".mjs":   true,
	".cjs":   true,
	".coffee": true,
	
	".go":    true,
	".rs":    true,
	".swift": true,
	".dart":  true,
	".zig":   true,
	".d":     true,
	
	".m":     true,
	".mm":    true,
	
	".py":    true,
	".rb":    true,
	".php":   true,
	".pl":    true,
	".pm":    true,
	".lua":   true,
	".tcl":   true,
	
	".sh":    true,
	".bash":  true,
	".zsh":   true,
	".fish":  true,
	".ps1":   true,
	".bat":   true,
	".cmd":   true,
	
	".hs":    true,
	".elm":   true,
	".ml":    true,
	".fs":    true,
	".fsx":   true,
	".clj":   true,
	".cljs":  true,
	".scm":   true,
	".lisp":  true,
	".lsp":   true,
	".el":    true,
	
	".r":     true,
	".R":     true,
	".jl":    true,
	".nb":    true,
	
	".html":  true,
	".htm":   true,
	".xml":   true,
	".svg":   true,
	".vue":   true,
	".svelte": true,
	".astro": true,
	
	".css":   true,
	".scss":  true,
	".sass":  true,
	".less":  true,
	".styl":  true,
	
	".twig":  true,
	".erb":   true,
	".ejs":   true,
	".hbs":   true,
	".mustache": true,
	".pug":   true,
	".jade":  true,
	".liquid": true,
	
	".yaml":  true,
	".yml":   true,
	".toml":  true,
	".ini":   true,
	".cfg":   true,
	".conf":  true,
	".json":  true,
	".jsonc": true,
	".json5": true,
	
	".md":    true,
	".markdown": true,
	".mdx":   true,
	".tex":   true,
	".rst":   true,
	".asciidoc": true,
	".adoc":  true,
	
	".sql":   true,
	".plsql": true,
	".psql":  true,
	
	".asm":   true,
	".s":     true,
	".S":     true,
	
	".v":     true,
	".vh":    true,
	".sv":    true,
	".vhd":   true,
	".vhdl":  true,
	
	".gd":    true,
	".hlsl":  true,
	".glsl":  true,
	".shader": true,
	
	".pas":   true,
	".pp":    true,
	".ada":   true,
	".adb":   true,
	".ads":   true,
	".f":     true,
	".f90":   true,
	".f95":   true,
	".for":   true,
	".cob":   true,
	".cbl":   true,
	".pro":   true,
	".erl":   true,
	".ex":    true,
	".exs":   true,
	".nim":   true,
	".cr":    true,
	".odin":  true,
	".jai":   true,
	
	".mk":    true,
	".cmake": true,
	".gradle": true,
	".sbt":   true,
	".bazel": true,
	".bzl":   true,
	".dockerfile": true,
	
	".tf":    true,
	".hcl":   true,
	".nomad": true,
	".consul": true,
	".vault": true,
}

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
	
	// 限制检查前1000字节以提高性能
	if len(content) > 1000 {
		content = content[:1000]
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
	
	// 如果MATLAB特征较多，判定为MATLAB
	if matlabCount >= 2 {
		return "matlab"
	}
	
	// 默认返回 Objective-C
	return "objc"
}

// detectRFileType 检测 R 语言文件
func detectRFileType(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "unknown"
	}
	
	if len(content) > 500 {
		content = content[:500]
	}
	
	contentStr := strings.ToLower(string(content))
	rKeywords := []string{"library(", "data.frame", "<-", "ggplot", "dplyr"}
	
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
	
	if len(content) > 200 {
		content = content[:200]
	}
	
	if strings.Contains(strings.ToLower(string(content)), ".section") ||
		strings.Contains(strings.ToLower(string(content)), ".global") {
		return "assembly"
	}
	
	return "unknown"
}

// detectDFileType 检测 D 语言文件
func detectDFileType(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "unknown"
	}
	
	if strings.Contains(string(content), "import std.") {
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
	
	if strings.Contains(strings.ToUpper(string(content)), "PROGRAM") {
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
	
	contentStr := strings.ToLower(string(content))
	if strings.Contains(contentStr, "qt") || strings.Contains(contentStr, "target") {
		return "qmake"
	}
	if strings.Contains(contentStr, "?-") || strings.Contains(contentStr, ":-") {
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
	if strings.Contains(contentStr, "?-") || strings.Contains(contentStr, ":-") {
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
	
	contentStr := strings.ToLower(string(content))
	if strings.Contains(contentStr, "program") || strings.Contains(contentStr, "begin") {
		return "pascal"
	}
	if strings.Contains(contentStr, "class") || strings.Contains(contentStr, "node") {
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
	
	if strings.Contains(string(content), "module") || strings.Contains(string(content), "endmodule") {
		return "verilog"
	}
	
	return "unknown"
}

// isSupportedFile 检查文件是否为支持的类型
func isSupportedFile(filePath string, force bool) bool {
	if force {
		return true
	}
	
	ext := strings.ToLower(filepath.Ext(filePath))
	return supportedExtensions[ext]
}
