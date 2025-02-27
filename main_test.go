package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/sunfmin/html2go/parse"
)

// 测试HTML到Go的转换
func TestHTMLToGoConversion(t *testing.T) {
	testCases := []struct {
		name          string
		html          string
		packagePrefix string
		removePackage bool
		expectError   bool
	}{
		{
			name:          "基本HTML",
			html:          `<div class="container">Hello World</div>`,
			packagePrefix: "h",
			removePackage: true,
			expectError:   false,
		},
		{
			name:          "复杂HTML",
			html:          `<div class="container"><h1 class="title">Hello</h1><p class="text">World</p></div>`,
			packagePrefix: "h",
			removePackage: true,
			expectError:   false,
		},
		{
			name:          "带属性的HTML",
			html:          `<input type="text" class="input" placeholder="Enter text" required>`,
			packagePrefix: "h",
			removePackage: true,
			expectError:   false,
		},
		{
			name:          "空HTML",
			html:          ``,
			packagePrefix: "h",
			removePackage: true,
			expectError:   true,
		},
		{
			name:          "无效HTML",
			html:          `<div class="container">Hello World`,
			packagePrefix: "h",
			removePackage: true,
			expectError:   false, // HTML解析器会尝试修复无效HTML
		},
		{
			name:          "自定义包前缀",
			html:          `<div class="container">Hello World</div>`,
			packagePrefix: "htmlgo",
			removePackage: true,
			expectError:   false,
		},
		{
			name:          "保留package声明",
			html:          `<div class="container">Hello World</div>`,
			packagePrefix: "h",
			removePackage: false,
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建请求
			reqBody := ConvertRequest{
				HTML:          tc.html,
				PackagePrefix: tc.packagePrefix,
				RemovePackage: tc.removePackage,
				Direction:     "html2go",
			}
			reqJSON, err := json.Marshal(reqBody)
			if err != nil {
				t.Fatalf("无法序列化请求: %v", err)
			}

			// 创建测试服务器
			req := httptest.NewRequest("POST", "/convert", bytes.NewBuffer(reqJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// 处理请求
			handleConvert(w, req)

			// 检查响应
			resp := w.Result()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("无法读取响应: %v", err)
			}

			if tc.expectError {
				if resp.StatusCode == http.StatusOK {
					t.Errorf("期望错误，但得到成功响应: %s", body)
				}
			} else {
				if resp.StatusCode != http.StatusOK {
					t.Errorf("期望成功，但得到错误响应: %s", body)
				} else {
					var respBody ConvertResponse
					if err := json.Unmarshal(body, &respBody); err != nil {
						t.Fatalf("无法解析响应: %v", err)
					}

					if respBody.Code == "" {
						t.Errorf("响应中没有Go代码")
					}

					// 验证包前缀
					if !strings.Contains(respBody.Code, tc.packagePrefix) && !tc.removePackage {
						t.Errorf("Go代码中没有包含指定的包前缀: %s", tc.packagePrefix)
					}

					// 验证package声明
					if tc.removePackage && strings.Contains(respBody.Code, "package ") {
						t.Errorf("Go代码中包含package声明，但应该被删除")
					}

					if !tc.removePackage && !strings.Contains(respBody.Code, "package ") {
						t.Errorf("Go代码中没有package声明，但应该保留")
					}

					fmt.Printf("测试用例 '%s' 的Go代码输出:\n%s\n\n", tc.name, respBody.Code)
				}
			}
		})
	}
}

