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
			result := removeComments(tt.input)
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
			name:     "不支持的文件",
			filePath: "README.md",
			force:    false,
			expected: false,
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
		removeComments(content)
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
