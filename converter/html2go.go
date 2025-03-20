package converter

import (
	"fmt"
	"strings"

	"github.com/zhangshanwen/html2go/parse"
)

// HTMLToGoConverter handles converting HTML to Go code
type HTMLToGoConverter struct {
	PackagePrefix string
}

// NewHTMLToGoConverter creates a new HTMLToGoConverter
func NewHTMLToGoConverter(packagePrefix string) *HTMLToGoConverter {
	// If packagePrefix is empty, use default "h"
	if packagePrefix == "" {
		packagePrefix = "h"
	}

	return &HTMLToGoConverter{
		PackagePrefix: packagePrefix,
	}
}

// Convert converts HTML to Go code
func (c *HTMLToGoConverter) Convert(html string) (string, error) {
	if strings.TrimSpace(html) == "" {
		return "", fmt.Errorf("HTML content cannot be empty")
	}

	// Create HTML reader
	htmlReader := strings.NewReader(html)

	// Use html2go/parse for conversion
	goCode := parse.GenerateHTMLGo(c.PackagePrefix, false, htmlReader)

	// Check conversion result
	if goCode == "" {
		return "", fmt.Errorf("HTML conversion failed: generated Go code is empty")
	}

	// Always remove package declaration
	goCode = removePackageDeclaration(goCode)

	return goCode, nil
}

// RemovePrefix removes the package prefix from the Go code
func (c *HTMLToGoConverter) RemovePrefix(goCode string) string {
	return strings.ReplaceAll(goCode, c.PackagePrefix+".", "")
}

// ReplacePrefix replaces the current package prefix with a new one
func (c *HTMLToGoConverter) ReplacePrefix(goCode string, newPrefix string) string {
	return strings.ReplaceAll(goCode, c.PackagePrefix+".", newPrefix+".")
}

// removePackageDeclaration removes package declaration, var n declaration, and Body() wrapper
func removePackageDeclaration(code string) string {
	// Find the first var declaration
	varIndex := strings.Index(code, "var ")
	if varIndex == -1 {
		return code
	}

	// Extract from var onwards
	codeWithoutPackage := strings.TrimSpace(code[varIndex:])

	// Remove var n = prefix and trailing semicolon
	if strings.HasPrefix(codeWithoutPackage, "var n = ") {
		codeWithoutVar := strings.TrimPrefix(codeWithoutPackage, "var n = ")
		// Remove trailing semicolon if present
		if strings.HasSuffix(codeWithoutVar, ";") {
			codeWithoutVar = codeWithoutVar[:len(codeWithoutVar)-1]
		}

		// Remove Body() wrapper
		if strings.HasPrefix(codeWithoutVar, "Body(") && strings.HasSuffix(codeWithoutVar, ")") {
			return codeWithoutVar[5 : len(codeWithoutVar)-1]
		}

		return codeWithoutVar
	}

	return codeWithoutPackage
}
