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
		".go":    true, // Go
		".c":     true, // C
		".cpp":   true, // C++
		".cc":    true, // C++
		".cxx":   true, // C++
		".h":     true, // C/C++ Header
		".hpp":   true, // C++ Header
		".java":  true, // Java
		".js":    true, // JavaScript
		".jsx":   true, // React JSX
		".ts":    true, // TypeScript
		".tsx":   true, // TypeScript JSX
		".cs":    true, // C#
		".php":   true, // PHP
		".swift":  true, // Swift
		".kt":    true, // Kotlin
		".rs":    true, // Rust
		".scala": true, // Scala
		".dart":  true, // Dart
		".m":     true, // Objective-C
		".mm":    true, // Objective-C++
	}

	// CLI 参数
	targetFile string
	forceMode  bool
	verbose    bool
	showVersion bool
)

// removeComments 删除代码中的注释，支持 // 和 /* */ 格式
// 智能处理字符串字面量，不会删除字符串内的注释符号
func removeComments(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inBlockComment := false

	for _, line := range lines {
		processedLine := line
		
		if inBlockComment {
			// 在块注释中，查找结束标记
			if endIndex := strings.Index(line, "*/"); endIndex != -1 {
				processedLine = line[endIndex+2:]
				inBlockComment = false
			} else {
				// 整行都在块注释中
				processedLine = ""
			}
		}
		
		if !inBlockComment {
			// 处理行注释 //
			if idx := strings.Index(processedLine, "//"); idx != -1 {
				// 检查是否在字符串中
				if !isInString(processedLine, idx) {
					processedLine = processedLine[:idx]
				}
			}
			
			// 处理块注释 /* */
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
		result = append(result, processedLine)
	}
	
	return strings.Join(result, "\n")
}

// isInString 检查指定位置是否在字符串字面量中
// 支持单引号和双引号字符串，正确处理转义字符
func isInString(line string, pos int) bool {
	inSingleQuote := false
	inDoubleQuote := false
	escaped := false
	
	for i, char := range line {
		if i >= pos {
			break
		}
		
		if escaped {
			escaped = false
			continue
		}
		
		switch char {
		case '\\':
			escaped = true
		case '\'':
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			}
		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			}
		}
	}
	
	return inSingleQuote || inDoubleQuote
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
	
	// 删除注释
	newContent := removeComments(string(content))
	
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
	Long: `fuck-comment 是一个跨平台的CLI工具，用于删除代码文件中的 // 和 /* */ 注释。

支持的编程语言包括：
Go, C/C++, Java, JavaScript, TypeScript, C#, PHP, Swift, Kotlin, Rust, Scala, Dart, Objective-C 等

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
