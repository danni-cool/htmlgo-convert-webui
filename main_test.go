package main

import (
	"strings"
	"testing"
)

func TestFormatGoCode(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		packagePrefix string
		want          string
		wantErr       bool
	}{
		{
			name:          "基础嵌套格式化测试",
			input:         `Div(H1(Text("Hello")).Class("title"), P(Text("World")).Class("text")).Class("container")`,
			packagePrefix: "",
			want: `	Div(
		H1(
			Text("Hello"),
		).Class("title"),
		P(
			Text("World"),
		).Class("text"),
	).Class("container")`,
			wantErr: false,
		},
		{
			name:          "带包前缀的嵌套格式化测试",
			input:         `htmlgo.Div(htmlgo.H1(htmlgo.Text("Hello")).Class("title"), htmlgo.P(htmlgo.Text("World")).Class("text")).Class("container")`,
			packagePrefix: "h",
			want: `	h.Div(
		h.H1(
			h.Text("Hello"),
		).Class("title"),
		h.P(
			h.Text("World"),
		).Class("text"),
	).Class("container")`,
			wantErr: false,
		},
		{
			name:          "带自定义包前缀的嵌套格式化测试",
			input:         `htmlgo.Div(htmlgo.H1(htmlgo.Text("Hello")).Class("title"), htmlgo.P(htmlgo.Text("World")).Class("text")).Class("container")`,
			packagePrefix: "html",
			want: `	html.Div(
		html.H1(
			html.Text("Hello"),
		).Class("title"),
		html.P(
			html.Text("World"),
		).Class("text"),
	).Class("container")`,
			wantErr: false,
		},
		{
			name: "复杂嵌套结构测试",
			input: `Div(
				Div(
					H1(Text("Title")).Class("header"),
					P(Text("Content")).Class("content"),
				).Class("inner"),
				Button(Text("Click me")).Class("btn"),
			).Class("outer")`,
			packagePrefix: "",
			want: `	Div(
		Div(
			H1(
				Text("Title"),
			).Class("header"),
			P(
				Text("Content"),
			).Class("content"),
		).Class("inner"),
		Button(
			Text("Click me"),
		).Class("btn"),
	).Class("outer")`,
			wantErr: false,
		},
		{
			name:          "多属性格式化测试",
			input:         `Input("").Type("text").Class("input").Id("username").Attr("placeholder", "Enter username").Attr("required", "required")`,
			packagePrefix: "",
			want: `	Input("").
		Type("text").
		Class("input").
		Id("username").
		Attr("placeholder", "Enter username").
		Attr("required", "required")`,
			wantErr: false,
		},
		{
			name:          "带包前缀的多属性格式化测试",
			input:         `htmlgo.Input("").Type("text").Class("input").Id("username").Attr("placeholder", "Enter username").Attr("required", "required")`,
			packagePrefix: "h",
			want: `	h.Input("").
		Type("text").
		Class("input").
		Id("username").
		Attr("placeholder", "Enter username").
		Attr("required", "required")`,
			wantErr: false,
		},
		{
			name:          "带长包前缀的多属性格式化测试",
			input:         `htmlgo.Input("").Type("text").Class("input").Id("username").Attr("placeholder", "Enter username").Attr("required", "required")`,
			packagePrefix: "myhtml",
			want: `	myhtml.Input("").
		Type("text").
		Class("input").
		Id("username").
		Attr("placeholder", "Enter username").
		Attr("required", "required")`,
			wantErr: false,
		},
		{
			name: "带包前缀的复杂嵌套结构测试",
			input: `htmlgo.Div(
				htmlgo.Div(
					htmlgo.H1(htmlgo.Text("Title")).Class("header"),
					htmlgo.P(htmlgo.Text("Content")).Class("content"),
				).Class("inner"),
				htmlgo.Button(htmlgo.Text("Click me")).Class("btn"),
			).Class("outer")`,
			packagePrefix: "h",
			want: `	h.Div(
		h.Div(
			h.H1(
				h.Text("Title"),
			).Class("header"),
			h.P(
				h.Text("Content"),
			).Class("content"),
		).Class("inner"),
		h.Button(
			h.Text("Click me"),
		).Class("btn"),
	).Class("outer")`,
			wantErr: false,
		},
		{
			name:          "新格式的多属性测试",
			input:         `Input("").Type("text").Class("input").Attr("placeholder", "Enter text").Attr("required", "required")`,
			packagePrefix: "",
			want: `	Input("").
		Type("text").
		Class("input").
		Attr("placeholder", "Enter text").
		Attr("required", "required")`,
			wantErr: false,
		},
		{
			name:          "新格式的带包前缀多属性测试",
			input:         `htmlgo.Input("").Type("text").Class("input").Attr("placeholder", "Enter text").Attr("required", "required")`,
			packagePrefix: "h",
			want: `	h.Input("").
		Type("text").
		Class("input").
		Attr("placeholder", "Enter text").
		Attr("required", "required")`,
			wantErr: false,
		},
		{
			name:          "单个input元素格式化测试",
			input:         `Input("").Type("text").Class("input").Attr("placeholder", "Enter text").Attr("required", "required")`,
			packagePrefix: "",
			want: `	Input("").
		Type("text").
		Class("input").
		Attr("placeholder", "Enter text").
		Attr("required", "required")`,
			wantErr: false,
		},
		{
			name:          "单个input元素带包前缀格式化测试",
			input:         `htmlgo.Input("").Type("text").Class("input").Attr("placeholder", "Enter text").Attr("required", "required")`,
			packagePrefix: "h",
			want: `	h.Input("").
		Type("text").
		Class("input").
		Attr("placeholder", "Enter text").
		Attr("required", "required")`,
			wantErr: false,
		},
		{
			name:          "单个input元素带包前缀格式化测试2",
			input:         `Input("").Type("text").Class("input").Attr("placeholder", "Enter text").Attr("required", "required")`,
			packagePrefix: "h",
			want: `	h.Input("").
		Type("text").
		Class("input").
		Attr("placeholder", "Enter text").
		Attr("required", "required")`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatGoCode(tt.input, tt.packagePrefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("formatGoCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 规范化空白字符进行比较
			gotNormalized := normalizeWhitespace(got)
			wantNormalized := normalizeWhitespace(tt.want)

			if gotNormalized != wantNormalized {
				t.Errorf("formatGoCode() = \n%v\nwant\n%v", got, tt.want)
				// 打印详细的差异
				t.Logf("Got lines:")
				for i, line := range strings.Split(got, "\n") {
					t.Logf("%d: %q", i+1, line)
				}
				t.Logf("Want lines:")
				for i, line := range strings.Split(tt.want, "\n") {
					t.Logf("%d: %q", i+1, line)
				}
			}
		})
	}
}

// 规范化空白字符，用于比较
func normalizeWhitespace(s string) string {
	// 将所有空白字符替换为单个空格
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		// 保留缩进的制表符
		indent := strings.Repeat("\t", strings.Count(line, "\t"))
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			lines[i] = indent + trimmed
		} else {
			lines[i] = ""
		}
	}
	// 移除空行
	var nonEmptyLines []string
	for _, line := range lines {
		if line != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}
	return strings.Join(nonEmptyLines, "\n")
}
