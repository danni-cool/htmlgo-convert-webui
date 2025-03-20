package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Static file server
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// API endpoint for code conversion
	http.HandleFunc("/convert", main.HandleConvert)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	fmt.Printf("Server started at http://localhost:%s\n", port)
	log.Printf("Server started at http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
