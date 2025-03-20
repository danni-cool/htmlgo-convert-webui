package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/zhangshanwen/html2go/parse"
)

// 请求结构体
type ConvertRequest struct {
	HTML          string `json:"html"`
	GoCode        string `json:"goCode"`
	PackagePrefix string `json:"packagePrefix"`
	Direction     string `json:"direction"` // "html2go" 或 "go2html"
}

// 响应结构体
type ConvertResponse struct {
	Code  string `json:"code"`
	HTML  string `json:"html"`
	Error string `json:"error,omitempty"`
}

// 删除代码中的package声明、var n声明和Body()包裹，根据包前缀处理
func removePackageDeclaration(code string) string {
	// 查找第一个var声明的位置
	varIndex := strings.Index(code, "var ")
	if varIndex == -1 {
		return code
	}

	// 截取从var开始的部分
	codeWithoutPackage := strings.TrimSpace(code[varIndex:])

	// 移除var n =前缀和最后的可能存在的分号
	if strings.HasPrefix(codeWithoutPackage, "var n = ") {
		codeWithoutVar := strings.TrimPrefix(codeWithoutPackage, "var n = ")
		// 移除末尾可能的分号
		if strings.HasSuffix(codeWithoutVar, ";") {
			codeWithoutVar = codeWithoutVar[:len(codeWithoutVar)-1]
		}

		// 移除Body()包裹
		if strings.HasPrefix(codeWithoutVar, "Body(") && strings.HasSuffix(codeWithoutVar, ")") {
			return codeWithoutVar[5 : len(codeWithoutVar)-1]
		}

		return codeWithoutVar
	}

	return codeWithoutPackage
}

// getFriendlyErrorMessage returns a more user-friendly error message
func getFriendlyErrorMessage(errorMsg string) string {
	// Common error messages mapped to user-friendly explanations
	errorMap := map[string]string{
		"syntax error: unexpected if, expected expression": "错误: 在Go代码中不能直接使用if表达式作为赋值。请使用函数或变量来存储条件结果。",
		"undefined:":   "错误: 找不到指定的变量或函数。请检查您是否正确导入了所有必要的包，以及变量或函数名是否拼写正确。",
		"cannot use":   "错误: 类型不匹配。请确保您的变量类型与函数期望的类型一致。",
		"syntax error": "语法错误: 您的代码包含Go语法错误。请检查括号、逗号、分号等是否正确。",
	}

	// 检查是否有匹配的友好错误信息
	for key, value := range errorMap {
		if strings.Contains(errorMsg, key) {
			return value + "\n原始错误: " + errorMsg
		}
	}

	// 返回原始错误消息的友好包装
	return "转换过程中出现错误: " + errorMsg
}

