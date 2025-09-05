package main

import "strings"

// CommentRule 定义注释处理规则
type CommentRule struct {
	StartPattern string
	EndPattern   string
	IsLineComment bool
	ProtectFunc  func(line string, pos int) bool
}

// ProtectionContext 保护上下文结构体
type ProtectionContext struct {
	Line        string
	Pos         int
	FileType    string
	CommentStart string
}

// shouldProtectInContext 检查是否应该在特定上下文中保护注释符号
func shouldProtectInContext(line string, pos int, fileType string, commentStart string) bool {
	ctx := ProtectionContext{
		Line:         line,
		Pos:          pos,
		FileType:     fileType,
		CommentStart: commentStart,
	}
	return checkProtectionRules(ctx)
}

// checkProtectionRules 检查保护规则
func checkProtectionRules(ctx ProtectionContext) bool {
	switch ctx.FileType {
	case "c", "cpp", "cc", "cxx", "h", "hpp", "java", "javascript", "js", "typescript", "ts", "go", "rust", "rs", "swift", "kotlin", "scala", "dart", "cs":
		// 首先检查是否在普通字符串内（单引号或双引号）
		if isInStringWithType(ctx.Line, ctx.Pos, StringTypeQuote) {
			return true
		}
		
		// 保护反引号代码块中的注释符号
		if isInStringWithType(ctx.Line, ctx.Pos, StringTypeBacktick) {
			return true
		}
		
		// JavaScript正则表达式保护
		if (ctx.FileType == "javascript" || ctx.FileType == "js") && (ctx.CommentStart == "//" || ctx.CommentStart == "/*") {
			beforeComment := ctx.Line[:ctx.Pos]
			// 检查是否在正则表达式字面量内
			if strings.Contains(beforeComment, "= /") || strings.Contains(beforeComment, "(/") || 
			   strings.Contains(beforeComment, " /") || strings.Contains(beforeComment, "\t/") {
				// 更精确的正则表达式检测
				regexStart := -1
				for i := len(beforeComment) - 1; i >= 0; i-- {
					if beforeComment[i] == '/' && (i == 0 || 
						beforeComment[i-1] == '=' || beforeComment[i-1] == '(' || 
						beforeComment[i-1] == ' ' || beforeComment[i-1] == '\t') {
						regexStart = i
						break
					}
				}
				if regexStart != -1 {
					// 检查从正则开始到当前位置是否没有结束的斜杠
					afterRegexStart := beforeComment[regexStart+1:]
					unescapedSlashCount := 0
					for i, char := range afterRegexStart {
						if char == '/' && !isEscaped(afterRegexStart, i) {
							unescapedSlashCount++
						}
					}
					if unescapedSlashCount == 0 {
						return true // 在正则表达式内
					}
				}
			}
		}
		
		// 特殊处理：检查是否在字符串拼接中的反引号代码块内
		if ctx.FileType == "go" && (ctx.CommentStart == "//" || ctx.CommentStart == "/*") {
			beforeComment := ctx.Line[:ctx.Pos]
			// 检查当前行是否包含反引号且在字符串拼接中
			if strings.Contains(beforeComment, "\"") && strings.Contains(beforeComment, "`") {
				// 检查反引号是否在字符串内部
				lastQuote := strings.LastIndex(beforeComment, "\"")
				if lastQuote >= 0 {
					afterQuote := beforeComment[lastQuote+1:]
					// 如果在最后一个引号之后有反引号，说明可能在代码块内
					if strings.Contains(afterQuote, "`") {
						// 计算反引号数量，奇数表示在代码块内
						backtickCount := strings.Count(afterQuote, "`")
						if backtickCount%2 == 1 {
							return true
						}
					}
				}
			}
		}
		
		break
	case "php":
		// 首先检查是否在普通字符串内
		if isInStringWithType(ctx.Line, ctx.Pos, StringTypeQuote) {
			return true
		}
		
		// PHP特殊处理：保留行尾的井号注释
		if ctx.CommentStart == "#" {
			beforeComment := ctx.Line[:ctx.Pos]
			// 如果井号前有代码内容，保护这个井号注释
			if strings.TrimSpace(beforeComment) != "" {
				return true
			}
		}
		
		break
	case "yaml", "yml":
		if ctx.CommentStart == "#" {
			beforeComment := ctx.Line[:ctx.Pos]
			// 保护字符串内的#
			if isInAnyString(ctx.Line, ctx.Pos) {
				return true
			}
			
			// 保护Shell变量展开中的#（如${VAR#pattern}）
			if strings.Contains(beforeComment, "${") {
				// 检查整行的Shell变量语法
				fullLine := ctx.Line
				openBraces := strings.Count(fullLine[:ctx.Pos], "{")
				closeBraces := strings.Count(fullLine[:ctx.Pos], "}")
				if openBraces > closeBraces {
					// 检查#后面是否有}来确认这是Shell变量语法
					afterHash := fullLine[ctx.Pos+1:]
					if strings.Contains(afterHash, "}") {
						return true
					}
				}
			}
			
			// 保护URL中的锚点
			if strings.Contains(beforeComment, "http") && strings.Contains(ctx.Line[ctx.Pos:], "#") {
				return true
			}
			
			// 保护行首注释（仅保护结构性注释）
			if strings.TrimSpace(beforeComment) == "" {
				// 检查是否为结构性注释
				comment := strings.TrimSpace(ctx.Line[ctx.Pos:])
				
				// 保护markdown风格标题 (# ## ### 等)
				if strings.HasPrefix(comment, "# #") || strings.HasPrefix(comment, "# ##") || strings.HasPrefix(comment, "# ###") ||
				   strings.HasPrefix(comment, "## ") || strings.HasPrefix(comment, "### ") {
					return true
				}
				
				// 保护结构性注释的通用模式
				if isStructuralComment(comment) {
					return true
				}
				
				// 其他行首注释不保护（普通注释）
				return false
			}
			
			// 对于行尾注释，只保护字符串内和特殊URL情况，不保护普通注释
			return false
		}
	case "css", "scss", "sass", "less":
		// CSS中保护URL和content属性中的注释符号
		if ctx.CommentStart == "/*" || ctx.CommentStart == "//" {
			// 检查是否在url()函数中
			if strings.Contains(ctx.Line[:ctx.Pos], "url(") && !strings.Contains(ctx.Line[:ctx.Pos], ")") {
				return true
			}
			// 检查是否在content属性中
			if strings.Contains(ctx.Line[:ctx.Pos], "content:") {
				return true
			}
		}
	case "html", "xml", "svg":
		// HTML/XML中保护属性值和CDATA中的注释符号
		if ctx.CommentStart == "<!--" {
			// 检查是否在CDATA中
			if strings.Contains(ctx.Line[:ctx.Pos], "<![CDATA[") && !strings.Contains(ctx.Line[:ctx.Pos], "]]>") {
				return true
			}
		}
		// 保护条件语句和不完整的语句
		if ctx.CommentStart == "//" || ctx.CommentStart == "/*" {
			beforeComment := strings.TrimSpace(ctx.Line[:ctx.Pos])
			// 保护不完整的条件语句
			if strings.Contains(beforeComment, "if ") && !strings.Contains(beforeComment, "{") {
				return true
			}
			if strings.Contains(beforeComment, "for ") && !strings.Contains(beforeComment, "{") {
				return true
			}
			if strings.Contains(beforeComment, "while ") && !strings.Contains(beforeComment, "{") {
				return true
			}
			// 保护包含 != 的语句（但不包括Rust的情况）
			if strings.Contains(beforeComment, "!=") && !strings.Contains(beforeComment, "{") && ctx.FileType != "rust" {
				return true
			}
			// Rust特殊保护
			if ctx.FileType == "rust" || ctx.FileType == "rs" {
				// 保护println!宏调用
				if strings.Contains(beforeComment, "println!") && !strings.Contains(beforeComment, ";") {
					return true
				}
				if strings.Contains(beforeComment, "panic!") && !strings.Contains(beforeComment, ";") {
					return true
				}
			}
		}
	case "python", "py":
		return checkPythonProtection(ctx)
	case "shell", "bash", "zsh", "sh":
		return checkShellProtection(ctx)
	}
	// Rust特殊保护
	if ctx.FileType == "rust" || ctx.FileType == "rs" {
		beforeComment := ctx.Line[:ctx.Pos]
		// 保护println!宏调用
		if strings.Contains(beforeComment, "println!") && !strings.Contains(beforeComment, ";") {
			return true
		}
		if strings.Contains(beforeComment, "panic!") && !strings.Contains(beforeComment, ";") {
			return true
		}
	}
	
	// 通用字符串保护
	return isInAnyString(ctx.Line, ctx.Pos)
}

