
package main

import (
    "fmt"
    "net/http"
    "log"
    "path/filepath"
    "os"
)

func main() {
    // Get current working directory
    wd, err := os.Getwd()
    if err != nil {
        log.Fatal("Failed to get working directory:", err)
    }
    
    // Set the web directory path
    webDir := filepath.Join(wd, "web")
    
    // Create a file server that serves files from the web directory
    fs := http.FileServer(http.Dir(webDir))
    
    // Handle all requests
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // Set CORS headers
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        
        // Handle preflight requests
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        // Log the request
        log.Printf("Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
        
        // Serve the file
        fs.ServeHTTP(w, r)
    })
    
    fmt.Println("=== LSCC Dashboard Server ===")
    fmt.Println("Server starting on port 5000")
    fmt.Printf("Web directory: %s\n", webDir)
    fmt.Println("Dashboard URL: http://0.0.0.0:5000")
    fmt.Println("Dashboard URL: http://localhost:5000")
    fmt.Println("==============================")
    
    log.Fatal(http.ListenAndServe("0.0.0.0:5000", nil))
}
