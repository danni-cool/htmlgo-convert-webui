package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

// RunGoCode executes Go code in a sandbox
func RunGoCode(goCode string, includePackage bool) (string, error) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "go-sandbox")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create temporary Go file
	tempFile := filepath.Join(tempDir, "main.go")

	// Write the Go code to the file
	codeToWrite := goCode
	if includePackage {
		codeToWrite = fmt.Sprintf("package main\n\n%s", goCode)
	}

	if err := os.WriteFile(tempFile, []byte(codeToWrite), 0o644); err != nil {
		return "", fmt.Errorf("failed to write temp file: %v", err)
	}

	// Execute the Go code
	cmd := exec.Command("go", "run", tempFile)
	output, err := cmd.CombinedOutput()

	return string(output), err
}

// SetupServer configures and starts the HTTP server
func SetupServer() {
	// Static file server
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// Conversion API endpoint
	http.HandleFunc("/convert", handleConvert)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Server started at http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
