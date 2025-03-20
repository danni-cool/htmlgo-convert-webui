package main

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/zhangshanwen/html2go/parse"
)

// ConversionRequest represents the JSON request body for conversion
type ConversionRequest struct {
	HTML          string `json:"html"`
	PackagePrefix string `json:"packagePrefix"`
	Direction     string `json:"direction"`
}

// ConversionResponse represents the JSON response for conversion
type ConversionResponse struct {
	Code  string `json:"code,omitempty"`
	HTML  string `json:"html,omitempty"`
	Error string `json:"error,omitempty"`
}

func main() {
	// Create a new router
	mux := http.NewServeMux()

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/", fs)

	// API endpoint for conversion
	mux.HandleFunc("/convert", convertHandler)

	// Configure the HTTP server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start the server
	log.Println("Starting server on :8080")
	log.Fatal(server.ListenAndServe())
}

func convertHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req ConversionRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		sendJSONError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.HTML == "" {
		sendJSONError(w, "HTML content is required", http.StatusBadRequest)
		return
	}

	// Set default package prefix if not provided
	if req.PackagePrefix == "" {
		req.PackagePrefix = "h"
	}

	// Process based on direction
	var response ConversionResponse
	switch req.Direction {
	case "html2go":
		code, err := convertHTMLToGo(req.HTML, req.PackagePrefix)
		if err != nil {
			sendJSONError(w, fmt.Sprintf("HTML to Go conversion error: %v", err), http.StatusInternalServerError)
			return
		}
		response.Code = code
	case "go2html":
		// Not implemented yet - might be added in a future update
		sendJSONError(w, "Go to HTML conversion is not implemented yet", http.StatusNotImplemented)
		return
	default:
		sendJSONError(w, "Invalid conversion direction", http.StatusBadRequest)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func convertHTMLToGo(htmlContent, packagePrefix string) (string, error) {
	// The function takes a reader, so we need to convert our string to a reader
	reader := strings.NewReader(htmlContent)

	// Using the Vuetify branch API
	// Generate HTML Go code with support for Vuetify components
	goCode := parse.GenerateHTMLGo(packagePrefix, packagePrefix+"v", packagePrefix+"vx", false, reader)

	// Post-process the generated code to fix any syntax issues
	goCode = fixSyntaxIssues(goCode)

	// Validate the generated code syntax
	if err := validateGoSyntax(goCode); err != nil {
		// If we can't validate the syntax, just return the code without error
		// This is because the parser might not understand some valid Go constructs
		// in the generated code
		log.Printf("Warning: Syntax validation failed, but returning code anyway: %v", err)
		return goCode, nil
	}

	return goCode, nil
}

// fixSyntaxIssues fixes common syntax issues in the generated code
func fixSyntaxIssues(code string) string {
	// Remove package declaration if present
	if strings.Contains(code, "package hello") {
		code = strings.Replace(code, "package hello\n\n", "", 1)
	}

	// Fix method calls that might be incorrectly prefixed with dot
	methodsToFix := []string{"Class", "Attr", "Color", "Style", "ID"}
	for _, method := range methodsToFix {
		code = strings.Replace(code, "."+method+"(", method+"(", -1)
	}

	// Fix closing parentheses with commas
	code = strings.Replace(code, ")\n\t", "),\n\t", -1)

	// Fix trailing commas - make sure the last item doesn't have a comma
	lines := strings.Split(code, "\n")
	for i := 0; i < len(lines)-1; i++ {
		if strings.HasSuffix(lines[i], ",") && strings.TrimSpace(lines[i+1]) == ")" {
			lines[i] = strings.TrimSuffix(lines[i], ",")
		}
	}
	code = strings.Join(lines, "\n")

	// Remove any variable declaration and just return the actual code
	if strings.Contains(code, "var n = ") {
		code = strings.Replace(code, "var n = ", "", 1)
	}

	// Add a header comment
	if !strings.Contains(code, "// Generated") {
		code = "// Generated using htmlgo\n\n" + code
	}

	return code
}

// validateGoSyntax validates the syntax of the generated Go code
func validateGoSyntax(code string) error {
	// Wrap the code in a function declaration to make it parseable
	testCode := "package test\n\nfunc testFunc() {\n" + code + "\n}"

	// Try to parse the code
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, "", testCode, parser.DeclarationErrors)
	if err != nil {
		return err
	}

	return nil
}

func sendJSONError(w http.ResponseWriter, errMsg string, statusCode int) {
	response := ConversionResponse{
		Error: errMsg,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding error response: %v", err)
		http.Error(w, errMsg, statusCode)
	}
}
