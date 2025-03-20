package parser_utils

import (
	"fmt"
	"regexp"
	"strings"
)

// GoToHTMLParser handles parsing Go code and converting it to HTML
type GoToHTMLParser struct {
	// Mapping from tag name to tag information
	TagMappings map[string]TagMapping
	// Package prefixes to recognize (e.g., "h", "htmlgo")
	PackagePrefixes []string
	// Map from package prefix to actual package name
	PrefixToPackage map[string]string
}

// NewGoToHTMLParser creates a new GoToHTMLParser with the given mappings
func NewGoToHTMLParser() *GoToHTMLParser {
	mappings := GetAllTagMappings()

	// Extract all package prefixes
	prefixMap := make(map[string]string)
	var prefixes []string

	for _, mapping := range mappings {
		if _, exists := prefixMap[mapping.GoPackage]; !exists {
			prefixMap[mapping.GoPackage] = mapping.GoPackage
			prefixes = append(prefixes, mapping.GoPackage)
		}
	}

	// Add common package prefixes
	for _, prefix := range []string{"h", "htmlgo"} {
		if _, exists := prefixMap[prefix]; !exists {
			prefixMap[prefix] = "htmlgo"
			prefixes = append(prefixes, prefix)
		}
	}

	// Add vuetify prefixes
	for _, prefix := range []string{"v", "vuetify"} {
		if _, exists := prefixMap[prefix]; !exists {
			prefixMap[prefix] = "vuetify"
			prefixes = append(prefixes, prefix)
		}
	}

	// Add vuetifyx prefixes
	for _, prefix := range []string{"vx", "vuetifyx"} {
		if _, exists := prefixMap[prefix]; !exists {
			prefixMap[prefix] = "vuetifyx"
			prefixes = append(prefixes, prefix)
		}
	}

	return &GoToHTMLParser{
		TagMappings:     mappings,
		PackagePrefixes: prefixes,
		PrefixToPackage: prefixMap,
	}
}

// ParseGoCode parses Go code and returns HTML representation
func (p *GoToHTMLParser) ParseGoCode(goCode string) (string, error) {
	// Preprocess the code to handle common issues
	processedCode := p.preprocessGoCode(goCode)

	// Find the main expression (n or other variable)
	mainExpr, err := p.findMainExpression(processedCode)
	if err != nil {
		return "", err
	}

	// Now parse the HTML node from the main expression
	html, err := p.parseNodeExpression(mainExpr)
	if err != nil {
		return fmt.Sprintf("<!-- Error parsing Go expression: %s -->", err), err
	}

	// Beautify the HTML
	beautifiedHTML := p.beautifyHTML(html)

	return beautifiedHTML, nil
}

// preprocessGoCode handles common issues in Go code
func (p *GoToHTMLParser) preprocessGoCode(goCode string) string {
	// Remove package declaration if present
	packagePattern := regexp.MustCompile(`(?m)^package\s+\w+\s*`)
	goCode = packagePattern.ReplaceAllString(goCode, "")

	// Remove imports if present
	importBlockPattern := regexp.MustCompile(`(?s)import\s*\((.*?)\)`)
	goCode = importBlockPattern.ReplaceAllString(goCode, "")

	importLinePattern := regexp.MustCompile(`(?m)^import\s+.*$`)
	goCode = importLinePattern.ReplaceAllString(goCode, "")

	// Check if code has a direct if expression error
	directIfPattern := regexp.MustCompile(`(var\s+\w+\s*=|:=)\s*if\b`)
	if directIfPattern.MatchString(goCode) {
		return "ERROR: syntax error: unexpected if, expected expression"
	}

	return goCode
}

// findMainExpression finds the main expression in Go code (n or other variable)
func (p *GoToHTMLParser) findMainExpression(goCode string) (string, error) {
	// Check for direct error message
	if strings.HasPrefix(goCode, "ERROR:") {
		return "", fmt.Errorf(strings.TrimPrefix(goCode, "ERROR:"))
	}

	// First, look for 'var n = ...' or 'n := ...'
	reVarN := regexp.MustCompile(`(?m)var\s+n\s*=\s*(.+?)(?:;|\n|$)`)
	reNColonEqual := regexp.MustCompile(`(?m)n\s*:=\s*(.+?)(?:;|\n|$)`)

	if matches := reVarN.FindStringSubmatch(goCode); len(matches) > 1 {
		return strings.TrimSpace(matches[1]), nil
	}

	if matches := reNColonEqual.FindStringSubmatch(goCode); len(matches) > 1 {
		return strings.TrimSpace(matches[1]), nil
	}

	// If we can't find n, look for any tag expression that might be the node
	for _, prefix := range p.PackagePrefixes {
		pattern := fmt.Sprintf(`(?m)(\b%s\.[A-Z][a-zA-Z0-9]*\(\).*)`, regexp.QuoteMeta(prefix))
		re := regexp.MustCompile(pattern)

		if matches := re.FindStringSubmatch(goCode); len(matches) > 1 {
			return strings.TrimSpace(matches[1]), nil
		}
	}

	return "", fmt.Errorf("could not find main node expression in Go code")
}