// checkPythonProtection 检查Python的保护规则
func checkPythonProtection(ctx ProtectionContext) bool {
	if ctx.CommentStart == "#" {
		// 首先检查是否在普通字符串内
		if isInStringWithType(ctx.Line, ctx.Pos, StringTypeQuote) {
			return true
		}
		
		beforeComment := ctx.Line[:ctx.Pos]
		
		// 保护docstring中的#（仅在docstring内部）
		if strings.Contains(beforeComment, `"""`) {
			firstTriple := strings.Index(beforeComment, `"""`)
			afterFirst := beforeComment[firstTriple+3:]
			if !strings.Contains(afterFirst, `"""`) {
				return true // 在未闭合的docstring内部
			}
		}
		if strings.Contains(beforeComment, "'''") {
			firstTriple := strings.Index(beforeComment, "'''")
			afterFirst := beforeComment[firstTriple+3:]
			if !strings.Contains(afterFirst, "'''") {
				return true // 在未闭合的docstring内部
			}
		}
		
		// 保护字符串中的URL锚点（只有当#确实在字符串内部时才保护）
		if isInAnyString(ctx.Line, ctx.Pos) && strings.Contains(beforeComment, "http") {
			return true
		}
		
		// 保护Python原始字符串中的#
		if strings.Contains(beforeComment, "r\"") || strings.Contains(beforeComment, "r'") {
			quoteCount := strings.Count(beforeComment, "\"") + strings.Count(beforeComment, "'")
			if quoteCount%2 == 1 {
				return true
			}
		}
		
		// f-string处理：只保护{}内部的#，不保护字符串外的注释
		if strings.Contains(beforeComment, "f\"") || strings.Contains(beforeComment, "f'") {
			// 检查#是否在f-string的{}内部
			braceCount := 0
			inFString := false
			var stringChar byte
			
			for i, char := range []byte(beforeComment) {
				if char == 'f' && i < len(beforeComment)-1 && (beforeComment[i+1] == '"' || beforeComment[i+1] == '\'') {
					inFString = true
					stringChar = beforeComment[i+1]
				} else if inFString && char == stringChar && !isEscaped(beforeComment, i) {
					inFString = false
				} else if inFString && char == '{' {
					braceCount++
				} else if inFString && char == '}' {
					braceCount--
				}
			}
			
			// 只有在f-string的{}内部才保护#
			return inFString && braceCount > 0
		}
	}
	return false
}

