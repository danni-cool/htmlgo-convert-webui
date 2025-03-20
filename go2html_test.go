package main

import (
	"fmt"
	"testing"
)

func TestGoToHTMLSpecificCases(t *testing.T) {
	testCases := []struct {
		name     string
		goCode   string
		expected string
		hasError bool
	}{
		{
			name:     "基本元素",
			goCode:   `var n = htmlgo.Div().Text("Hello World")`,
			expected: "<div>Hello World</div>",
			hasError: false,
		},
		{
			name:     "带属性的元素",
			goCode:   `var n = htmlgo.Div().Class("container").Text("Hello World")`,
			expected: `<div class="container">Hello World</div>`,
			hasError: false,
		},
		{
			name:     "嵌套元素",
			goCode:   `var n = htmlgo.Div().Children(htmlgo.H1().Text("标题"), htmlgo.P().Text("段落"))`,
			expected: "<div><h1>标题</h1><p>段落</p></div>",
			hasError: false,
		},
		{
			name:     "使用n :=语法",
			goCode:   `n := htmlgo.Div().Text("使用:=语法")`,
			expected: "<div>使用:=语法</div>",
			hasError: false,
		},
		{
			name:     "不包含n变量",
			goCode:   `htmlgo.Div().Text("没有n变量")`,
			expected: "<!-- 警告: 没有生成HTML输出，请检查您的代码是否正确定义了变量 'n' -->",
			hasError: false,
		},
		{
			name:     "语法错误",
			goCode:   `var n = if true { htmlgo.Div() } else { htmlgo.Span() }`,
			expected: "编译或执行错误",
			hasError: true,
		},
		{
			name:     "使用h包前缀",
			goCode:   `var n = h.Div().Text("使用h包前缀")`,
			expected: "<div>使用h包前缀</div>",
			hasError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := convertGoToHTML(tc.goCode)

			if err != nil && !tc.hasError {
				t.Errorf("期望无错误，但得到错误: %v", err)
			}

			if err == nil && tc.hasError {
				t.Errorf("期望有错误，但没有得到错误")
			}

			// 对于预期有错误的情况，只检查结果是否包含预期的错误信息
			if tc.hasError {
				if result == "" || !contains(result, tc.expected) {
					t.Errorf("期望结果包含 %q，但得到 %q", tc.expected, result)
				}
				return
			}

			// 对于预期无错误的情况，检查结果是否与预期匹配
			// 忽略空白字符进行比较
			cleanResult := removeWhitespace(result)
			cleanExpected := removeWhitespace(tc.expected)

			if !contains(cleanResult, cleanExpected) {
				t.Errorf("\n期望: %q\n得到: %q", tc.expected, result)
			}

			// 打印测试结果，方便调试
			fmt.Printf("测试用例: %s\n输入: %s\n输出: %s\n\n", tc.name, tc.goCode, result)
		})
	}
}

// 辅助函数：检查字符串是否包含子串（忽略大小写）
func contains(s, substr string) bool {
	return s != "" && (s == substr || removeWhitespace(s) == removeWhitespace(substr) ||
		containsIgnoreCase(removeWhitespace(s), removeWhitespace(substr)))
}

// 辅助函数：忽略大小写检查字符串是否包含子串
func containsIgnoreCase(s, substr string) bool {
	return s != "" && substr != "" && (s == substr ||
		fmt.Sprintf("%v", s) == fmt.Sprintf("%v", substr) ||
		fmt.Sprintf("%v", s) == substr || s == fmt.Sprintf("%v", substr))
}

// 辅助函数：移除字符串中的所有空白字符
func removeWhitespace(s string) string {
	var result []rune
	for _, r := range s {
		if r != ' ' && r != '\n' && r != '\t' && r != '\r' {
			result = append(result, r)
		}
	}
	return string(result)
}