// parseNodeExpression parses a Go expression into an HTML node
func (p *GoToHTMLParser) parseNodeExpression(expr string) (string, error) {
	// Find the starting tag and attributes
	tag, restExpr, err := p.findTag(expr)
	if err != nil {
		return "", err
	}

	// Find mapping for this tag
	tagMapping, exists := p.findTagMappingByFunc(tag)
	if !exists {
		return "", fmt.Errorf("unknown tag function: %s", tag)
	}

	// Start building HTML
	html := "<" + tagMapping.TagName

	// Parse chained attributes
	attributes, content, err := p.parseAttributes(restExpr, tagMapping)
	if err != nil {
		return "", err
	}

	// Add attributes to HTML
	for attrName, attrValue := range attributes {
		html += fmt.Sprintf(` %s="%s"`, attrName, attrValue)
	}

	// Close the tag
	if tagMapping.SelfClose {
		html += " />"
		return html, nil
	}

	html += ">"

	// Add content if any
	if content != "" {
		html += content
	}

	// Close tag
	html += "</" + tagMapping.TagName + ">"

	return html, nil
}

// findTag extracts the tag name from a Go expression
func (p *GoToHTMLParser) findTag(expr string) (string, string, error) {
	// Look for expressions like: htmlgo.Div() or h.Div() etc.
	tagPattern := regexp.MustCompile(`^(\w+)\.([A-Z][a-zA-Z0-9]*)\(\)(.*)$`)

	matches := tagPattern.FindStringSubmatch(expr)
	if len(matches) < 4 {
		return "", "", fmt.Errorf("invalid tag expression: %s", expr)
	}

	prefix := matches[1]
	tagFunc := matches[2]
	restExpr := matches[3]

	// Check if prefix is in our list of known prefixes
	found := false
	for _, knownPrefix := range p.PackagePrefixes {
		if prefix == knownPrefix {
			found = true
			break
		}
	}

	if !found {
		return "", "", fmt.Errorf("unknown package prefix: %s", prefix)
	}

	return prefix + "." + tagFunc, restExpr, nil
}

// findTagMappingByFunc finds a tag mapping by its Go function name
func (p *GoToHTMLParser) findTagMappingByFunc(fullFuncName string) (TagMapping, bool) {
	// Split into prefix and function name
	parts := strings.Split(fullFuncName, ".")
	if len(parts) != 2 {
		return TagMapping{}, false
	}

	prefix := parts[0]
	funcName := parts[1]

	// Find the actual package for this prefix
	packageName, exists := p.PrefixToPackage[prefix]
	if !exists {
		return TagMapping{}, false
	}

	// Look for a matching tag with the correct function name and package
	for _, mapping := range p.TagMappings {
		if mapping.GoFuncName == funcName && mapping.GoPackage == packageName {
			return mapping, true
		}
	}

	// Special case fallback: try to infer from the function name for standard HTML tags
	// For example, if funcName is "Div", look for a tag mapping for "div"
	tagName := strings.ToLower(funcName)
	mapping, exists := p.TagMappings[tagName]

	return mapping, exists
}

// parseAttributes extracts attributes and content from a chained method call expression
func (p *GoToHTMLParser) parseAttributes(expr string, tagMapping TagMapping) (map[string]string, string, error) {
	attributes := make(map[string]string)
	var content string

	// Special case for empty expression
	if strings.TrimSpace(expr) == "" {
		return attributes, content, nil
	}

	// Iterate through chained method calls like .Class("container").ID("main").Text("Hello")
	attrPattern := regexp.MustCompile(`\.([A-Z][a-zA-Z0-9]*)\((?:"([^"]*)"|\((.*?)\))\)`)
	matches := attrPattern.FindAllStringSubmatch(expr, -1)

	for _, match := range matches {
		methodName := match[1]
		methodValue := match[2]

		// If the second capture group is empty, use the third one (for complex expressions)
		if methodValue == "" && len(match) > 3 {
			methodValue = match[3]
		}

		// Check if this is a special case for content
		if methodName == "Text" || methodName == "InnerText" {
			content = methodValue
			continue
		}

		// Check if this is a special case for children
		if methodName == "Children" {
			// TODO: Implement recursive parsing of child nodes
			content = fmt.Sprintf("<!-- Complex children structure not implemented: %s -->", methodValue)
			continue
		}

		// Find the corresponding HTML attribute for this method
		attrName := ""
		for htmlAttr, goMethod := range tagMapping.Attributes {
			if goMethod == methodName {
				attrName = htmlAttr
				break
			}
		}

		// If no mapping found, use the method name as attribute name in kebab-case
		if attrName == "" {
			attrName = p.camelToKebab(methodName)
		}

		attributes[attrName] = methodValue
	}

	return attributes, content, nil
}

// camelToKebab converts camelCase to kebab-case
func (p *GoToHTMLParser) camelToKebab(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result.WriteRune('-')
			result.WriteRune(r - 'A' + 'a')
		} else {
			result.WriteRune(r)
		}
	}
	return strings.ToLower(result.String())
}

// beautifyHTML adds proper indentation and line breaks to HTML
func (p *GoToHTMLParser) beautifyHTML(html string) string {
	// Simple beautification: just add newlines between tags
	html = strings.ReplaceAll(html, "><", ">\n<")
	return html
}
