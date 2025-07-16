
package main

import (
    "fmt"
    "net/http"
    "log"
)

func main() {
    // Serve static files from current directory
    fs := http.FileServer(http.Dir("."))
    http.Handle("/", fs)
    
    fmt.Println("Dashboard server starting on port 5000")
    fmt.Println("Access the dashboard at: http://0.0.0.0:5000")
    
    log.Fatal(http.ListenAndServe("0.0.0.0:5000", nil))
}
