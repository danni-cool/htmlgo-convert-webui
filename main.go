package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sunfmin/html2go/parse"
)

// 请求结构体
type ConvertRequest struct {
	HTML          string `json:"html"`
	GoCode        string `json:"goCode"`
	PackagePrefix string `json:"packagePrefix"`
	RemovePackage bool   `json:"removePackage"`
	Direction     string `json:"direction"` // "html2go" 或 "go2html"
}

// 响应结构体
type ConvertResponse struct {
	Code string `json:"code"`
	HTML string `json:"html"`
}

// 删除代码中的package声明
func removePackageDeclaration(code string) string {
	// 查找第一个var声明的位置
	varIndex := strings.Index(code, "var ")
	if varIndex == -1 {
		return code
	}

	// 截取从var开始的部分
	return strings.TrimSpace(code[varIndex:])
}

// 将Go代码转换为HTML
func convertGoToHTML(goCode string) (string, error) {
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

	// 准备Go代码
	fullGoCode := `package main

import (
	"fmt"
	"strings"

	"github.com/theplant/htmlgo"
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
		fmt.Println("<!-- 错误: 变量 'n' 未定义或为nil -->")
	}
}
`
	// 写入临时文件
	err = os.WriteFile(tempFile, []byte(fullGoCode), 0o644)
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
		return "<!-- 编译或执行错误: " + strings.ReplaceAll(errorMsg, "--", "-") + " -->", nil
	}

	result := string(output)
	if strings.TrimSpace(result) == "" {
		return "<!-- 警告: 没有生成HTML输出，请检查您的代码是否正确定义了变量 'n' -->", nil
	}

	return result, nil
}

func main() {
	// 静态文件服务
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// 转换API
	http.HandleFunc("/convert", handleConvert)

	// 获取端口号，优先使用环境变量PORT，如果未设置则默认使用8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 启动服务器
	log.Printf("服务器启动在 http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
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
			// 返回详细的错误信息
			errorResp := map[string]string{
				"error": "转换失败: " + err.Error(),
				"type":  "go_error",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(errorResp)
			return
		}
		resp.HTML = html
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

		// 设置默认包前缀
		packagePrefix := req.PackagePrefix
		if packagePrefix == "" {
			packagePrefix = "h"
		}

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

		goCode := parse.GenerateHTMLGo(packagePrefix, false, htmlReader)

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

		// 如果需要，删除package声明
		if req.RemovePackage {
			goCode = removePackageDeclaration(goCode)
		}

		resp.Code = goCode
	}

	// 返回转换后的代码
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