// convertGoToHTML 使用临时文件执行Go代码并生成HTML
func convertGoToHTML(goCode string) (string, error) {
	// 检查是否包含非法的直接if表达式
	if strings.Contains(goCode, "var n = if") || strings.Contains(goCode, "n := if") {
		return getFriendlyErrorMessage("syntax error: unexpected if, expected expression"), fmt.Errorf("syntax error: unexpected if, expected expression")
	}

	// 检查并替换包前缀
	// 如果代码使用了h.作为包前缀，替换为htmlgo.
	if strings.Contains(goCode, "h.") && !strings.Contains(goCode, "htmlgo.") {
		goCode = strings.ReplaceAll(goCode, "h.", "htmlgo.")
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "go2html")
	if err != nil {
		return "", fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建临时Go文件
	tempFile := filepath.Join(tempDir, "main.go")

	// 准备完整的Go代码
	completeGoCode := `package main

import (
	"fmt"
	"strings"

	"github.com/theplant/htmlgo"
	h "github.com/theplant/htmlgo"
)

func main() {
	// 用户提供的代码
	` + goCode + `
	
	// 输出HTML
	if n != nil {
		html := htmlgo.MustString(n, nil)
		// 美化HTML输出
		html = strings.ReplaceAll(html, "><", ">\n<")
		fmt.Println(html)
	} else {
		fmt.Println("<!-- 警告: 没有生成HTML输出，请检查您的代码是否正确定义了变量 'n' -->")
	}
}
`
	// 写入临时文件
	err = os.WriteFile(tempFile, []byte(completeGoCode), 0o644)
	if err != nil {
		return "", fmt.Errorf("写入临时文件失败: %v", err)
	}

	// 执行Go代码
	cmd := exec.Command("go", "run", tempFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 尝试提取更有用的错误信息
		errorMsg := string(output)
		if strings.Contains(errorMsg, "main.go:") {
			lines := strings.Split(errorMsg, "\n")
			for _, line := range lines {
				if strings.Contains(line, "main.go:") && strings.Contains(line, ": ") {
					parts := strings.SplitN(line, ": ", 2)
					if len(parts) > 1 {
						errorMsg = parts[1]
						break
					}
				}
			}
		}
		return getFriendlyErrorMessage(errorMsg), nil
	}

	result := string(output)
	if strings.TrimSpace(result) == "" {
		return "<!-- 警告: 没有生成HTML输出，请检查您的代码是否正确定义了变量 'n' -->", nil
	}

	return result, nil
}

// 处理HTML到Go的转换请求
func handleConvert(w http.ResponseWriter, r *http.Request) {
	// 只接受POST请求
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "只支持POST方法",
			"type":  "request_error",
		})
		return
	}

	// 解析请求体
	var req ConvertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "无法解析请求: " + err.Error(),
			"type":  "request_error",
		})
		return
	}

	// 准备响应
	resp := ConvertResponse{}

	// 根据转换方向处理
	if req.Direction == "go2html" {
		// Go代码转HTML
		if strings.TrimSpace(req.GoCode) == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Go代码不能为空",
				"type":  "go_error",
			})
			return
		}

		html, err := convertGoToHTML(req.GoCode)
		if err != nil {
			// 返回详细的错误信息，但不将状态码设为错误，因为我们已经返回了可用的HTML错误消息
			resp.HTML = html
			resp.Error = err.Error()
		} else {
			resp.HTML = html
		}
	} else {
		// HTML转Go代码（默认方向）
		if strings.TrimSpace(req.HTML) == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "HTML内容不能为空",
				"type":  "html_error",
			})
			return
		}

		// 设置包前缀
		packagePrefix := req.PackagePrefix
		// 注意：这里不再设置默认值，而是保留空字符串
		// 如果前端传递了空字符串，我们就使用空字符串

		// 使用html2go/parse进行转换
		htmlReader := strings.NewReader(req.HTML)

		// 使用recover捕获可能的panic
		defer func() {
			if r := recover(); r != nil {
				// 将panic转换为错误响应
				errorMsg := fmt.Sprintf("转换过程中发生错误: %v", r)
				errorResp := map[string]string{
					"error": errorMsg,
					"type":  "html_error",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(errorResp)
			}
		}()

		// 如果packagePrefix为空，则使用默认值"h"进行转换
		// 但在输出前会删除所有前缀
		prefixForConversion := packagePrefix
		if prefixForConversion == "" {
			prefixForConversion = "h"
		}

		goCode := parse.GenerateHTMLGo(prefixForConversion, false, htmlReader)

		// 检查转换结果
		if goCode == "" {
			errorResp := map[string]string{
				"error": "HTML转换失败: 生成的Go代码为空",
				"type":  "html_error",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(errorResp)
			return
		}

		// 始终删除package声明
		goCode = removePackageDeclaration(goCode)

		// 根据用户输入的包前缀处理代码
		if packagePrefix == "" {
			// 如果用户删除了包前缀，则从代码中也删除前缀
			goCode = strings.ReplaceAll(goCode, prefixForConversion+".", "")
		} else if packagePrefix != prefixForConversion {
			// 如果用户指定了非默认的包前缀，则替换默认前缀
			goCode = strings.ReplaceAll(goCode, prefixForConversion+".", packagePrefix+".")
		}

		resp.Code = goCode
	}

	// 返回转换后的代码
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	// 设置HTTP路由
	http.HandleFunc("/convert", handleConvert)

	// 设置静态文件服务
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// 确定端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // 默认端口
	}

	// 启动服务器
	fmt.Printf("服务器启动在端口 http://localhost:%s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
	}
}