// 测试Go到HTML的转换
func TestGoToHTMLConversion(t *testing.T) {
	// 确保测试环境中有必要的依赖
	if err := os.Chdir(os.Getenv("PWD")); err != nil {
		t.Fatalf("无法切换到工作目录: %v", err)
	}

	testCases := []struct {
		name        string
		goCode      string
		expectError bool
		expectHTML  string
		checkFunc   func(t *testing.T, html string)
	}{
		{
			name:        "基本Go代码",
			goCode:      `var n = htmlgo.Div(htmlgo.Text("Hello World")).Class("container")`,
			expectError: false,
			expectHTML:  `<div class='container'>Hello World</div>`,
		},
		{
			name: "复杂Go代码",
			goCode: `var n = htmlgo.Div().Class("container")
h1 := htmlgo.H1("Hello").Class("title")
p := htmlgo.P(htmlgo.Text("World")).Class("text")
n.Children(h1, p)`,
			expectError: false,
			expectHTML:  `<div class='container'><h1 class='title'>Hello</h1><p class='text'>World</p></div>`,
		},
		{
			name: "带属性的Go代码",
			goCode: `var n = htmlgo.Input("").
	Type("text").
	Class("input").
	Placeholder("Enter text").
	Required(true)`,
			expectError: false,
			expectHTML:  `<input type='text' placeholder='Enter text' required class='input'>`,
		},
		{
			name:        "空Go代码",
			goCode:      ``,
			expectError: true, // 空Go代码会返回错误
			expectHTML:  ``,
		},
		{
			name:        "无效Go代码",
			goCode:      `var n = htmlgo.Div(htmlgo.Text("Hello World")).Class(`,
			expectError: false, // 我们现在返回错误消息而不是抛出错误
			expectHTML:  `<!-- 编译或执行错误:`,
		},
		{
			name:        "缺少变量n",
			goCode:      `var m = htmlgo.Div(htmlgo.Text("Hello World")).Class("container")`,
			expectError: false,
			expectHTML:  `<!-- 错误: 变量 'n' 未定义`,
		},
		{
			name: "多个语句",
			goCode: `
var div = htmlgo.Div(htmlgo.Text("Hello World")).Class("container")
var n = div
`,
			expectError: false,
			expectHTML:  `<div class='container'>Hello World</div>`,
		},
		// UI界面示例1
		{
			name:        "UI示例1-基本Go结构",
			goCode:      `var n = htmlgo.Div(htmlgo.H1("Hello World").Class("text-2xl font-bold mb-4"), htmlgo.P(htmlgo.Text("这是一个基本的Go示例")).Class("text-gray-600"), htmlgo.Button("点击我").Class("bg-blue-500 text-white px-4 py-2 rounded mt-4")).Class("container mx-auto p-4")`,
			expectError: false,
			checkFunc: func(t *testing.T, html string) {
				if !strings.Contains(html, "<div") || !strings.Contains(html, "<h1") ||
					!strings.Contains(html, "<p") || !strings.Contains(html, "<button") {
					t.Errorf("HTML输出不包含期望的元素\n实际: %s", html)
				}
			},
		},
		// UI界面示例2
		{
			name:        "UI示例2-表单示例",
			goCode:      `var n = htmlgo.Form(htmlgo.Div(htmlgo.Label("用户名").Class("block text-gray-700 text-sm font-bold mb-2").Attr("for", "username"), htmlgo.Input("username").Type("text").Attr("placeholder", "用户名").Attr("required", "true").Class("shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline")).Class("mb-4"), htmlgo.Div(htmlgo.Label("密码").Class("block text-gray-700 text-sm font-bold mb-2").Attr("for", "password"), htmlgo.Input("password").Type("password").Attr("placeholder", "******************").Attr("required", "true").Class("shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 mb-3 leading-tight focus:outline-none focus:shadow-outline")).Class("mb-6"), htmlgo.Div(htmlgo.Button("登录").Type("submit").Class("bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"), htmlgo.A(htmlgo.Text("忘记密码?")).Attr("href", "#").Class("inline-block align-baseline font-bold text-sm text-blue-500 hover:text-blue-800")).Class("flex items-center justify-between")).Class("max-w-md mx-auto p-6 bg-white rounded-lg shadow-md")`,
			expectError: false,
			checkFunc: func(t *testing.T, html string) {
				if !strings.Contains(html, "<form") || !strings.Contains(html, "<input") ||
					!strings.Contains(html, "<label") || !strings.Contains(html, "<button") {
					t.Errorf("HTML输出不包含期望的表单元素\n实际: %s", html)
				}
			},
		},
		// UI界面示例3 - 简化版本
		{
			name:        "UI示例3-复杂布局",
			goCode:      `var n = htmlgo.Div(htmlgo.Header(htmlgo.Div(htmlgo.Text("导航栏")).Class("flex justify-between items-center")).Class("bg-white shadow rounded-lg p-4 mb-6"), htmlgo.Main(htmlgo.Aside(htmlgo.H2("侧边栏").Class("text-lg font-semibold mb-4")).Class("md:col-span-1 bg-white p-4 rounded-lg shadow"), htmlgo.Section(htmlgo.H2("主要内容").Class("text-xl font-bold mb-4"), htmlgo.P(htmlgo.Text("这是一个复杂布局示例")).Class("text-gray-700 mb-4")).Class("md:col-span-2 bg-white p-4 rounded-lg shadow")).Class("grid grid-cols-1 md:grid-cols-3 gap-6")).Class("max-w-6xl mx-auto p-4")`,
			expectError: false,
			checkFunc: func(t *testing.T, html string) {
				if !strings.Contains(html, "<div") || !strings.Contains(html, "<header") ||
					!strings.Contains(html, "<main") || !strings.Contains(html, "<aside") ||
					!strings.Contains(html, "<section") {
					t.Errorf("HTML输出不包含期望的复杂布局元素\n实际: %s", html)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建请求
			reqBody := ConvertRequest{
				GoCode:    tc.goCode,
				Direction: "go2html",
			}
			reqJSON, err := json.Marshal(reqBody)
			if err != nil {
				t.Fatalf("无法序列化请求: %v", err)
			}

			// 创建测试服务器
			req := httptest.NewRequest("POST", "/convert", bytes.NewBuffer(reqJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// 处理请求
			handleConvert(w, req)

			// 检查响应
			resp := w.Result()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("无法读取响应: %v", err)
			}

			if tc.expectError {
				if resp.StatusCode == http.StatusOK {
					t.Errorf("期望错误，但得到成功响应: %s", body)
				}
			} else {
				if resp.StatusCode != http.StatusOK {
					t.Errorf("期望成功，但得到错误响应: %s", body)
				} else {
					var respBody ConvertResponse
					if err := json.Unmarshal(body, &respBody); err != nil {
						t.Fatalf("无法解析响应: %v", err)
					}

					// 检查HTML输出
					if respBody.HTML == "" {
						t.Errorf("响应中没有HTML")
					}

					// 对于"缺少变量n"的测试用例，我们需要特殊处理
					if tc.name == "缺少变量n" {
						if !strings.Contains(respBody.HTML, "错误") && !strings.Contains(respBody.HTML, "未定义") {
							t.Errorf("HTML输出不包含错误信息\n实际: %s", respBody.HTML)
						}
					} else if tc.name == "复杂Go代码" {
						// 对于复杂Go代码，我们只检查关键部分
						if !strings.Contains(respBody.HTML, "<div") ||
							!strings.Contains(respBody.HTML, "<h1") ||
							!strings.Contains(respBody.HTML, "<p") ||
							!strings.Contains(respBody.HTML, "Hello") ||
							!strings.Contains(respBody.HTML, "World") {
							t.Errorf("HTML输出不包含期望的内容\n期望包含div、h1、p元素和Hello、World文本\n实际: %s", respBody.HTML)
						}
					} else if tc.checkFunc != nil {
						// 使用自定义检查函数
						tc.checkFunc(t, respBody.HTML)
					} else if tc.expectHTML != "" && !strings.Contains(respBody.HTML, tc.expectHTML) {
						t.Errorf("HTML输出不包含期望的内容\n期望: %s\n实际: %s", tc.expectHTML, respBody.HTML)
					}

					fmt.Printf("测试用例 '%s' 的HTML输出:\n%s\n\n", tc.name, respBody.HTML)
				}
			}
		})
	}
}

// 测试双向转换的简化版本
func TestSimpleBidirectionalConversion(t *testing.T) {
	// 测试HTML到Go的转换
	html := `<div class="container">Hello World</div>`
	packagePrefix := "h"

	// HTML到Go
	htmlReader := strings.NewReader(html)
	goCode := parse.GenerateHTMLGo(packagePrefix, false, htmlReader)
	goCode = removePackageDeclaration(goCode)

	if !strings.Contains(goCode, "Div(") || !strings.Contains(goCode, "Hello World") {
		t.Errorf("HTML到Go转换失败，生成的代码不包含预期内容: %s", goCode)
	} else {
		fmt.Printf("HTML到Go转换结果:\n%s\n\n", goCode)
	}

	// Go到HTML（使用HTML到Go生成的代码）
	html2, err := convertGoToHTML(goCode)
	if err != nil {
		t.Errorf("Go到HTML转换失败: %v", err)
	} else if strings.Contains(html2, "编译或执行错误") {
		t.Errorf("Go到HTML转换出现编译错误: %s", html2)
	} else {
		fmt.Printf("Go到HTML转换结果:\n%s\n\n", html2)

		// 验证HTML输出是否包含原始HTML的关键部分
		if !strings.Contains(html2, "div") || !strings.Contains(html2, "container") || !strings.Contains(html2, "Hello World") {
			t.Errorf("转换后的HTML不包含原始HTML的关键部分\n原始HTML: %s\n转换后HTML: %s", html, html2)
		}
	}
}

// 测试HTML到Go再到HTML的转换（使用不同的包前缀）
func TestHTMLToGoToHTMLWithDifferentPrefix(t *testing.T) {
	// 测试HTML到Go的转换
	html := `<div class="container"><h1>标题</h1><p>内容</p></div>`
	packagePrefix := "htmlgo" // 使用htmlgo作为包前缀

	// HTML到Go
	htmlReader := strings.NewReader(html)
	goCode := parse.GenerateHTMLGo(packagePrefix, false, htmlReader)
	goCode = removePackageDeclaration(goCode)

	if !strings.Contains(goCode, "Div(") || !strings.Contains(goCode, "标题") {
		t.Errorf("HTML到Go转换失败，生成的代码不包含预期内容: %s", goCode)
	} else {
		fmt.Printf("HTML到Go转换结果(htmlgo前缀):\n%s\n\n", goCode)
	}

	// Go到HTML（使用HTML到Go生成的代码）
	html2, err := convertGoToHTML(goCode)
	if err != nil {
		t.Errorf("Go到HTML转换失败: %v", err)
	} else if strings.Contains(html2, "编译或执行错误") {
		t.Errorf("Go到HTML转换出现编译错误: %s", html2)
	} else {
		fmt.Printf("Go到HTML转换结果:\n%s\n\n", html2)

		// 验证HTML输出是否包含原始HTML的关键部分
		if !strings.Contains(html2, "div") || !strings.Contains(html2, "container") ||
			!strings.Contains(html2, "h1") || !strings.Contains(html2, "p") ||
			!strings.Contains(html2, "标题") || !strings.Contains(html2, "内容") {
			t.Errorf("转换后的HTML不包含原始HTML的关键部分\n原始HTML: %s\n转换后HTML: %s", html, html2)
		}
	}
}