// checkShellProtection 检查Shell脚本的保护规则
func checkShellProtection(ctx ProtectionContext) bool {
	if ctx.CommentStart == "#" {
		// 首先检查是否在普通字符串内
		if isInStringWithType(ctx.Line, ctx.Pos, StringTypeQuote) {
			return true
		}
		
		// 保护shebang
		if ctx.Pos == 0 && strings.HasPrefix(ctx.Line, "#!") {
			return true
		}
		// 保护变量替换中的#，如 ${GITHUB_REF#refs/tags/}
		beforeComment := ctx.Line[:ctx.Pos]
		if strings.Contains(beforeComment, "${") {
			// 检查是否在变量替换的#操作符位置
			if strings.Count(beforeComment, "{") > strings.Count(beforeComment, "}") {
				return true
			}
		}
		// 保护条件语句中的#
		if strings.Contains(beforeComment, "[ ") && !strings.Contains(beforeComment, " ]") {
			return true
		}
		// 保护字符串中的URL锚点（只有当#确实在字符串内部时才保护）
		if isInAnyString(ctx.Line, ctx.Pos) && strings.Contains(beforeComment, "http") {
			return true
		}
		// 保护颜色代码（更精确的检查）
		if strings.Contains(beforeComment, "#") && len(beforeComment) >= 6 {
			// 检查是否是颜色代码格式
			lastHash := strings.LastIndex(beforeComment, "#")
			if lastHash >= 0 && lastHash < len(beforeComment)-1 {
				afterHash := beforeComment[lastHash+1:]
				if len(afterHash) == 6 || len(afterHash) == 3 {
					// 检查是否全为十六进制字符
					isHex := true
					for _, c := range afterHash {
						if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
							isHex = false
							break
						}
					}
					if isHex {
						return true
					}
				}
			}
		}
	}
	return false
}

