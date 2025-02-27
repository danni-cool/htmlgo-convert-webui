package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sunfmin/html2go/parse"
)

// 测试HTML到Go的转换
func TestHTMLToGo(t *testing.T) {
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
func TestGoToHTML(t *testing.T) {
	testCases := []struct {
		name        string
		goCode      string
		expectError bool
		expectHTML  string
	}{
		{
			name:        "基本Go代码",
			goCode:      `var n = h.Div("Hello World").Class("container")`,
			expectError: false,
			expectHTML:  `<div class="container">Hello World</div>`,
		},
		{
			name: "复杂Go代码",
			goCode: `var n = h.Div(
				h.H1("Hello").Class("title"),
				h.P("World").Class("text")
			).Class("container")`,
			expectError: false,
			expectHTML:  `<div class="container"><h1 class="title">Hello</h1><p class="text">World</p></div>`,
		},
		{
			name: "带属性的Go代码",
			goCode: `var n = h.Input("").
				Type("text").
				Class("input").
				Placeholder("Enter text").
				Required(true)`,
			expectError: false,
			expectHTML:  `<input class="input" placeholder="Enter text" required type="text">`,
		},
		{
			name:        "空Go代码",
			goCode:      ``,
			expectError: false, // 不会报错，但会返回警告
			expectHTML:  `<!-- 警告: 没有生成HTML输出`,
		},
		{
			name:        "无效Go代码",
			goCode:      `var n = h.Div("Hello World").Class(`,
			expectError: false, // 我们现在返回错误消息而不是抛出错误
			expectHTML:  `<!-- 编译或执行错误:`,
		},
		{
			name:        "缺少变量n",
			goCode:      `var m = h.Div("Hello World").Class("container")`,
			expectError: false,
			expectHTML:  `<!-- 错误: 变量 'n' 未定义或为nil -->`,
		},
		{
			name: "多个语句",
			goCode: `
				var div = h.Div("Hello World").Class("container")
				var n = div
			`,
			expectError: false,
			expectHTML:  `<div class="container">Hello World</div>`,
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

				// 验证HTML内容
				if !strings.Contains(respBody.HTML, tc.expectHTML) {
					t.Errorf("HTML输出不包含期望的内容\n期望: %s\n实际: %s", tc.expectHTML, respBody.HTML)
				}

				fmt.Printf("测试用例 '%s' 的HTML输出:\n%s\n\n", tc.name, respBody.HTML)
			}
		})
	}
}

// 测试HTML和Go之间的双向转换
func TestBidirectionalConversion(t *testing.T) {
	testCases := []struct {
		name          string
		html          string
		packagePrefix string
	}{
		{
			name:          "简单元素",
			html:          `<div>Hello World</div>`,
			packagePrefix: "h",
		},
		{
			name:          "带类的元素",
			html:          `<div class="container">Hello World</div>`,
			packagePrefix: "h",
		},
		{
			name:          "嵌套元素",
			html:          `<div class="container"><h1>Title</h1><p>Content</p></div>`,
			packagePrefix: "h",
		},
		{
			name:          "表单元素",
			html:          `<form><input type="text" placeholder="Enter text" required></form>`,
			packagePrefix: "h",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 第一步：HTML到Go
			htmlReader := strings.NewReader(tc.html)
			goCode := parse.GenerateHTMLGo(tc.packagePrefix, false, htmlReader)
			goCode = removePackageDeclaration(goCode)

			fmt.Printf("HTML到Go转换结果:\n%s\n\n", goCode)

			// 第二步：Go到HTML
			html, err := convertGoToHTML(goCode)
			if err != nil {
				t.Fatalf("Go到HTML转换失败: %v", err)
			}

			fmt.Printf("Go到HTML转换结果:\n%s\n\n", html)

			// 第三步：HTML到Go（再次）
			htmlReader = strings.NewReader(html)
			goCode2 := parse.GenerateHTMLGo(tc.packagePrefix, false, htmlReader)
			goCode2 = removePackageDeclaration(goCode2)

			fmt.Printf("再次HTML到Go转换结果:\n%s\n\n", goCode2)

			// 验证：最终的Go代码应该与原始Go代码在功能上等效
			// 注意：由于格式化和属性顺序的差异，我们不能直接比较字符串
			// 这里我们只是简单地检查一些关键部分
			if !strings.Contains(goCode2, tc.packagePrefix) {
				t.Errorf("最终Go代码中没有包含指定的包前缀: %s", tc.packagePrefix)
			}

			// 检查主要元素是否存在
			if strings.Contains(tc.html, "<div") && !strings.Contains(goCode2, "Div(") {
				t.Errorf("最终Go代码中缺少Div元素")
			}
			if strings.Contains(tc.html, "<h1") && !strings.Contains(goCode2, "H1(") {
				t.Errorf("最终Go代码中缺少H1元素")
			}
			if strings.Contains(tc.html, "<p") && !strings.Contains(goCode2, "P(") {
				t.Errorf("最终Go代码中缺少P元素")
			}
			if strings.Contains(tc.html, "<form") && !strings.Contains(goCode2, "Form(") {
				t.Errorf("最终Go代码中缺少Form元素")
			}
			if strings.Contains(tc.html, "<input") && !strings.Contains(goCode2, "Input(") {
				t.Errorf("最终Go代码中缺少Input元素")
			}
		})
	}
}
