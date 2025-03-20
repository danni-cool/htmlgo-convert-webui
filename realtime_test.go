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
)

// 测试双向实时转换交互模式
func TestRealtimeConversion(t *testing.T) {
	testCases := []struct {
		name          string
		initialHTML   string
		packagePrefix string
	}{
		{
			name:          "简单元素",
			initialHTML:   `<div>Hello World</div>`,
			packagePrefix: "h",
		},
		{
			name:          "带类的元素",
			initialHTML:   `<div class="container">Hello World</div>`,
			packagePrefix: "h",
		},
		{
			name:          "嵌套元素",
			initialHTML:   `<div class="container"><h1>Title</h1><p>Content</p></div>`,
			packagePrefix: "h",
		},
	}

	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(handleConvert))
	defer server.Close()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 模拟用户在HTML编辑器中输入HTML
			htmlInput := tc.initialHTML
			fmt.Printf("初始HTML输入:\n%s\n\n", htmlInput)

			// 第一步：HTML到Go转换（模拟HTML编辑器内容变化触发的转换）
			goCode, err := testHTMLToGo(server.URL, htmlInput, tc.packagePrefix)
			if err != nil {
				t.Fatalf("HTML到Go转换失败: %v", err)
			}
			fmt.Printf("转换后的Go代码:\n%s\n\n", goCode)

			// 第二步：模拟用户修改Go代码（在Go编辑器中添加注释）
			// 确保Go代码中包含变量n的定义
			modifiedGoCode := "var n = " + goCode + "\n// 这是用户添加的注释"
			fmt.Printf("用户修改后的Go代码:\n%s\n\n", modifiedGoCode)

			// 第三步：Go到HTML转换（模拟Go编辑器内容变化触发的转换）
			html, err := testGoToHTML(server.URL, modifiedGoCode)
			if err != nil {
				t.Fatalf("Go到HTML转换失败: %v", err)
			}
			fmt.Printf("Go代码转换回HTML:\n%s\n\n", html)

			// 第四步：模拟用户修改HTML（在HTML编辑器中添加注释）
			modifiedHTML := html + "\n<!-- 这是用户添加的HTML注释 -->"
			fmt.Printf("用户修改后的HTML:\n%s\n\n", modifiedHTML)

			// 第五步：再次HTML到Go转换（模拟HTML编辑器内容变化触发的转换）
			finalGoCode, err := testHTMLToGo(server.URL, modifiedHTML, tc.packagePrefix)
			if err != nil {
				t.Fatalf("最终HTML到Go转换失败: %v", err)
			}
			fmt.Printf("最终Go代码:\n%s\n\n", finalGoCode)

			// 验证转换结果
			// 1. 检查初始HTML到Go的转换是否成功
			if !strings.Contains(goCode, "Div(") && strings.Contains(tc.initialHTML, "<div") {
				t.Errorf("初始Go代码中缺少Div元素")
			}

			// 2. 检查Go到HTML的转换是否保留了原始HTML的结构
			if !strings.Contains(html, "<div") && strings.Contains(tc.initialHTML, "<div") {
				t.Errorf("转换回的HTML中缺少div元素")
			}

			// 3. 检查最终Go代码是否包含了用户的修改
			if strings.Contains(finalGoCode, "// 这是用户添加的注释") {
				t.Errorf("最终Go代码不应该包含用户添加的注释，因为它是从修改后的HTML生成的")
			}
		})
	}
}

// 测试用的HTML到Go转换请求
func testHTMLToGo(serverURL, html, packagePrefix string) (string, error) {
	reqBody := ConvertRequest{
		HTML:          html,
		PackagePrefix: packagePrefix,
		Direction:     "html2go",
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("请求JSON编码失败: %v", err)
	}

	resp, err := http.Post(serverURL+"/convert", "application/json", bytes.NewBuffer(reqJSON))
	if err != nil {
		return "", fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("服务器返回错误: %s - %s", resp.Status, string(body))
	}

	var result ConvertResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("响应解码失败: %v", err)
	}

	return result.Code, nil
}

// 测试用的Go到HTML转换请求
func testGoToHTML(serverURL, goCode string) (string, error) {
	reqBody := ConvertRequest{
		GoCode:    goCode,
		Direction: "go2html",
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("请求JSON编码失败: %v", err)
	}

	resp, err := http.Post(serverURL+"/convert", "application/json", bytes.NewBuffer(reqJSON))
	if err != nil {
		return "", fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("服务器返回错误: %s - %s", resp.Status, string(body))
	}

	var result ConvertResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("响应解码失败: %v", err)
	}

	return result.HTML, nil
}
