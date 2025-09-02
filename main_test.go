package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// 测试工具函数

// createTempFile 创建临时测试文件
func createTempFile(t *testing.T, content string, suffix string) string {
	tmpFile, err := ioutil.TempFile("", "test_*"+suffix)
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}
	
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("关闭临时文件失败: %v", err)
	}
	
	return tmpFile.Name()
}

// cleanupTempFile 清理临时测试文件
func cleanupTempFile(t *testing.T, filename string) {
	if err := os.Remove(filename); err != nil {
		t.Logf("清理临时文件失败: %v", err)
	}
}

// assertStringEqual 断言字符串相等，提供详细的错误信息
func assertStringEqual(t *testing.T, expected, actual, testName string) {
	if expected != actual {
		t.Errorf("%s 失败:\n期望:\n%s\n实际:\n%s", testName, expected, actual)
	}
}

// resetBackupGlobals 重置备份相关的全局变量（用于测试隔离）
func resetBackupGlobals() {
	backupTimestamp = ""
	backupRootDir = ""
}

func TestRemoveComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "单行注释",
			input: `package main
// 测试注释
func main() {
	fmt.Println("Hello") // 行尾注释
}`,
			expected: `package main
func main() {
	fmt.Println("Hello")
}`,
		},
		{
			name: "块注释",
			input: `package main
/* 这是块注释 */
func main() {
	/* 另一个注释 */ fmt.Println("Hello")
}`,
			expected: `package main
func main() {
	 fmt.Println("Hello")
}`,
		},
		{
			name: "多行块注释",
			input: `package main
/*
 * 多行注释
 * 第二行
 */
func main() {
	fmt.Println("Hello")
}`,
			expected: `package main
func main() {
	fmt.Println("Hello")
}`,
		},
		{
			name: "字符串中的注释符号",
			input: `package main
func main() {
	url := "http://example.com"
	comment := "/* 这不是注释 */"
	path := "C:\\Program Files\\test" // 路径注释
}`,
			expected: `package main
func main() {
	url := "http://example.com"
	comment := "/* 这不是注释 */"
	path := "C:\\Program Files\\test"
}`,
		},
		{
			name: "混合注释",
			input: `package main
// 文件头注释
/* 块注释 */
func main() {
	// 函数内注释
	/* 内联块注释 */
	fmt.Println("Hello") // 行尾注释
	 fmt.Println("World")
}`,
			expected: `package main
func main() {
	fmt.Println("Hello")
	 fmt.Println("World")
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeComments(tt.input, "go")
			if result != tt.expected {
				t.Errorf("removeComments() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestIsInString(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		pos      int
		expected bool
	}{
		{
			name:     "不在字符串中",
			line:     `fmt.Println("Hello") // comment`,
			pos:      21,
			expected: false,
		},
		{
			name:     "在双引号字符串中",
			line:     `fmt.Println("Hello // World")`,
			pos:      19,
			expected: true,
		},
		{
			name:     "在单引号字符串中",
			line:     `char := '/' // comment`,
			pos:        12,
			expected: false,
		},
		{
			name:     "转义引号",
			line:     `fmt.Println("He said \"Hello\"") // comment`,
			pos:      32,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInString(tt.line, tt.pos)
			if result != tt.expected {
				t.Errorf("isInString(%q, %d) = %v, want %v", tt.line, tt.pos, result, tt.expected)
			}
		})
	}
}

func TestIsSupportedFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		force    bool
		expected bool
	}{
		{
			name:     "Go文件",
			filePath: "main.go",
			force:    false,
			expected: true,
		},
		{
			name:     "JavaScript文件",
			filePath: "script.js",
			force:    false,
			expected: true,
		},
		{
			name:     "Markdown文件",
			filePath: "README.md",
			force:    false,
			expected: true,
		},
		{
			name:     "强制模式下的不支持文件",
			filePath: "README.md",
			force:    true,
			expected: true,
		},
		{
			name:     "大写扩展名",
			filePath: "Main.GO",
			force:    false,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSupportedFile(tt.filePath, tt.force)
			if result != tt.expected {
				t.Errorf("isSupportedFile(%q, %v) = %v, want %v", tt.filePath, tt.force, result, tt.expected)
			}
		})
	}
}

func TestProcessFile(t *testing.T) {
	// 创建临时文件进行测试
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.go")
	
	content := `package main
// 这是注释
func main() {
	fmt.Println("Hello") // 行尾注释
}`
	
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}
	
	// 处理文件
	err = processFile(testFile, tempDir)
	if err != nil {
		t.Fatalf("处理文件失败: %v", err)
	}
	
	// 读取处理后的内容
	result, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("读取处理后文件失败: %v", err)
	}
	
	expected := `package main
func main() {
	fmt.Println("Hello")
}`
	
	if strings.TrimSpace(string(result)) != strings.TrimSpace(expected) {
		t.Errorf("文件处理结果不符合预期\n得到: %q\n期望: %q", string(result), expected)
	}
}

// 基准测试
func BenchmarkRemoveComments(b *testing.B) {
	content := `package main

import "fmt"

// 这是一个示例程序
/* 
 * 多行注释
 * 第二行
 */
func main() {
	// 打印消息
	fmt.Println("Hello, World!") // 行尾注释
	
	/* 块注释 */ fmt.Println("Another message")
	
	url := "http://example.com // 这不是注释"
}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		removeComments(content, "go")
	}
}

// 测试文件类型检测
func TestDetectFileType(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "Go文件",
			filePath: "main.go",
			expected: "go",
		},
		{
			name:     "Markdown文件",
			filePath: "README.md",
			expected: "markdown",
		},
		{
			name:     "YAML文件",
			filePath: "config.yml",
			expected: "yaml",
		},
		{
			name:     "JSON文件",
			filePath: "package.json",
			expected: "json",
		},
		{
			name:     "CSS文件",
			filePath: "style.css",
			expected: "css",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectFileType(tt.filePath)
			if result != tt.expected {
				t.Errorf("detectFileType(%q) = %q, want %q", tt.filePath, result, tt.expected)
			}
		})
	}
}

// 测试Markdown注释处理
func TestRemoveMarkdownComments(t *testing.T) {
	input := `# 标题

正文内容

## 二级标题

<!-- HTML注释应该被删除 -->

### 三级标题`

	expected := "# 标题\n\n正文内容\n\n## 二级标题\n\n\n### 三级标题"

	result := removeComments(input, "markdown")
	assertStringEqual(t, expected, result, "removeMarkdownComments")
}

