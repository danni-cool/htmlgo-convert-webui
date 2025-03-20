package parser_utils

import (
	"strings"
)

// Error categories
const (
	ErrSyntax      = "syntax"
	ErrUndefined   = "undefined"
	ErrType        = "type"
	ErrCompilation = "compilation"
	ErrExecution   = "execution"
	ErrParse       = "parse"
)

// ErrorHandler processes error messages into human-readable format
type ErrorHandler struct {
	ErrorPatterns map[string]string
}

// NewErrorHandler creates a new ErrorHandler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		ErrorPatterns: map[string]string{
			"syntax error: unexpected if, expected expression": ErrSyntax,
			"undefined:":           ErrUndefined,
			"cannot use":           ErrType,
			"syntax error":         ErrSyntax,
			"not enough arguments": ErrCompilation,
		},
	}
}

// GetFriendlyErrorMessage generates a user-friendly error message
func (h *ErrorHandler) GetFriendlyErrorMessage(errMsg string) string {
	// Check for specific error patterns
	for pattern, category := range h.ErrorPatterns {
		if strings.Contains(errMsg, pattern) {
			switch category {
			case ErrSyntax:
				if strings.Contains(errMsg, "unexpected if, expected expression") {
					return h.getIfExpressionErrorMessage()
				}
				return "<!-- 语法错误: " + errMsg + " -->\n<!-- 请检查您的代码语法是否正确，包括括号、逗号等 -->"

			case ErrUndefined:
				return "<!-- 错误: 找不到变量或函数: " + errMsg + " -->\n<!-- 请确保您已正确导入所有必要的包，并且变量名拼写正确 -->"

			case ErrType:
				return "<!-- 类型错误: " + errMsg + " -->\n<!-- 请确保变量类型与函数期望的类型一致 -->"

			case ErrCompilation:
				return "<!-- 编译错误: " + errMsg + " -->"
			}
		}
	}

	// Default error message for unrecognized errors
	return "<!-- 错误: " + errMsg + " -->"
}

// Direct if expression error - most common error
func (h *ErrorHandler) getIfExpressionErrorMessage() string {
	return `<!-- 编译或执行错误: syntax error: unexpected if, expected expression -->
<!-- 请尝试以下正确写法: -->
<!--
    // 方法1: 使用函数
    var n = htmlgo.Div().Text(func() string {
        if condition {
            return "真"
        }
        return "假"
    }())
    
    // 方法2: 使用map模拟三元运算符
    condition := true
    var n = htmlgo.Div().Text(map[bool]string{true: "真", false: "假"}[condition])
-->
`
}
