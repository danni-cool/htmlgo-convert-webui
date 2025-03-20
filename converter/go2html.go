package converter

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"tailwind-converter/parser_utils"
)

// GoToHTMLConverter handles converting Go code to HTML
type GoToHTMLConverter struct {
	Parser       *parser_utils.GoToHTMLParser
	ErrorHandler *parser_utils.ErrorHandler
}

// NewGoToHTMLConverter creates a new GoToHTMLConverter
func NewGoToHTMLConverter() *GoToHTMLConverter {
	return &GoToHTMLConverter{
		Parser:       parser_utils.NewGoToHTMLParser(),
		ErrorHandler: parser_utils.NewErrorHandler(),
	}
}

// Convert converts Go code to HTML using two methods:
// 1. Static parsing with regexes for simple cases
// 2. Dynamic execution in a sandbox for complex cases
func (c *GoToHTMLConverter) Convert(goCode string) (string, error) {
	// Check for empty code
	if strings.TrimSpace(goCode) == "" {
		return "<!-- Go code cannot be empty -->", fmt.Errorf("Go code cannot be empty")
	}

	// Check for simple syntax errors first before trying to parse
	if strings.Contains(goCode, "var n = if") ||
		strings.Contains(goCode, "n := if") ||
		strings.Contains(goCode, "= if true") ||
		strings.Contains(goCode, "= if false") {
		errMsg := "syntax error: unexpected if, expected expression"
		return c.ErrorHandler.GetFriendlyErrorMessage(errMsg), fmt.Errorf(errMsg)
	}

	// Try to parse with the static parser first
	html, err := c.Parser.ParseGoCode(goCode)
	if err == nil && !strings.Contains(html, "<!--") {
		// Static parsing succeeded without error comments
		return html, nil
	}

	// If static parsing fails or produces error comments, try dynamic execution
	return c.executeDynamically(goCode)
}

// executeDynamically runs the Go code in a sandbox and returns the HTML output
func (c *GoToHTMLConverter) executeDynamically(goCode string) (string, error) {
	// Check and replace package prefixes
	// If code uses h. as package prefix, replace with htmlgo.
	if strings.Contains(goCode, "h.") && !strings.Contains(goCode, "htmlgo.") {
		goCode = strings.ReplaceAll(goCode, "h.", "htmlgo.")
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "go2html")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create temporary Go file
	tempFile := filepath.Join(tempDir, "main.go")

	// Prepare complete Go code
	completeGoCode := fmt.Sprintf(`package main

import (
	"fmt"
	"strings"

	"github.com/theplant/htmlgo"
	h "github.com/theplant/htmlgo"
)

func main() {
	// User-provided code
	%s
	
	// Output HTML
	if n != nil {
		html := htmlgo.MustString(n, nil)
		// Beautify HTML output
		html = strings.ReplaceAll(html, "><", ">\n<")
		fmt.Println(html)
	} else {
		fmt.Println("<!-- Warning: No HTML output generated. Please check if your code correctly defines the variable 'n' -->")
	}
}
`, goCode)

	// Write to temporary file
	err = os.WriteFile(tempFile, []byte(completeGoCode), 0o644)
	if err != nil {
		return "", fmt.Errorf("failed to write temp file: %v", err)
	}

	// Execute Go code
	cmd := exec.Command("go", "run", tempFile)
	output, err := cmd.CombinedOutput()

	// Process the execution result
	result := string(output)

	// Handle execution errors
	if err != nil {
		// Try to extract a more useful error message
		errorMsg := c.extractErrorFromOutput(result)
		if errorMsg == "" {
			errorMsg = err.Error()
		}

		// Get a friendly error message
		return c.ErrorHandler.GetFriendlyErrorMessage(errorMsg), err
	}

	// Trim result and check if empty
	resultTrimmed := strings.TrimSpace(result)
	if resultTrimmed == "" {
		return "<!-- Warning: No HTML output generated. Please check if your code correctly defines the variable 'n' -->", nil
	}

	return result, nil
}

// extractErrorFromOutput tries to find the error message in the compiler/runtime output
func (c *GoToHTMLConverter) extractErrorFromOutput(output string) string {
	// Find error lines in Go compiler/runtime output
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "main.go:") && strings.Contains(line, ": ") {
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}