// getCommentRulesForLanguage 获取指定语言的注释规则
func getCommentRulesForLanguage(fileType string) []CommentRule {
	// C风格语言 (// 和 /* */)
	cStyleRules := []CommentRule{
		{StartPattern: "//", EndPattern: "", IsLineComment: true},
		{StartPattern: "/*", EndPattern: "*/", IsLineComment: false},
	}
	
	// 井号注释语言 (#)
	hashStyleRules := []CommentRule{
		{StartPattern: "#", EndPattern: "", IsLineComment: true},
	}
	
	// 双破折号语言 (--)
	dashStyleRules := []CommentRule{
		{StartPattern: "--", EndPattern: "", IsLineComment: true},
	}
	_ = dashStyleRules // 避免未使用变量错误
	
	switch fileType {
	case "javascript", "js", "typescript", "ts", "go":
		return cStyleRules
	case "c", "cpp", "cc", "cxx", "h", "hpp":
		return cStyleRules
	case "java", "scala", "kotlin", "kt", "groovy":
		return cStyleRules
	case "rust", "rs", "swift", "dart", "cs":
		return cStyleRules
	case "php":
		// PHP支持多种注释风格
		return []CommentRule{
			{StartPattern: "//", EndPattern: "", IsLineComment: true},
			{StartPattern: "/*", EndPattern: "*/", IsLineComment: false},
			{StartPattern: "#", EndPattern: "", IsLineComment: true},
		}
	case "python", "py", "ruby", "rb", "shell", "bash", "zsh", "sh", "fish":
		return hashStyleRules
	case "perl", "pl", "pm", "tcl":
		return hashStyleRules
	case "r", "R":
		return hashStyleRules
	case "yaml", "yml", "toml", "ini", "cfg", "conf":
		return hashStyleRules
	case "sql", "plsql", "psql":
		return []CommentRule{
			{StartPattern: "--", EndPattern: "", IsLineComment: true},
			{StartPattern: "/*", EndPattern: "*/", IsLineComment: false},
		}
	case "lua":
		return []CommentRule{
			{StartPattern: "--", EndPattern: "", IsLineComment: true},
			{StartPattern: "--[[", EndPattern: "]]", IsLineComment: false},
		}
	case "haskell", "hs", "elm":
		return []CommentRule{
			{StartPattern: "--", EndPattern: "", IsLineComment: true},
			{StartPattern: "{-", EndPattern: "-}", IsLineComment: false},
		}
	case "ml", "ocaml":
		return []CommentRule{
			{StartPattern: "(*", EndPattern: "*)", IsLineComment: false},
		}
	case "css", "scss", "sass", "less":
		return []CommentRule{
			{StartPattern: "/*", EndPattern: "*/", IsLineComment: false},
		}
	case "html", "htm", "xml", "svg":
		return []CommentRule{
			{StartPattern: "<!--", EndPattern: "-->", IsLineComment: false},
		}
	case "latex", "tex":
		return []CommentRule{
			{StartPattern: "%", EndPattern: "", IsLineComment: true},
		}
	case "matlab", "m":
		return []CommentRule{
			{StartPattern: "%", EndPattern: "", IsLineComment: true},
			{StartPattern: "%{", EndPattern: "%}", IsLineComment: false},
		}
	case "assembly", "asm", "s", "S":
		return []CommentRule{
			{StartPattern: ";", EndPattern: "", IsLineComment: true},
			{StartPattern: "#", EndPattern: "", IsLineComment: true},
			{StartPattern: "//", EndPattern: "", IsLineComment: true},
		}
	case "fortran", "f", "f90", "f95", "for":
		return []CommentRule{
			{StartPattern: "!", EndPattern: "", IsLineComment: true},
			{StartPattern: "C", EndPattern: "", IsLineComment: true},
			{StartPattern: "c", EndPattern: "", IsLineComment: true},
		}
	case "lisp", "lsp", "scm", "clj", "cljs":
		return []CommentRule{
			{StartPattern: ";", EndPattern: "", IsLineComment: true},
		}
	case "verilog", "v", "vh", "sv":
		return []CommentRule{
			{StartPattern: "//", EndPattern: "", IsLineComment: true},
			{StartPattern: "/*", EndPattern: "*/", IsLineComment: false},
		}
	case "markdown", "md", "mdx":
		return []CommentRule{
			{StartPattern: "<!--", EndPattern: "-->", IsLineComment: false},
		}
	case "json", "jsonc", "json5":
		return cStyleRules
	default:
		return hashStyleRules // 默认使用井号注释
	}
}