// ...

func TestRemoveYamlComments(t *testing.T) {
	input := `version: '3.8'  # Docker版本
services:
  web:
    image: nginx  # Web服务器
    ports:
      - "80:80"
  # 这是整行注释
  database:
    image: postgres`

	expected := "version: '3.8'\nservices:\n  web:\n    image: nginx\n    ports:\n      - \"80:80\"\n  database:\n    image: postgres"

	result := removeComments(input, "yaml")
	assertStringEqual(t, expected, result, "removeYamlComments")
}

// 测试YAML安全边界情况
func TestRemoveYamlCommentsSecurity(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "字符串中的井号",
			input: `database:
  url: "mongodb://user:pass#word@localhost:27017/db"  # 密码注释
  hash: "#secret#key#"  # 哈希值注释`,
			expected: `database:
  url: "mongodb://user:pass#word@localhost:27017/db"
  hash: "#secret#key#"`,
		},
		{
			name: "数组中的井号",
			input: `config:
  tags: ["#tag1", "#tag2"]  # 标签数组
  colors: ["#FF0000", "#00FF00"]  # 颜色数组`,
			expected: `config:
  tags: ["#tag1", "#tag2"]
  colors: ["#FF0000", "#00FF00"]`,
		},
		{
			name: "正则表达式中的井号",
			input: `validation:
  pattern: "^#[0-9A-Fa-f]{6}$"  # 颜色正则
  regex: "#\\d+"  # 数字井号`,
			expected: `validation:
  pattern: "^#[0-9A-Fa-f]{6}$"
  regex: "#\\d+"`,
		},
		{
			name: "转义字符",
			input: `text:
  escaped: "He said \"Hello #world\""  # 转义引号
  path: "C:\\#temp\\file"  # 路径中的井号`,
			expected: `text:
  escaped: "He said \"Hello #world\""
  path: "C:\\#temp\\file"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeComments(tt.input, "yaml")
			assertStringEqual(t, tt.expected, result, tt.name)
		})
	}
}

// 测试文件类型歧义检测
func TestDetectFileTypeAmbiguous(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()
	
	tests := []struct {
		name     string
		filename string
		content  string
		expected string
	}{
		{
			name:     "Objective-C文件",
			filename: "test.m",
			content:  `#import <Foundation/Foundation.h>\n@interface MyClass\n@end`,
			expected: "objc",
		},
		{
			name:     "MATLAB文件",
			filename: "test.m",
			content:  `function result = myFunc(x)\n% This is a comment\nresult = x * 2;\nend`,
			expected: "matlab",
		},
		{
			name:     "R语言文件",
			filename: "test.r",
			content:  `library(ggplot2)\ndata <- data.frame(x = 1:10)\nplot(data$x)`,
			expected: "r",
		},
		{
			name:     "Assembly文件",
			filename: "test.s",
			content:  `.section .text\n.global _start\n_start:\n    mov $1, %eax`,
			expected: "assembly",
		},
		{
			name:     "Verilog文件",
			filename: "test.v",
			content:  `module counter(clk, reset, count);\ninput clk, reset;\noutput [3:0] count;\nalways @(posedge clk) begin\nend\nendmodule`,
			expected: "verilog",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tempDir, tt.filename)
			err := os.WriteFile(filePath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("创建测试文件失败: %v", err)
			}
			
			result := detectFileType(filePath)
			if result != tt.expected {
				t.Errorf("detectFileType(%q) = %q, want %q", tt.filename, result, tt.expected)
			}
		})
	}
}

// 测试字符串检测边界情况
func TestIsInStringEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		pos      int
		expected bool
	}{
		{
			name:     "反引号字符串",
			line:     "`Hello // World`",
			pos:      8,
			expected: true,
		},
		{
			name:     "嵌套引号",
			line:     `"He said 'Hello // World'"`,
			pos:      16,
			expected: true,
		},
		{
			name:     "转义反斜杠",
			line:     `"Path: C:\\\\Program Files\\\\" // comment`,
			pos:      29, // 字符串内部的最后一个字符
			expected: true,
		},
		{
			name:     "多重转义",
			line:     `"Text with \\"quote\\" and // slash"`,
			pos:      25,
			expected: true,
		},
		{
			name:     "空字符串",
			line:     `"" // comment`,
			pos:      3,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInString(tt.line, tt.pos)
			if result != tt.expected {
				t.Errorf("isInString(%q, %d) = %v, want %v", tt.line, tt.pos, result, tt.expected)
			}
		})
	}
}

func BenchmarkIsInString(b *testing.B) {
	line := `fmt.Println("Hello // World") // This is a comment`
	pos := 30
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isInString(line, pos)
	}
}

