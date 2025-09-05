package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
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
)

// processFile 处理单个文件，删除其中的注释
func processFile(filePath, workingDir string) error {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}
	
	// 安全检查
	if err := isFileSafe(filePath, content, forceMode); err != nil {
		skippedFiles = append(skippedFiles, filePath)
		printWarning("%s", err.Error())
		return nil
	}
	
	// 检测文件类型
	fileType := detectFileType(filePath)
	if fileType == "unknown" {
		skippedFiles = append(skippedFiles, filePath)
		printWarning("无法识别文件类型: %s", filePath)
		return nil
	}
	
	// 删除注释
	originalContent := string(content)
	processedContent := removeComments(originalContent, fileType)
	
	// 检查是否有变化
	if originalContent == processedContent {
		// 无变化，不需要备份和写入
		relPath, _ := filepath.Rel(workingDir, filePath)
		fmt.Printf("%s |%s| 无变化\n", relPath, strings.ToUpper(fileType))
		return nil
	}
	
	// 创建备份
	if err := createBackup(filePath, workingDir); err != nil {
		return fmt.Errorf("创建备份失败: %v", err)
	}
	
	// 写入处理后的内容
	err = os.WriteFile(filePath, []byte(processedContent), 0644)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}
	
	// 记录处理的文件
	processedFiles = append(processedFiles, filePath)
	
	// 显示处理结果
	relPath, _ := filepath.Rel(workingDir, filePath)
	fmt.Printf("%s |%s| ✓\n", relPath, strings.ToUpper(fileType))
	
	return nil
}

// processDirectory 递归处理目录中的所有支持文件
func processDirectory(rootDir string) error {
	return filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// 跳过目录
		if d.IsDir() {
			// 跳过隐藏目录和备份目录
			if strings.HasPrefix(d.Name(), ".") || d.Name() == "bak" {
				return fs.SkipDir
			}
			return nil
		}
		
		// 跳过隐藏文件
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		
		// 检查是否为支持的文件类型
		if !isSupportedFile(path, forceMode) {
			return nil
		}
		
		// 处理文件
		if err := processFile(path, rootDir); err != nil {
			printError("处理文件 %s 失败: %v", path, err)
		}
		
		return nil
	})
}

var rootCmd = &cobra.Command{
	Use:   "fuck-comment [directory]",
	Short: "删除代码注释的命令行工具",
	Long: "删除代码文件中的注释，支持137种文件扩展名。\n\n" +
		"支持的注释格式：\n" +
		"  `//`         行注释 (C/C++, Go, Java, JavaScript等)\n" +
		"  `/* */`      块注释 (C/C++, Go, Java, JavaScript等)\n" +
		"  `#`          井号注释 (Python, Shell, YAML等)\n" +
		"  `--`         双破折号注释 (SQL, Haskell等)\n" +
		"  `;`          分号注释 (Assembly, Lisp等)\n" +
		"  `%`          百分号注释 (LaTeX, MATLAB等)\n" +
		"  `!`          感叹号注释 (Fortran等)\n" +
		"  `<!-- -->`   HTML注释 (HTML, XML等)\n\n" +
		"安全特性：\n" +
		"  • 自动备份到 bak/ 目录\n" +
		"  • 跳过二进制文件\n" +
		"  • 保护字符串中的注释符号\n" +
		"  • 保护URL锚点和Shell变量\n\n" +
		"参数说明：\n" +
		"  -f, --file string    指定要处理的单个文件\n" +
		"      --force          强制处理所有文件类型（包括二进制文件）\n" +
		"      --version        显示版本信息\n\n" +
		"使用示例:\n" +
		"  fuck-comment              删除当前目录所有支持文件的注释\n" +
		"  fuck-comment /path/to/dir 删除指定目录及其子目录的注释\n" +
		"  fuck-comment -f main.go   删除指定文件的注释\n" +
		"  fuck-comment --force      强制处理所有文件类型\n\n" +
		"注意事项：\n" +
		"  • 处理前会自动创建备份，备份文件保存在 bak/ 目录\n" +
		"  • 默认跳过二进制文件和隐藏文件\n" +
		"  • 使用 --force 参数可强制处理所有文件类型",
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
			
			// 显示处理结果摘要
			printSummary()
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
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}