// removeCommentsByRules 根据注释规则删除注释
func removeCommentsByRules(content string, fileType string, rules []CommentRule) string {
	lines := strings.Split(content, "\n")
	var result []string
	var inBlockComment bool
	var inMultiLineString bool
	var inBacktickString bool
	inYAMLMultiLineBlock := false
	yamlBlockIndent := 0
	var blockEndPattern string

	for _, line := range lines {
		originalLine := line
		processedLine := line
		
		// 如果是空行，直接保留
		if strings.TrimSpace(line) == "" {
			result = append(result, line)
			continue
		}
		
		// YAML多行字符串块检测
		if fileType == "yaml" || fileType == "yml" {
			trimmedLine := strings.TrimSpace(line)
			currentIndent := len(line) - len(strings.TrimLeft(line, " \t"))
			
			// 检测多行字符串块开始 (|, >, |-, >-)
			if strings.Contains(line, ": |") || strings.Contains(line, ": >") || 
			   strings.Contains(line, ": |-") || strings.Contains(line, ": >-") {
				inYAMLMultiLineBlock = true
				yamlBlockIndent = currentIndent
			} else if inYAMLMultiLineBlock {
				// 检查是否退出多行字符串块
				if trimmedLine != "" && currentIndent <= yamlBlockIndent {
					inYAMLMultiLineBlock = false
				}
			}
			
			// 如果在YAML多行字符串块中，保护所有内容
			if inYAMLMultiLineBlock {
				result = append(result, originalLine)
				continue
			}
		}
		
		// 检查多行字符串状态 - 在处理注释之前更新状态
		oldMultiLineState := inMultiLineString
		oldBacktickState := inBacktickString
		
		// 跟踪反引号字符串状态（用于Go/JS/TS模板字符串）
		if fileType == "go" || fileType == "js" || fileType == "ts" || fileType == "jsx" || fileType == "tsx" || fileType == "javascript" {
			backtickCount := 0
			for i := 0; i < len(line); i++ {
				if line[i] == '`' && !isEscaped(line, i) {
					backtickCount++
				}
			}
			if backtickCount%2 == 1 {
				inBacktickString = !inBacktickString
			}
		}
		
		// Python docstring 处理
		if fileType == "python" || fileType == "py" {
			tempInMultiLine := inMultiLineString
			singleLineDocstring := false
			
			// 检查是否有三引号
			if strings.Contains(line, `"""`) || strings.Contains(line, "'''") {
				// 检查单行docstring
				if strings.Count(line, `"""`) >= 2 || strings.Count(line, "'''") >= 2 {
					// 可能是单行docstring
					startPos := -1
					endPos := -1
					quote := ""
					
					if pos := strings.Index(line, `"""`); pos != -1 {
						startPos = pos
						quote = `"""`
					} else if pos := strings.Index(line, "'''"); pos != -1 {
						startPos = pos
						quote = "'''"
					}
					
					if startPos != -1 {
						// 查找结束位置
						endPos = strings.Index(line[startPos+3:], quote)
						if endPos != -1 {
							endPos += startPos + 3 + 3 // 加上开始位置和三引号长度
						}
						
						if endPos < len(line) {
							beforeEnd := line[:endPos]
							afterEnd := line[endPos:]
							// 删除docstring后的注释
							if pos := strings.Index(afterEnd, "#"); pos != -1 {
								afterEnd = strings.TrimRight(afterEnd[:pos], " \t")
							}
							processedLine = beforeEnd + afterEnd
							singleLineDocstring = true
						} else {
							// 单行docstring占据整行，不影响多行状态
							singleLineDocstring = true
						}
					}
				}
				
				if !singleLineDocstring {
					// 计算不在字符串内的三引号数量
					count := 0
					for i := 0; i <= len(line)-3; i++ {
						if line[i:i+3] == "'''" && !isInQuoteString(line, i) {
							count++
							if count%2 == 1 {
								tempInMultiLine = !tempInMultiLine
							}
							i += 2 // 跳过这个三引号
						}
					}
				}
			}
			
			// 如果这一行开始时在多行字符串中，整行都应该被保护
			// 如果这一行结束了多行字符串，需要处理字符串结束后的注释
			if !singleLineDocstring && inMultiLineString && !tempInMultiLine {
				// 多行字符串在这一行结束，需要找到结束位置并处理后面的注释
				var endPos int = -1
				if strings.Contains(line, `"""`) {
					endPos = strings.Index(line, `"""`) + 3
				} else if strings.Contains(line, "'''") {
					endPos = strings.Index(line, "'''") + 3
				}
				
				if endPos > 0 && endPos < len(line) {
					// 多行字符串结束后还有内容，需要处理注释
					beforeEnd := line[:endPos]
					afterEnd := line[endPos:]
					
					// 处理字符串结束后的部分
					processedAfter := afterEnd
					// 删除Python行注释
					if pos := strings.Index(processedAfter, "#"); pos != -1 {
						processedAfter = strings.TrimRight(processedAfter[:pos], " \t")
					}
					
					processedLine = beforeEnd + processedAfter
				}
			}
			
			inMultiLineString = tempInMultiLine
		}
		
		// 如果之前在多行字符串中，跳过注释处理
		if oldMultiLineState {
			result = append(result, processedLine)
			continue
		}
		
		// 如果之前在反引号字符串中但现在不在，说明模板字符串结束了，需要处理外部注释
		if oldBacktickState && !inBacktickString {
			// 这行包含了模板字符串的结束，检查外部注释
			if strings.Contains(line, "`") {
				lastBacktick := -1
				for i := len(line) - 1; i >= 0; i-- {
					if line[i] == '`' && !isEscaped(line, i) {
						lastBacktick = i
						break
					}
				}
				
				if lastBacktick != -1 && lastBacktick < len(line)-1 {
					afterBacktick := line[lastBacktick+1:]
					
					// 检查是否有注释符号
					for _, rule := range rules {
						if rule.IsLineComment {
							if pos := strings.Index(afterBacktick, rule.StartPattern); pos != -1 {
								// 找到外部注释，删除它
								beforeEnd := line[:lastBacktick+1]
								afterEnd := afterBacktick[:pos]
								processedLine = beforeEnd + strings.TrimRight(afterEnd, " \t")
								break
							}
						}
					}
				}
			}
		}
		
		// 如果当前在反引号字符串中，跳过注释处理
		if inBacktickString {
			result = append(result, originalLine)
			continue
		}
		
		
		// 如果在块注释中
		if inBlockComment {
			if pos := strings.Index(processedLine, blockEndPattern); pos != -1 {
				// 找到块注释结束，保留结束后的内容
				afterComment := processedLine[pos+len(blockEndPattern):]
				inBlockComment = false
				
				// 如果结束后还有内容，继续处理这部分内容
				if strings.TrimSpace(afterComment) != "" {
					// 递归处理剩余内容
					remaining := removeCommentsByRules(afterComment, fileType, rules)
					result = append(result, remaining)
				}
				// 如果结束后没有内容，跳过这一行（不添加空行）
			}
			// 整行都在块注释中，跳过这一行（不添加空行）
			continue
		}
		
		// 处理行注释和块注释
		for _, rule := range rules {
			if rule.IsLineComment {
				// 处理行注释：需要找到第一个不在字符串内的注释符号
				pos := -1
				// YAML特殊处理：区分结构性注释和普通注释
				if fileType == "yaml" || fileType == "yml" {
					// 遍历所有可能的#位置
					for i := 0; i <= len(processedLine)-len(rule.StartPattern); i++ {
						if strings.HasPrefix(processedLine[i:], rule.StartPattern) {
							// 检查是否在字符串内
							if isInAnyString(processedLine, i) {
								continue
							}
							
							beforeComment := processedLine[:i]
							// 如果#前只有空白字符，这是行首注释，检查是否为结构性注释
							if strings.TrimSpace(beforeComment) == "" {
								// 行首注释，检查是否需要保护（只保护结构性注释）
								if shouldProtectInContext(originalLine, i, fileType, rule.StartPattern) {
									pos = -1 // 保护结构性注释，不删除
									break
								} else {
									pos = i // 删除普通注释
									break
								}
							} else {
								// 行尾注释，检查是否需要保护（Shell变量等）
								if !shouldProtectInContext(originalLine, i, fileType, rule.StartPattern) {
									pos = i
									break
								}
							}
						}
					}
				} else {
					// 其他语言的原有逻辑
					for i := 0; i <= len(processedLine)-len(rule.StartPattern); i++ {
						if strings.HasPrefix(processedLine[i:], rule.StartPattern) {
							// 检查是否在字符串内（包括原始字符串和正则表达式）
							if !isInAnyString(originalLine, i) && !isInRegex(originalLine, i) {
								// 检查是否需要保护
								protected := shouldProtectInContext(originalLine, i, fileType, rule.StartPattern)
								if !protected {
									pos = i
									break
								}
							}
						}
					}
				}
				// 如果找到了注释位置，处理注释删除
				if pos != -1 {
					beforeComment := processedLine[:pos]
					// 如果注释前只有空白字符，则整行都是注释
					if strings.TrimSpace(beforeComment) == "" {
						processedLine = "" // 整行注释，变成空行
					} else {
						// 删除注释但去除尾部空格
						processedLine = strings.TrimRight(beforeComment, " \t")
					}
					break
				}
			} else {
				// 处理块注释
				if pos := strings.Index(processedLine, rule.StartPattern); pos != -1 {
					if !shouldProtectInContext(originalLine, pos, fileType, rule.StartPattern) && 
					   !isInAnyString(originalLine, pos) && !isInBacktickString(originalLine, pos) && !isInRegex(originalLine, pos) {
						beforeComment := processedLine[:pos]
						
						// 检查同一行是否有结束标记
						if endPos := strings.Index(processedLine[pos:], rule.EndPattern); endPos != -1 {
							// 同一行内的块注释
							actualEndPos := pos + endPos + len(rule.EndPattern)
							afterComment := processedLine[actualEndPos:]
							
							// 对于XML/HTML注释，不添加额外空格
							replacement := ""
							if fileType != "xml" && fileType != "html" && fileType != "htm" {
								// 智能处理空格：只在需要时添加空格
								needSpace := false
								if len(beforeComment) > 0 && len(afterComment) > 0 {
									lastCharBefore := beforeComment[len(beforeComment)-1]
									firstCharAfter := afterComment[0]
									if lastCharBefore != ' ' && lastCharBefore != '\t' && 
									   firstCharAfter != ' ' && firstCharAfter != '\t' && firstCharAfter != '\n' {
										needSpace = true
									}
								}
								
								if needSpace {
									replacement = " "
								}
							}
							
							processedLine = beforeComment + replacement + afterComment
						} else {
							// 跨行块注释开始
							if strings.TrimSpace(beforeComment) == "" {
								processedLine = "" // 整行注释，变成空行
							} else {
								// 保持原有的尾随空格，如果没有则添加一个
								if strings.HasSuffix(beforeComment, " ") || strings.HasSuffix(beforeComment, "\t") {
									processedLine = beforeComment
								} else {
									processedLine = beforeComment + " "
								}
							}
							inBlockComment = true
							blockEndPattern = rule.EndPattern
						}
						break
					}
				}
			}
		}
		
		// 如果处理后的行是空的且原始行不是空的，跳过这一行
		if strings.TrimSpace(processedLine) == "" && strings.TrimSpace(originalLine) != "" {
			continue
		}
		
		result = append(result, processedLine)
	}
	
	return strings.Join(result, "\n")
}

// removeCommentsByFileType 根据文件类型删除注释的统一函数
func removeCommentsByFileType(content, fileType string) string {
	rules := getCommentRulesForLanguage(fileType)
	return removeCommentsByRules(content, fileType, rules)
}

// removeComments 移除指定文件类型的注释
func removeComments(content, fileType string) string {
	// 统一使用规则处理所有文件类型
	return removeCommentsByFileType(content, fileType)
}
