package converter

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestConvertHTMLToGo(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "基础div测试",
			input:    `<div class="container">Hello</div>`,
			expected: `htmlgo.Div(htmlgo.Text("Hello")).Class("container")`,
			wantErr:  false,
		},
		{
			name: "嵌套元素测试",
			input: `<div class="container">
				<h1 class="title">Hello</h1>
				<p class="text">World</p>
			</div>`,
			expected: `htmlgo.Div(htmlgo.H1(htmlgo.Text("Hello")).Class("title"), htmlgo.P(htmlgo.Text("World")).Class("text")).Class("container")`,
			wantErr:  false,
		},
		{
			name:     "多属性测试",
			input:    `<input type="text" id="username" class="input" placeholder="Enter username" required>`,
			expected: `htmlgo.Input("").Type("text").Id("username").Class("input").Attr("placeholder", "Enter username").Attr("required", "required")`,
			wantErr:  false,
		},
		{
			name:     "空元素测试",
			input:    `<div></div>`,
			expected: `htmlgo.Div()`,
			wantErr:  false,
		},
		{
			name:     "样式属性测试",
			input:    `<div style="color: red; font-size: 16px;">Test</div>`,
			expected: `htmlgo.Div(htmlgo.Text("Test")).Style("color: red; font-size: 16px;")`,
			wantErr:  false,
		},
		{
			name:    "无效HTML测试",
			input:   `<div><unclosed>`,
			wantErr: true,
		},
		// Vue 组件测试用例
		{
			name:     "Vue v-if指令测试",
			input:    `<div v-if="isVisible" class="container">Content</div>`,
			expected: `htmlgo.Div(htmlgo.Text("Content")).Class("container").Attr("v-if", "isVisible")`,
			wantErr:  false,
		},
		{
			name:     "Vue v-for指令测试",
			input:    `<li v-for="item in items" class="item">{{ item.name }}</li>`,
			expected: `htmlgo.Li().Class("item").Attr("v-for", "item in items").Attr("v-text", " item.name ")`,
			wantErr:  false,
		},
		{
			name:     "Vue v-model指令测试",
			input:    `<input v-model="message" type="text" class="input">`,
			expected: `htmlgo.Input("").Type("text").Class("input").Attr("v-model", "message")`,
			wantErr:  false,
		},
		{
			name:     "Vue 简写属性绑定测试",
			input:    `<img :src="imageUrl" :alt="imageAlt" class="image">`,
			expected: `htmlgo.Img().Class("image").Attr(":src", "imageUrl").Attr(":alt", "imageAlt")`,
			wantErr:  false,
		},
		{
			name:     "Vue 事件处理测试",
			input:    `<button @click="handleClick" class="btn">Click Me</button>`,
			expected: `htmlgo.Button(htmlgo.Text("Click Me")).Class("btn").Attr("@click", "handleClick")`,
			wantErr:  false,
		},
		{
			name:     "Vue 插值表达式测试",
			input:    `<p class="text">{{ message }}</p>`,
			expected: `htmlgo.P().Class("text").Attr("v-text", " message ")`,
			wantErr:  false,
		},
		{
			name: "Vue 复杂组件测试",
			input: `<div class="app">
				<h1 v-if="showTitle" class="title">{{ title }}</h1>
				<ul class="list">
					<li v-for="(item, index) in items" :key="index" @click="selectItem(item)" class="item">
						{{ item.name }} - {{ item.price }}
					</li>
				</ul>
				<input v-model="newItem" type="text" class="input" placeholder="Add new item">
				<button @click="addItem" class="btn">Add</button>
			</div>`,
			expected: `htmlgo.Div(htmlgo.H1().Class("title").Attr("v-if", "showTitle").Attr("v-text", " title "), htmlgo.Ul(htmlgo.Li().Class("item").Attr("v-for", "(item, index) in items").Attr(":key", "index").Attr("@click", "selectItem(item)").Attr("v-text", " item.name  -  item.price ")).Class("list"), htmlgo.Input("").Type("text").Class("input").Attr("placeholder", "Add new item").Attr("v-model", "newItem"), htmlgo.Button(htmlgo.Text("Add")).Class("btn").Attr("@click", "addItem")).Class("app")`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 打印HTML解析结果
			doc, err := html.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Logf("HTML解析错误: %v", err)
			} else {
				var printNode func(*html.Node, int)
				printNode = func(n *html.Node, depth int) {
					indent := strings.Repeat("  ", depth)
					switch n.Type {
					case html.ElementNode:
						t.Logf("%s<%s>", indent, n.Data)
					case html.TextNode:
						text := strings.TrimSpace(n.Data)
						if text != "" {
							t.Logf("%s%q", indent, text)
						}
					}
					for c := n.FirstChild; c != nil; c = c.NextSibling {
						printNode(c, depth+1)
					}
				}
				t.Log("HTML解析结果:")
				printNode(doc, 0)
			}

			got, err := ConvertHTMLToGo(tt.input)

			// 错误检查
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertHTMLToGo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			// 移除所有空白字符后比较
			gotCleaned := removeWhitespace(got)
			expectedCleaned := removeWhitespace(tt.expected)

			if gotCleaned != expectedCleaned {
				t.Errorf("ConvertHTMLToGo() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// 移除所有空白字符，用于比较
func removeWhitespace(s string) string {
	return strings.Join(strings.Fields(s), "")
}

// 测试特殊情况
func TestConvertHTMLToGoSpecialCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "空输入测试",
			input:    "",
			expected: "htmlgo.",
			wantErr:  false,
		},
		{
			name:     "只有文本测试",
			input:    "Hello World",
			expected: `htmlgo.Text("Hello World")`,
			wantErr:  false,
		},
		{
			name:     "HTML实体测试",
			input:    `<div>&lt;Hello&gt;</div>`,
			expected: `htmlgo.Div(htmlgo.Text("<Hello>"))`,
			wantErr:  false,
		},
		{
			name:     "数据属性测试",
			input:    `<div data-test="value">Content</div>`,
			expected: `htmlgo.Div(htmlgo.Text("Content")).Attr("data-test", "value")`,
			wantErr:  false,
		},
		{
			name:     "布尔属性测试",
			input:    `<input type="checkbox" checked disabled>`,
			expected: `htmlgo.Input("").Type("checkbox").Attr("checked", "checked").Attr("disabled", "disabled")`,
			wantErr:  false,
		},
		// Vue 特殊情况测试
		{
			name:     "Vue 多个指令组合测试",
			input:    `<input v-model="form.name" :placeholder="placeholderText" @input="validateInput" type="text" class="form-control">`,
			expected: `htmlgo.Input("").Type("text").Class("form-control").Attr("v-model", "form.name").Attr(":placeholder", "placeholderText").Attr("@input", "validateInput")`,
			wantErr:  false,
		},
		{
			name:     "Vue 动态类名测试",
			input:    `<div :class="{ active: isActive, 'text-danger': hasError }" class="base">Content</div>`,
			expected: `htmlgo.Div(htmlgo.Text("Content")).Class("base").Attr(":class", "{ active: isActive, 'text-danger': hasError }")`,
			wantErr:  false,
		},
		{
			name:     "Vue 动态样式测试",
			input:    `<div :style="{ color: activeColor, fontSize: fontSize + 'px' }" class="styled">Text</div>`,
			expected: `htmlgo.Div(htmlgo.Text("Text")).Class("styled").Attr(":style", "{ color: activeColor, fontSize: fontSize + 'px' }")`,
			wantErr:  false,
		},
		{
			name:     "Vue 插槽测试",
			input:    `<slot name="header">Default header</slot>`,
			expected: `htmlgo.Slot(htmlgo.Text("Default header")).Attr("name", "header")`,
			wantErr:  false,
		},
		{
			name:     "Vue 组件引用测试",
			input:    `<my-component :prop="value" @event="handler">Content</my-component>`,
			expected: `htmlgo.MyComponent(htmlgo.Text("Content")).Attr(":prop", "value").Attr("@event", "handler")`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 打印HTML解析结果
			doc, err := html.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Logf("HTML解析错误: %v", err)
			} else {
				var printNode func(*html.Node, int)
				printNode = func(n *html.Node, depth int) {
					indent := strings.Repeat("  ", depth)
					switch n.Type {
					case html.ElementNode:
						t.Logf("%s<%s>", indent, n.Data)
					case html.TextNode:
						text := strings.TrimSpace(n.Data)
						if text != "" {
							t.Logf("%s%q", indent, text)
						}
					}
					for c := n.FirstChild; c != nil; c = c.NextSibling {
						printNode(c, depth+1)
					}
				}
				t.Log("HTML解析结果:")
				printNode(doc, 0)
			}

			got, err := ConvertHTMLToGo(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertHTMLToGo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			gotCleaned := removeWhitespace(got)
			expectedCleaned := removeWhitespace(tt.expected)

			if gotCleaned != expectedCleaned {
				t.Errorf("ConvertHTMLToGo() = %v, want %v", got, tt.expected)
			}
		})
	}
}