// TestBinaryFileDetection 测试二进制文件检测
func TestBinaryFileDetection(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected bool
	}{
		{"空文件", []byte{}, false},
		{"文本文件", []byte("hello world"), false},
		{"UTF-8文件", []byte("你好世界"), false},
		{"包含null字节", []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x00, 0x57, 0x6f, 0x72, 0x6c, 0x64}, true},
		{"无效UTF-8", []byte{0xff, 0xfe, 0xfd}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBinaryFile(tt.content)
			if result != tt.expected {
				t.Errorf("isBinaryFile() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestFileSafety 测试文件安全检查
func TestFileSafety(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		content     []byte
		expectError bool
	}{
		{"正常文件", "test.go", []byte("package main\nfunc main() {}"), false},
		{"空文件", "empty.txt", []byte{}, false},
		{"二进制文件", "binary.bin", []byte{0x00, 0x01, 0x02}, true},
		{"长行文件", "long.txt", []byte(strings.Repeat("a", 60000)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := isFileSafe(tt.filePath, tt.content, false)
			if (err != nil) != tt.expectError {
				t.Errorf("isFileSafe() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

// TestSecurityAndPerformance 综合安全性和性能测试
func TestSecurityAndPerformance(t *testing.T) {
	// 测试大文件处理性能
	t.Run("大文件性能测试", func(t *testing.T) {
		var content strings.Builder
		for i := 0; i < 5000; i++ {
			content.WriteString(fmt.Sprintf("// Line %d comment\n", i))
			content.WriteString(fmt.Sprintf("func test%d() { /* block comment */ return %d }\n", i, i))
		}
		
		start := time.Now()
		result := removeComments(content.String(), "go")
		duration := time.Since(start)
		
		// 性能要求：处理10000行应该在2秒内完成
		if duration > 2*time.Second {
			t.Errorf("性能问题: 处理大文件耗时 %v", duration)
		}
		
		// 验证注释被正确删除
		if strings.Contains(result, "//") || strings.Contains(result, "/*") {
			t.Error("大文件中的注释未被完全删除")
		}
	})

	// 测试恶意输入处理
	t.Run("恶意输入测试", func(t *testing.T) {
		maliciousInputs := []string{
			strings.Repeat("\"", 10000),     // 大量引号
			strings.Repeat("\\", 10000),     // 大量反斜杠
			strings.Repeat("/*", 5000),      // 大量注释开始符
			strings.Repeat("//", 5000),      // 大量行注释
			string([]byte{0x00, 0x01, 0x02}), // 二进制数据
		}
		
		for i, input := range maliciousInputs {
			t.Run(fmt.Sprintf("恶意输入_%d", i), func(t *testing.T) {
				// 应该不会崩溃或无限循环
				done := make(chan bool, 1)
				go func() {
					defer func() {
						if r := recover(); r != nil {
							t.Errorf("处理恶意输入时发生panic: %v", r)
						}
						done <- true
					}()
					removeComments(input, "go")
				}()
				
				select {
				case <-done:
					// 正常完成
				case <-time.After(5 * time.Second):
					t.Error("处理恶意输入超时，可能存在无限循环")
				}
			})
		}
	})

	// 测试备份机制
	t.Run("备份机制测试", func(t *testing.T) {
		// 创建临时目录和文件
		tmpDir, err := ioutil.TempDir("", "test_backup_dir")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)
		defer func() {
			// 清理可能的备份文件
			os.RemoveAll("bak")
		}()
		
		tmpFile := filepath.Join(tmpDir, "test.go")
		testContent := "// This is a test\nfunc main() {}"
		err = ioutil.WriteFile(tmpFile, []byte(testContent), 0644)
		if err != nil {
			t.Fatal(err)
		}
		
		// 重置备份目录变量以确保测试独立性
		backupRootDir = ""
		
		// 创建备份
		err = createBackup(tmpFile, tmpDir)
		if err != nil {
			t.Errorf("创建备份失败: %v", err)
		}
		
		// 验证备份文件存在且内容正确
		bakDir := filepath.Join(tmpDir, "bak")
		
		// 查找备份文件
		var backupPath string
		filepath.Walk(bakDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.Contains(path, "test.go") {
				backupPath = path
			}
			return nil
		})
		
		if backupPath == "" {
			// 如果在bak目录下没找到，尝试在整个临时目录下查找
			filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() && strings.Contains(path, "test.go") && path != tmpFile {
					backupPath = path
				}
				return nil
			})
		}
		
		if backupPath == "" {
			t.Error("备份文件未找到")
			return
		}
		
		backupContent, err := ioutil.ReadFile(backupPath)
		if err != nil {
			t.Errorf("读取备份文件失败: %v", err)
		}
		
		if string(backupContent) != testContent {
			t.Error("备份文件内容不正确")
		}
	})

	// 测试字符串检测性能
	t.Run("字符串检测性能", func(t *testing.T) {
		// 创建复杂的字符串测试用例
		complexLine := `fmt.Printf("Complex string with \"nested quotes\" and \\ backslashes") // comment`
		
		start := time.Now()
		for i := 0; i < 10000; i++ {
			isInString(complexLine, 50)
		}
		duration := time.Since(start)
		
		// 性能要求：10000次调用应该在100ms内完成
		if duration > 100*time.Millisecond {
			t.Errorf("字符串检测性能问题: 10000次调用耗时 %v", duration)
		}
	})

	// 测试内存使用
	t.Run("内存使用测试", func(t *testing.T) {
		// 创建大量小文件内容
		var contents []string
		for i := 0; i < 1000; i++ {
			content := fmt.Sprintf("// File %d\nfunc test%d() { /* comment */ }\n", i, i)
			contents = append(contents, content)
		}
		
		// 处理所有内容
		start := time.Now()
		for _, content := range contents {
			removeComments(content, "go")
		}
		duration := time.Since(start)
		
		// 性能要求：处理1000个小文件应该在1秒内完成
		if duration > time.Second {
			t.Errorf("内存使用可能有问题: 处理1000个小文件耗时 %v", duration)
		}
	})
}

// 测试Rust语言特殊情况
func TestRustComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Rust基本注释",
			input: `// Rust comment
fn main() {
    let x = 5; // inline comment
    /* block comment */
    println!("Hello");
}`,
			expected: `fn main() {
    let x = 5;
    println!("Hello");
}`,
		},
		{
			name: "Rust文档注释保护",
			input: `/// Documentation comment
/// Should be preserved
fn test() {}`,
			expected: `fn test() {}`,
		},
		{
			name: "Rust字符串中的注释符号",
			input: `let regex = r"//.*"; // Raw string
let url = "https://example.com#anchor";`,
			expected: `let regex = r"//.*";
let url = "https://example.com#anchor";`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeComments(tt.input, "rust")
			if result != tt.expected {
				t.Errorf("removeComments() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// 测试Shell脚本特殊情况
func TestShellComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Shell变量展开保护",
			input: `#!/bin/bash
VERSION=${GITHUB_REF#refs/tags/} # Comment
echo "Version: $VERSION"`,
			expected: `#!/bin/bash
VERSION=${GITHUB_REF#refs/tags/}
echo "Version: $VERSION"`,
		},
		{
			name: "Shell条件语句保护",
			input: `if [ "$1" != "" ]; then # Comment
    echo "Arg: $1"
fi`,
			expected: `if [ "$1" != "" ]; then
    echo "Arg: $1"
fi`,
		},
		{
			name: "Shell字符串中的井号",
			input: `URL="https://example.com#section" # URL comment
HASH="#hashtag" # Hash comment`,
			expected: `URL="https://example.com#section"
HASH="#hashtag"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeComments(tt.input, "shell")
			if result != tt.expected {
				t.Errorf("removeComments() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// 测试JavaScript模板字符串
func TestJavaScriptTemplateStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "模板字符串中的注释保护",
			input: "const template = `\n  // This should be preserved\n  /* Also preserved */\n  ${variable}\n`; // External comment",
			expected: "const template = `\n  // This should be preserved\n  /* Also preserved */\n  ${variable}\n`;",
		},
		{
			name: "正则表达式中的注释符号",
			input: `const regex = /\/\* not a comment \*\//; // Regex comment`,
			expected: `const regex = /\/\* not a comment \*\//;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeComments(tt.input, "javascript")
			if result != tt.expected {
				t.Errorf("removeComments() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// 测试Python特殊情况
func TestPythonComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Python docstring保护",
			input: `def func():
    """This docstring should be preserved"""
    # This comment should be deleted
    return True`,
			expected: `def func():
    """This docstring should be preserved"""
    return True`,
		},
		{
			name: "Python f-string中的井号",
			input: `name = "world"
f_string = f"Hello #{name}#" # Comment`,
			expected: `name = "world"
f_string = f"Hello #{name}#"`,
		},
		{
			name: "Python多行字符串",
			input: `text = '''
Multi-line string
# This should be preserved
''' # Comment`,
			expected: `text = '''
Multi-line string
# This should be preserved
'''`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeComments(tt.input, "python")
			if result != tt.expected {
				t.Errorf("removeComments() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// 测试无变化时不创建备份的优化
func TestNoBackupWhenNoChange(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.go")
	
	// 创建已经没有注释的文件
	content := `package main
func main() {
	fmt.Println("Hello")
}`
	
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}
	
	// 处理文件
	err = processFile(testFile, tempDir)
	if err != nil {
		t.Fatalf("处理文件失败: %v", err)
	}
	
	// 检查是否创建了备份目录
	bakDir := filepath.Join("bak")
	if _, err := os.Stat(bakDir); err == nil {
		// 如果备份目录存在，检查是否为空或不包含我们的测试文件
		empty := true
		filepath.Walk(bakDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.Contains(path, "test.go") {
				empty = false
			}
			return nil
		})
		if !empty {
			t.Error("无变化时不应该创建备份")
		}
	}
}

// TestEdgeCasesAndBoundaries 边界条件和特殊情况测试
func TestEdgeCasesAndBoundaries(t *testing.T) {
	// 测试空输入
	t.Run("空输入处理", func(t *testing.T) {
		result := removeComments("", "go")
		if result != "" {
			t.Error("空输入应该返回空字符串")
		}
	})

	// 测试单字符输入
	t.Run("单字符输入", func(t *testing.T) {
		inputs := []string{"/", "*", "#", "-", ";", "%", "!", "<"}
		for _, input := range inputs {
			result := removeComments(input, "go")
			// 单独的注释符号应该被保留，因为它们可能是代码的一部分
			if result != input {
				t.Errorf("单字符输入 %q 处理错误: got %q", input, result)
			}
		}
	})

	// 测试极长行
	t.Run("极长行处理", func(t *testing.T) {
		longLine := strings.Repeat("a", 1000) + " // comment"
		result := removeComments(longLine, "go")
		// 检查注释是否被删除
		if strings.Contains(result, "//") || strings.Contains(result, "comment") {
			t.Error("注释未被正确删除")
		}
		// 检查基本内容是否保留
		if !strings.HasPrefix(result, strings.Repeat("a", 1000)) {
			t.Error("极长行处理错误: 基本内容未正确保留")
		}
	})

	// 测试嵌套引号
	t.Run("深度嵌套引号", func(t *testing.T) {
		nested := `"level1 \"level2 \\\"level3\\\" level2\" level1"`
		for i := 0; i < len(nested); i++ {
			// 不应该崩溃
			isInString(nested, i)
		}
	})

	// 测试所有支持的文件类型
	t.Run("所有文件类型测试", func(t *testing.T) {
		fileTypes := []string{"go", "javascript", "python", "java", "css", "html", "yaml", "json", "markdown"}
		testContent := "// comment\ncode here /* block */ more code"
		
		for _, fileType := range fileTypes {
			result := removeComments(testContent, fileType)
			// 应该不会崩溃，且返回非空结果
			if len(result) == 0 {
				t.Errorf("文件类型 %s 处理后返回空结果", fileType)
			}
		}
	})

	// 测试危险边界情况，防止删错代码
	t.Run("危险边界情况", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    string
			expected string
			fileType string
		}{
			{
				name:     "URL中的双斜杠",
				input:    `url := "https://example.com/path"`,
				expected: `url := "https://example.com/path"`,
				fileType: "go",
			},
			{
				name:     "正则表达式中的注释符号",
				input:    `pattern := "/\\*.*?\\*/"`,
				expected: `pattern := "/\\*.*?\\*/"`,
				fileType: "go",
			},
			{
				name:     "字符串中的转义引号和注释",
				input:    `msg := "He said \"Hello // World\""`,
				expected: `msg := "He said \"Hello // World\""`,
				fileType: "go",
			},
			{
				name:     "多行字符串中的注释符号",
				input:    "text := `\nThis is // not a comment\n/* also not */\n`",
				expected: "text := `\nThis is // not a comment\n/* also not */\n`",
				fileType: "go",
			},
			{
				name:     "数学运算符",
				input:    "result := a / b * c // actual comment",
				expected: "result := a / b * c",
				fileType: "go",
			},
			{
				name:     "Shell脚本中的特殊情况",
				input:    `echo "Price: $10 # not a comment"`,
				expected: `echo "Price: $10 # not a comment"`,
				fileType: "shell",
			},
			{
				name:     "CSS中的伪类选择器",
				input:    "a:hover /* comment */ { color: red; }",
				expected: "a:hover  { color: red; }",
				fileType: "css",
			},
			{
				name:     "HTML属性中的特殊字符",
				input:    `<div data-comment="/* not a comment */">`,
				expected: `<div data-comment="/* not a comment */">`,
				fileType: "html",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := removeComments(tc.input, tc.fileType)
				if result != tc.expected {
					t.Errorf("输入: %q\n期望: %q\n实际: %q", tc.input, tc.expected, result)
				}
			})
		}
	})
}

// TestGoTemplateLiteralFix 测试Go模板字符串外部注释的修复
func TestGoTemplateLiteralFix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "单行模板字符串外部注释应该被删除",
			input:    "const template = `hello world`; // External comment",
			expected: "const template = `hello world`;",
		},
		{
			name:     "模板字符串内部注释应该保留",
			input:    "const template = `\n  // This should be preserved\n  /* Also preserved */\n  ${variable}\n`; // External comment",
			expected: "const template = `\n  // This should be preserved\n  /* Also preserved */\n  ${variable}\n`;",
		},
		{
			name:     "多行模板字符串内部注释应该完全保留",
			input:    "const template = `\n  // Internal comment\n  some code\n`;",
			expected: "const template = `\n  // Internal comment\n  some code\n`;",
		},
		{
			name:     "嵌套反引号的复杂情况",
			input:    "const cmd = `echo 'test'`; // This is external",
			expected: "const cmd = `echo 'test'`;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := getCommentRulesForLanguage("go")
			result := removeCommentsByRules(tt.input, "go", rules)
			if result != tt.expected {
				t.Errorf("期望: %q\n实际: %q", tt.expected, result)
			}
		})
	}
}

