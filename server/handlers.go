package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"tailwind-converter/converter"
)

// Request and response structures
type ConvertRequest struct {
	HTML          string `json:"html"`
	GoCode        string `json:"goCode"`
	PackagePrefix string `json:"packagePrefix"`
	Direction     string `json:"direction"` // "html2go" or "go2html"
}

type ConvertResponse struct {
	Code  string `json:"code"`
	HTML  string `json:"html"`
	Error string `json:"error,omitempty"`
}

// HandleConvert handles requests to convert between HTML and Go code
func HandleConvert(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		sendJSONError(w, http.StatusMethodNotAllowed, "Only POST method is supported", "request_error")
		return
	}

	// Decode request body
	var req ConvertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "Failed to parse request: "+err.Error(), "request_error")
		return
	}

	// Prepare response
	resp := ConvertResponse{}

	// Handle based on conversion direction
	switch req.Direction {
	case "go2html":
		// Go to HTML conversion
		if strings.TrimSpace(req.GoCode) == "" {
			sendJSONError(w, http.StatusBadRequest, "Go code cannot be empty", "go_error")
			return
		}

		// Create converter and convert
		converter := converter.NewGoToHTMLConverter()
		html, err := converter.Convert(req.GoCode)

		// Always return HTML, even if there was an error (it will be a user-friendly error message)
		resp.HTML = html
		if err != nil {
			resp.Error = err.Error()
			log.Printf("Go to HTML conversion error: %v", err)
		}

	default:
		// HTML to Go conversion (default direction)
		if strings.TrimSpace(req.HTML) == "" {
			sendJSONError(w, http.StatusBadRequest, "HTML content cannot be empty", "html_error")
			return
		}

		// Create converter
		converter := converter.NewHTMLToGoConverter(req.PackagePrefix)

		// Handle possible panics from html2go parser
		defer func() {
			if r := recover(); r != nil {
				errorMsg := fmt.Sprintf("Error during conversion: %v", r)
				sendJSONError(w, http.StatusInternalServerError, errorMsg, "html_error")
			}
		}()

		// Convert HTML to Go
		goCode, err := converter.Convert(req.HTML)
		if err != nil {
			sendJSONError(w, http.StatusInternalServerError, err.Error(), "html_error")
			return
		}

		// Process prefix if needed
		if req.PackagePrefix == "" {
			// Remove prefix if user deleted it
			goCode = converter.RemovePrefix(goCode)
		} else if req.PackagePrefix != converter.PackagePrefix {
			// Replace prefix if user specified a different one
			goCode = converter.ReplacePrefix(goCode, req.PackagePrefix)
		}

		resp.Code = goCode
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// sendJSONError sends a JSON error response
func sendJSONError(w http.ResponseWriter, status int, message string, errorType string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errorResp := map[string]string{
		"error": message,
		"type":  errorType,
	}

	json.NewEncoder(w).Encode(errorResp)
}
