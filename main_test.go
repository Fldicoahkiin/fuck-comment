package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	err = processFile(testFile)
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

	expected := `# 标题

正文内容

## 二级标题



### 三级标题`

	result := removeMarkdownComments(input)
	if result != expected {
		t.Errorf("removeMarkdownComments() = %q, want %q", result, expected)
	}
}

// 测试YAML注释处理
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

	expected := `version: '3.8'
services:
  web:
    image: nginx
    ports:
      - "80:80"

  database:
    image: postgres`

	result := removeYamlComments(input)
	if result != expected {
		t.Errorf("removeYamlComments() = %q, want %q", result, expected)
	}
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
			result := removeYamlComments(tt.input)
			if result != tt.expected {
				t.Errorf("removeYamlComments() = %q, want %q", result, tt.expected)
			}
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
			pos:      30,
			expected: false,
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