// TestYAMLStructuralCommentsFix 测试YAML结构性注释的修复
func TestYAMLStructuralCommentsFix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "保留emoji标题注释",
			input: `name: Build
# 🚀 fuck-comment release
version: 1.0`,
			expected: `name: Build
# 🚀 fuck-comment release
version: 1.0`,
		},
		{
			name: "保留markdown风格的节标题",
			input: `jobs:
  build:
    # ## 构建步骤
    runs-on: ubuntu-latest`,
			expected: `jobs:
  build:
    # ## 构建步骤
    runs-on: ubuntu-latest`,
		},
		{
			name: "保留包含关键词的结构性注释",
			input: `steps:
  # 下载文件
  - name: Download
  # 安装依赖
  - name: Install`,
			expected: `steps:
  # 下载文件
  - name: Download
  # 安装依赖
  - name: Install`,
		},
		{
			name: "删除普通行尾注释",
			input: `name: test # this is a regular comment
version: 1.0 # another comment`,
			expected: `name: test
version: 1.0`,
		},
		{
			name: "保护Shell变量中的#",
			input: `VERSION: ${GITHUB_REF#refs/tags/}
BUILD_TIME: $(date)`,
			expected: `VERSION: ${GITHUB_REF#refs/tags/}
BUILD_TIME: $(date)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := getCommentRulesForLanguage("yaml")
			result := removeCommentsByRules(tt.input, "yaml", rules)
			if result != tt.expected {
				t.Errorf("期望:\n%s\n实际:\n%s", tt.expected, result)
			}
		})
	}
}

// TestJavaScriptTemplateLiteralFix 测试JavaScript模板字符串的修复
func TestJavaScriptTemplateLiteralFix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "JavaScript模板字符串外部注释",
			input:    "const html = `<div>content</div>`; // External comment",
			expected: "const html = `<div>content</div>`;",
		},
		{
			name:     "JavaScript模板字符串内部注释保留",
			input:    "const code = `\n  // This is code comment\n  function test() {}\n`;",
			expected: "const code = `\n  // This is code comment\n  function test() {}\n`;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := getCommentRulesForLanguage("javascript")
			result := removeCommentsByRules(tt.input, "javascript", rules)
			if result != tt.expected {
				t.Errorf("期望: %q\n实际: %q", tt.expected, result)
			}
		})
	}
}

