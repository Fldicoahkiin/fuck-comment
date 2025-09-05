package main

import "fmt"

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
		return
	}
	
	fmt.Printf("\n")
	fmt.Printf(ColorGreen+"%d"+ColorReset+" 处理", len(processedFiles))
	if len(skippedFiles) > 0 {
		fmt.Printf(" | "+ColorYellow+"%d"+ColorReset+" 跳过", len(skippedFiles))
	}
	if backupRootDir != "" {
		fmt.Printf(" | 备份: "+ColorCyan+"%s"+ColorReset, backupRootDir)
	}
}
