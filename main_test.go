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
		// 创建临时文件
		tmpFile, err := ioutil.TempFile("", "test_backup_*.txt")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpFile.Name())
		defer func() {
			// 清理可能的备份文件
			os.RemoveAll("bak")
		}()
		
		testContent := "// This is a test\nfunc main() {}"
		tmpFile.WriteString(testContent)
		tmpFile.Close()
		
		// 创建备份
		err = createBackup(tmpFile.Name())
		if err != nil {
			t.Errorf("创建备份失败: %v", err)
		}
		
		// 验证备份文件存在且内容正确
		// 新的备份机制会在bak/时间戳目录下创建备份
		workDir, _ := os.Getwd()
		bakDir := filepath.Join(workDir, "bak")
		
		// 查找备份文件
		var backupPath string
		filepath.Walk(bakDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.Contains(path, filepath.Base(tmpFile.Name())) {
				backupPath = path
			}
			return nil
		})
		
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
		
		// 测试重复备份
		err = createBackup(tmpFile.Name())
		if err != nil {
			t.Errorf("重复备份失败: %v", err)
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
