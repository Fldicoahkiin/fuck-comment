package main

import "strings"

// StringType 字符串类型枚举
type StringType int

const (
	StringTypeAll StringType = iota // 所有类型字符串
	StringTypeQuote                 // 仅单双引号字符串
	StringTypeBacktick              // 仅反引号字符串
)

// isInStringWithType 统一的字符串检测函数
func isInStringWithType(line string, pos int, stringType StringType) bool {
	if pos >= len(line) {
		return false
	}
	
	var inSingleQuote, inDoubleQuote, inBacktick bool
	
	// 检查到pos位置之前的所有字符（不包括pos位置本身）
	for i := 0; i < pos; i++ {
		char := line[i]
		
		switch char {
		case '\'':
			if !inDoubleQuote && !inBacktick && (stringType == StringTypeAll || stringType == StringTypeQuote) {
				if !isEscaped(line, i) {
					inSingleQuote = !inSingleQuote
				}
			}
		case '"':
			if !inSingleQuote && !inBacktick && (stringType == StringTypeAll || stringType == StringTypeQuote) {
				if !isEscaped(line, i) {
					inDoubleQuote = !inDoubleQuote
				}
			}
		case '`':
			if !inSingleQuote && !inDoubleQuote && (stringType == StringTypeAll || stringType == StringTypeBacktick) {
				inBacktick = !inBacktick
			}
		}
	}
	
	switch stringType {
	case StringTypeQuote:
		return inSingleQuote || inDoubleQuote
	case StringTypeBacktick:
		return inBacktick
	default: // StringTypeAll
		return inSingleQuote || inDoubleQuote || inBacktick
	}
}

// isEscaped 检查字符是否被转义
func isEscaped(line string, pos int) bool {
	if pos == 0 {
		return false
	}
	
	backslashCount := 0
	for i := pos - 1; i >= 0 && line[i] == '\\'; i-- {
		backslashCount++
	}
	// 奇数个反斜杠表示当前字符被转义
	return backslashCount%2 == 1
}

// isStructuralComment 检查是否为结构性注释（通用模式）
func isStructuralComment(comment string) bool {
	// 去掉注释符号，获取纯内容
	content := strings.TrimSpace(strings.TrimPrefix(comment, "#"))
	
	// 空注释或只有符号的注释不是结构性的
	if len(content) == 0 {
		return false
	}
	
	// 排除明显的普通注释模式
	commonPhrases := []string{"这是", "这个", "用于", "表示", "注释", "说明"}
	for _, phrase := range commonPhrases {
		if strings.Contains(content, phrase) {
			return false
		}
	}
	
	// 1. 包含emoji的注释通常是结构性的
	if containsEmoji(content) {
		return true
	}
	
	// 2. 包含分隔符的注释通常是结构性的
	separators := []string{"===", "---", "***", "###", "+++", "~~~"}
	for _, sep := range separators {
		if strings.Contains(content, sep) {
			return true
		}
	}
	
	// 3. 以数字开头的注释通常是步骤或列表项
	if len(content) > 0 && (content[0] >= '0' && content[0] <= '9') {
		return true
	}
	
	// 4. 短且包含特殊字符的通常是结构性的
	if len(content) <= 15 {
		specialChars := []string{"→", "•", "★", "▶", "◆", "■", "▲", "►"}
		for _, char := range specialChars {
			if strings.Contains(content, char) {
				return true
			}
		}
	}
	
	// 5. 全大写且较短的注释通常是标题
	if strings.ToUpper(content) == content && len(content) > 2 && len(content) <= 20 {
		// 排除常见的普通注释词汇
		commonWords := []string{"TODO", "FIXME", "HACK", "NOTE", "WARNING"}
		for _, word := range commonWords {
			if strings.Contains(content, word) {
				return false
			}
		}
		return true
	}
	
	return false
}

// containsEmoji 检查字符串是否包含emoji
func containsEmoji(s string) bool {
	for _, r := range s {
		// 检查常见的emoji范围
		if (r >= 0x1F600 && r <= 0x1F64F) || // 表情符号
		   (r >= 0x1F300 && r <= 0x1F5FF) || // 杂项符号
		   (r >= 0x1F680 && r <= 0x1F6FF) || // 交通和地图符号
		   (r >= 0x2600 && r <= 0x26FF) ||   // 杂项符号
		   (r >= 0x2700 && r <= 0x27BF) ||   // 装饰符号
		   (r >= 0x1F900 && r <= 0x1F9FF) {  // 补充符号
			return true
		}
	}
	return false
}

// 保留原有函数名作为兼容性包装
func isInQuoteString(line string, pos int) bool {
	return isInStringWithType(line, pos, StringTypeQuote)
}

func isInAnyString(line string, pos int) bool {
	return isInStringWithType(line, pos, StringTypeAll)
}

func isInString(line string, pos int) bool {
	return isInStringWithType(line, pos, StringTypeAll)
}

func isInBacktickString(line string, pos int) bool {
	return isInStringWithType(line, pos, StringTypeBacktick)
}

// isInRegex 检查指定位置是否在正则表达式内
func isInRegex(line string, pos int) bool {
	if pos >= len(line) {
		return false
	}
	
	var inSingleQuote, inDoubleQuote, inBacktick bool
	var inRegex bool
	
	for i := 0; i < pos; i++ {
		char := line[i]
		
		switch char {
		case '\'':
			if !inDoubleQuote && !inBacktick && !inRegex {
				if !isEscaped(line, i) {
					inSingleQuote = !inSingleQuote
				}
			}
		case '"':
			if !inSingleQuote && !inBacktick && !inRegex {
				if !isEscaped(line, i) {
					inDoubleQuote = !inDoubleQuote
				}
			}
		case '`':
			if !inSingleQuote && !inDoubleQuote && !inRegex {
				inBacktick = !inBacktick
			}
		case '/':
			if !inSingleQuote && !inDoubleQuote && !inBacktick {
				if i > 0 && line[i-1] == '=' {
					if !isEscaped(line, i) {
						inRegex = !inRegex
					}
				}
			}
		}
	}
	
	return inRegex
}