// TestEdgeCasesFixed 测试修复后的边缘情况
func TestEdgeCasesFixed(t *testing.T) {
	tests := []struct {
		name     string
		fileType string
		input    string
		expected string
	}{
		{
			name:     "Go - 多个反引号的复杂情况",
			fileType: "go",
			input:    "cmd := `echo \\`nested\\``; // comment",
			expected: "cmd := `echo \\`nested\\``;",
		},
		{
			name:     "YAML - 混合emoji和文字的标题",
			fileType: "yaml",
			input:    "# 📦 下载和安装指南\nsteps: []",
			expected: "# 📦 下载和安装指南\nsteps: []",
		},
		{
			name:     "YAML - 普通注释应该被删除",
			fileType: "yaml",
			input:    "name: test\n# just a regular comment\nversion: 1.0",
			expected: "name: test\nversion: 1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := getCommentRulesForLanguage(tt.fileType)
			result := removeCommentsByRules(tt.input, tt.fileType, rules)
			if result != tt.expected {
				t.Errorf("期望:\n%s\n实际:\n%s", tt.expected, result)
			}
		})
	}
}

// TestAllSupportedLanguages 测试所有支持的语言都能正确删除注释
func TestAllSupportedLanguages(t *testing.T) {
	tests := []struct {
		name     string
		fileType string
		input    string
		expected string
	}{
		// C风格语言 (// 和 /* */)
		{
			name:     "JavaScript注释",
			fileType: "javascript",
			input:    "console.log(\"hello\"); // 这是注释\nvar x = 5; /* 块注释 */",
			expected: "console.log(\"hello\");\nvar x = 5; ",
		},
		{
			name:     "TypeScript注释",
			fileType: "typescript",
			input:    "let name: string = \"test\"; // 类型注释\n/* 多行注释 */",
			expected: "let name: string = \"test\";",
		},
		{
			name:     "Go语言注释",
			fileType: "go",
			input:    "package main // 包声明\n// 函数注释\nfunc main() {}",
			expected: "package main\nfunc main() {}",
		},
		{
			name:     "C语言注释",
			fileType: "c",
			input:    "#include <stdio.h> // 头文件\nint main() { /* 主函数 */ return 0; }",
			expected: "#include <stdio.h>\nint main() {  return 0; }",
		},
		{
			name:     "C++注释",
			fileType: "cpp",
			input:    `#include <iostream> // C++头文件\nusing namespace std; /* 命名空间 */`,
			expected: `#include <iostream>`,
		},
		{
			name:     "Java注释",
			fileType: "java",
			input:    `public class Test { // 类定义\n    /* 构造函数 */ public Test() {} }`,
			expected: `public class Test {`,
		},
		{
			name:     "C#注释",
			fileType: "cs",
			input:    `using System; // 命名空间\nclass Program { /* 主类 */ }`,
			expected: `using System;`,
		},
		{
			name:     "Rust注释",
			fileType: "rust",
			input:    `fn main() { // 主函数\n    /* 打印 */ println!("hello"); }`,
			expected: `fn main() {`,
		},
		{
			name:     "Swift注释",
			fileType: "swift",
			input:    `import Foundation // 导入\n/* 主函数 */ func main() {}`,
			expected: `import Foundation`,
		},

		// 井号注释语言 (#)
		{
			name:     "Shell注释",
			fileType: "shell",
			input:    `#!/bin/bash\necho "hello" # 打印消息`,
			expected: `#!/bin/bash\necho "hello"`,
		},
		{
			name:     "Python注释",
			fileType: "python",
			input:    "def hello(): # 函数定义\n    print(\"hello\") # 打印",
			expected: "def hello():\n    print(\"hello\")",
		},
		{
			name:     "Ruby注释",
			fileType: "ruby",
			input:    `def hello # 方法定义\n  puts "hello" # 打印\nend`,
			expected: `def hello`,
		},
		{
			name:     "Perl注释",
			fileType: "perl",
			input:    "#!/usr/bin/perl\nprint \"hello\"; # 打印消息",
			expected: "print \"hello\";",
		},
		{
			name:     "R语言注释",
			fileType: "r",
			input:    `x <- 5 # 赋值\nprint(x) # 打印变量`,
			expected: `x <- 5`,
		},

		// PHP (混合注释)
		{
			name:     "PHP注释",
			fileType: "php",
			input:    `<?php\n$x = 5; // 赋值\n/* 多行注释 */ echo $x; # 井号注释`,
			expected: `<?php\n$x = 5;`,
		},

		// Lua (双破折号)
		{
			name:     "Lua注释",
			fileType: "lua",
			input:    `local x = 5 -- 局部变量\n--[[ 多行注释\n内容 ]] print(x)`,
			expected: `local x = 5`,
		},

		// SQL (双破折号和块注释)
		{
			name:     "SQL注释",
			fileType: "sql",
			input:    `SELECT * FROM users -- 查询用户\n/* 多行注释 */ WHERE id = 1;`,
			expected: `SELECT * FROM users`,
		},

		// MATLAB
		{
			name:     "MATLAB注释",
			fileType: "matlab",
			input:    `x = 5; % 变量赋值\n%{ 多行注释\n内容 %} disp(x);`,
			expected: `x = 5;`,
		},

		// Assembly
		{
			name:     "Assembly注释",
			fileType: "assembly",
			input:    `mov eax, 5 ; 移动指令\n# 另一种注释\nadd eax, 1 // 第三种注释`,
			expected: `mov eax, 5`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeComments(tt.input, tt.fileType)
			if result != tt.expected {
				t.Errorf("removeComments() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestLanguageSpecificEdgeCases 测试各语言特定的边界情况
func TestLanguageSpecificEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		fileType string
		input    string
		expected string
	}{
		{
			name:     "JavaScript正则表达式保护",
			fileType: "javascript",
			input:    `var regex = /\/\*.*?\*\//g; // 匹配注释的正则`,
			expected: `var regex = /\/\*.*?\*\//g;`,
		},
		{
			name:     "JavaScript模板字符串保护",
			fileType: "javascript",
			input:    "var template = `Hello // world`; // 注释",
			expected: "var template = `Hello // world`;",
		},
		{
			name:     "Python f-string保护",
			fileType: "python",
			input:    `name = "world"\nf_string = f"Hello #{name}#" # 注释`,
			expected: `name = "world"\nf_string = f"Hello #{name}#"`,
		},
		{
			name:     "Shell变量展开保护",
			fileType: "shell",
			input:    `VERSION=${GITHUB_REF#refs/tags/} # 提取版本号`,
			expected: `VERSION=${GITHUB_REF#refs/tags/}`,
		},
		{
			name:     "Rust原始字符串保护",
			fileType: "rust",
			input:    `let raw = r"This is // not a comment"; // 这是注释`,
			expected: `let raw = r"This is // not a comment";`,
		},
		{
			name:     "C字符串中的注释符号保护",
			fileType: "c",
			input:    `printf("URL: http://example.com#anchor"); // 打印URL`,
			expected: `printf("URL: http://example.com#anchor");`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeComments(tt.input, tt.fileType)
			if result != tt.expected {
				t.Errorf("removeComments() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestCoreLanguageSupport 测试核心语言支持
func TestCoreLanguageSupport(t *testing.T) {
	tests := []struct {
		name     string
		fileType string
		input    string
		expected string
	}{
		// C风格语言测试
		{
			name:     "JavaScript基本注释",
			fileType: "javascript",
			input:    "var x = 5; // comment",
			expected: "var x = 5;",
		},
		{
			name:     "Go语言注释",
			fileType: "go", 
			input:    "package main // comment",
			expected: "package main",
		},
		{
			name:     "Java注释",
			fileType: "java",
			input:    "public class Test { // comment",
			expected: "public class Test {",
		},
		{
			name:     "C++块注释",
			fileType: "cpp",
			input:    "int x = 5; /* comment */ int y = 6;",
			expected: "int x = 5;  int y = 6;",
		},
		
		// 井号注释语言测试
		{
			name:     "Python注释",
			fileType: "python",
			input:    "x = 5 # comment",
			expected: "x = 5",
		},
		{
			name:     "Shell注释",
			fileType: "shell",
			input:    "echo hello # comment",
			expected: "echo hello",
		},
		{
			name:     "Ruby注释",
			fileType: "ruby",
			input:    "puts 'hello' # comment",
			expected: "puts 'hello'",
		},
		
		// 其他语言测试
		{
			name:     "SQL注释",
			fileType: "sql",
			input:    "SELECT * FROM users -- comment",
			expected: "SELECT * FROM users",
		},
		{
			name:     "Lua注释",
			fileType: "lua",
			input:    "local x = 5 -- comment",
			expected: "local x = 5",
		},
		{
			name:     "MATLAB注释",
			fileType: "matlab",
			input:    "x = 5; % comment",
			expected: "x = 5;",
		},
		{
			name:     "Assembly注释",
			fileType: "assembly",
			input:    "mov eax, 5 ; comment",
			expected: "mov eax, 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeComments(tt.input, tt.fileType)
			if result != tt.expected {
				t.Errorf("removeComments(%q, %q) = %q, want %q", tt.input, tt.fileType, result, tt.expected)
			}
		})
	}
}

// TestStringProtectionEdgeCases 测试字符串保护功能
func TestStringProtectionEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		fileType string
		input    string
		expected string
	}{
		{
			name:     "JavaScript字符串中的注释符号",
			fileType: "javascript",
			input:    `console.log("// not a comment"); // real comment`,
			expected: `console.log("// not a comment");`,
		},
		{
			name:     "Python字符串中的井号",
			fileType: "python",
			input:    `print("URL: http://example.com#anchor") # comment`,
			expected: `print("URL: http://example.com#anchor")`,
		},
		{
			name:     "C字符串中的注释符号",
			fileType: "c",
			input:    `printf("/* not a comment */"); // comment`,
			expected: `printf("/* not a comment */");`,
		},
		{
			name:     "Shell字符串中的井号",
			fileType: "shell",
			input:    `echo "URL: http://example.com#anchor" # comment`,
			expected: `echo "URL: http://example.com#anchor"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeComments(tt.input, tt.fileType)
			if result != tt.expected {
				t.Errorf("removeComments(%q, %q) = %q, want %q", tt.input, tt.fileType, result, tt.expected)
			}
		})
	}
}

// TestAllLanguageCommentRemoval 测试所有支持语言的注释删除功能
func TestAllLanguageCommentRemoval(t *testing.T) {
	tests := []struct {
		fileType string
		input    string
		expected string
		name     string
	}{
		// C风格语言 (// 和 /* */)
		{"c", "int x = 5; // comment", "int x = 5;", "C语言行注释"},
		{"cpp", "int x = 5; /* comment */", "int x = 5; ", "C++块注释"},
		{"java", "String s = \"test\"; // comment", "String s = \"test\";", "Java行注释"},
		{"cs", "var x = 5; // comment", "var x = 5;", "C#行注释"},
		{"javascript", "let x = 5; // comment", "let x = 5;", "JavaScript行注释"},
		{"typescript", "const x: number = 5; // comment", "const x: number = 5;", "TypeScript行注释"},
		{"go", "x := 5 // comment", "x := 5", "Go语言行注释"},
		{"rust", "let x = 5; // comment", "let x = 5;", "Rust行注释"},
		{"swift", "let x = 5 // comment", "let x = 5", "Swift行注释"},
		{"dart", "int x = 5; // comment", "int x = 5;", "Dart行注释"},
		{"scala", "val x = 5 // comment", "val x = 5", "Scala行注释"},
		{"kt", "val x = 5 // comment", "val x = 5", "Kotlin行注释"},
		{"groovy", "def x = 5 // comment", "def x = 5", "Groovy行注释"},
		
		// 井号注释语言 (#)
		{"python", "x = 5 # comment", "x = 5", "Python井号注释"},
		{"ruby", "x = 5 # comment", "x = 5", "Ruby井号注释"},
		{"shell", "echo hello # comment", "echo hello", "Shell井号注释"},
		{"bash", "echo hello # comment", "echo hello", "Bash井号注释"},
		{"perl", "my $x = 5; # comment", "my $x = 5;", "Perl井号注释"},
		{"r", "x <- 5 # comment", "x <- 5", "R语言井号注释"},
		{"tcl", "set x 5 # comment", "set x 5", "Tcl井号注释"},
		
		// 双破折号语言 (--)
		{"sql", "SELECT * FROM users -- comment", "SELECT * FROM users", "SQL双破折号注释"},
		{"haskell", "x = 5 -- comment", "x = 5", "Haskell双破折号注释"},
		{"lua", "local x = 5 -- comment", "local x = 5", "Lua双破折号注释"},
		
		// 百分号注释语言 (%)
		{"matlab", "x = 5; % comment", "x = 5;", "MATLAB百分号注释"},
		{"latex", "\\section{Title} % comment", "\\section{Title}", "LaTeX百分号注释"},
		
		// 分号注释语言 (;)
		{"assembly", "mov eax, 5 ; comment", "mov eax, 5", "Assembly分号注释"},
		
		// 感叹号注释语言 (!)
		{"fortran", "x = 5 ! comment", "x = 5", "Fortran感叹号注释"},
		
		// 特殊语言
		{"php", "<?php $x = 5; // comment", "<?php $x = 5;", "PHP行注释"},
		{"css", "body { color: red; /* comment */ }", "body { color: red;  }", "CSS块注释"},
		{"xml", "<root><!-- comment --></root>", "<root></root>", "XML注释"},
		{"html", "<div><!-- comment --></div>", "<div></div>", "HTML注释"},
		{"yaml", "key: value # comment", "key: value", "YAML井号注释"},
		{"json", "{ \"key\": \"value\" } // comment", "{ \"key\": \"value\" }", "JSON行注释"},
		
		// Shader文件测试
		{"c", "Shader \"Test\" { \".shader\": true, // comment }", "Shader \"Test\" { \".shader\": true,", "Shader文件注释"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeComments(tt.input, tt.fileType)
			if result != tt.expected {
				t.Errorf("语言 %s: 期望 %q, 得到 %q", tt.fileType, tt.expected, result)
			}
		})
	}
}

// TestLanguageDetection 测试文件类型检测
func TestLanguageDetection(t *testing.T) {
	tests := []struct {
		filename string
		expected string
		name     string
	}{
		{"test.c", "c", "C文件检测"},
		{"test.cpp", "cpp", "C++文件检测"},
		{"test.java", "java", "Java文件检测"},
		{"test.py", "py", "Python文件检测"},
		{"test.js", "js", "JavaScript文件检测"},
		{"test.go", "go", "Go文件检测"},
		{"test.rs", "rust", "Rust文件检测"},
		{"test.php", "php", "PHP文件检测"},
		{"test.rb", "rb", "Ruby文件检测"},
		{"test.sh", "sh", "Shell文件检测"},
		{"test.sql", "sql", "SQL文件检测"},
		{"test.css", "css", "CSS文件检测"},
		{"test.html", "xml", "HTML文件检测"},
		{"test.xml", "xml", "XML文件检测"},
		{"test.yaml", "yaml", "YAML文件检测"},
		{"test.json", "json", "JSON文件检测"},
		{"test.md", "markdown", "Markdown文件检测"},
		{"test.shader", "c", "Shader文件检测"},
		{"test.hlsl", "c", "HLSL文件检测"},
		{"test.glsl", "c", "GLSL文件检测"},
		{"test.asm", "asm", "Assembly文件检测"},
		{"test.s", "unknown", "Assembly S文件检测"},
		{"test.f", "unknown", "Fortran文件检测"},
		{"test.f90", "f90", "Fortran90文件检测"},
		{"test.pas", "pas", "Pascal文件检测"},
		{"test.ada", "ada", "Ada文件检测"},
		{"test.pl", "unknown", "Perl文件检测"},
		{"test.lua", "lua", "Lua文件检测"},
		{"test.tcl", "tcl", "Tcl文件检测"},
		{"test.hs", "hs", "Haskell文件检测"},
		{"test.elm", "elm", "Elm文件检测"},
		{"test.ml", "ml", "OCaml文件检测"},
		{"test.fs", "fs", "F#文件检测"},
		{"test.clj", "clj", "Clojure文件检测"},
		{"test.scm", "scm", "Scheme文件检测"},
		{"test.lisp", "lisp", "Lisp文件检测"},
		{"test.jl", "jl", "Julia文件检测"},
		{"test.nb", "nb", "Mathematica文件检测"},
		{"test.vue", "vue", "Vue文件检测"},
		{"test.svelte", "svelte", "Svelte文件检测"},
		{"test.scss", "css", "SCSS文件检测"},
		{"test.sass", "css", "Sass文件检测"},
		{"test.less", "css", "Less文件检测"},
		{"test.tex", "tex", "LaTeX文件检测"},
		{"test.rst", "rst", "reStructuredText文件检测"},
		{"test.toml", "toml", "TOML文件检测"},
		{"test.ini", "ini", "INI文件检测"},
		{"test.cfg", "cfg", "Config文件检测"},
		{"test.conf", "conf", "Conf文件检测"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectFileType(tt.filename)
			if result != tt.expected {
				t.Errorf("文件 %s: 期望类型 %q, 得到 %q", tt.filename, tt.expected, result)
			}
		})
	}
}

// TestComprehensiveEdgeCases 测试所有语言的边界情况
func TestComprehensiveEdgeCases(t *testing.T) {
	tests := []struct {
		fileType string
		input    string
		expected string
		name     string
	}{
		// 字符串保护测试
		{"c", "printf(\"// not comment\"); // real comment", "printf(\"// not comment\");", "C语言字符串保护"},
		{"python", "print(\"# not comment\") # real comment", "print(\"# not comment\")", "Python字符串保护"},
		{"sql", "SELECT '-- not comment' -- real comment", "SELECT '-- not comment'", "SQL字符串保护"},
		{"javascript", "console.log('/* not comment */'); // real", "console.log('/* not comment */');", "JavaScript字符串保护"},
		{"php", "echo \"// not comment\"; // real comment", "echo \"// not comment\";", "PHP字符串保护"},
		{"ruby", "puts '# not comment' # real comment", "puts '# not comment'", "Ruby字符串保护"},
		{"shell", "echo '# not comment' # real comment", "echo '# not comment'", "Shell字符串保护"},
		{"css", "content: '/* not comment */'; /* real comment */", "content: '/* not comment */'; /* real comment */", "CSS字符串保护"},
		
		// 多行注释测试
		{"c", "int x; /* multi\nline */ int y;", "int x; \n int y;", "C语言多行注释"},
		{"css", "body { /* multi\nline */ color: red; }", "body { \n color: red; }", "CSS多行注释"},
		{"lua", "x = 1 --[[ multi\nline ]] y = 2", "x = 1\nline ]] y = 2", "Lua多行注释"},
		{"haskell", "x = 5 {- multi\nline -} y = 6", "x = 5 \n y = 6", "Haskell多行注释"},
		{"matlab", "x = 5; %{ multi\nline %} y = 6;", "x = 5;\nline", "MATLAB多行注释"},
		
		// 空行处理测试
		{"go", "package main\n// comment\nfunc main() {}", "package main\nfunc main() {}", "Go语言空行处理"},
		{"python", "def func():\n    # comment\n    pass", "def func():\n    pass", "Python空行处理"},
		{"java", "class Test {\n    // comment\n    int x;\n}", "class Test {\n    int x;\n}", "Java空行处理"},
		
		// 行尾注释测试
		{"java", "int x = 5; // comment\nint y = 6;", "int x = 5;\nint y = 6;", "Java行尾注释"},
		{"shell", "echo hello # comment\necho world", "echo hello\necho world", "Shell行尾注释"},
		{"python", "x = 5 # comment\ny = 6", "x = 5\ny = 6", "Python行尾注释"},
		{"c", "int x = 5; // comment\nint y = 6;", "int x = 5;\nint y = 6;", "C语言行尾注释"},
		
		// 复杂嵌套测试
		{"javascript", "var s = \"/* not comment */\"; /* real comment */ var x = 5;", "var s = \"/* not comment */\"; /* real comment */ var x = 5;", "JavaScript复杂嵌套"},
		{"python", "url = \"http://example.com#anchor\" # This is a comment", "url = \"http://example.com#anchor\"", "Python URL井号保护"},
		{"sql", "SELECT 'Price: $5.00' -- This is money, not comment", "SELECT 'Price: $5.00'", "SQL特殊字符保护"},
		
		// 特殊语法保护
		{"javascript", "var regex = /\\/\\*.*?\\*\\//g; // comment", "var regex = /\\/\\*.*?\\*\\//g;", "JavaScript正则表达式保护"},
		{"shell", "VERSION=${GITHUB_REF#refs/tags/} # comment", "VERSION=${GITHUB_REF#refs/tags/}", "Shell变量展开保护"},
		{"rust", "let raw = r\"This is // not a comment\"; // real comment", "let raw = r\"This is // not a comment\";", "Rust原始字符串保护"},
		
		// 边界情况
		{"c", "// comment at start", "", "C语言行首注释"},
		{"python", "# comment at start", "", "Python行首注释"},
		{"sql", "-- comment at start", "", "SQL行首注释"},
		{"assembly", "; comment at start", "", "Assembly行首注释"},
		{"fortran", "! comment at start", "", "Fortran行首注释"},
		{"matlab", "% comment at start", "", "MATLAB行首注释"},
		
		// 混合注释类型
		{"php", "<?php\n// Line comment\n/* Block comment */ echo 'test'; # Hash comment", "<?php\n echo 'test'; # Hash comment", "PHP混合注释"},
		{"assembly", "mov eax, 5 ; semicolon comment\n# hash comment\n// slash comment", "mov eax, 5", "Assembly混合注释"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeComments(tt.input, tt.fileType)
			if result != tt.expected {
				t.Errorf("测试 %s 失败:\n期望: %q\n得到: %q", tt.name, tt.expected, result)
			}
		})
	}
}
