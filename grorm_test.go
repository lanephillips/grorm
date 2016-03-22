package grorm

import (
    "fmt"
    "net/http"
    "testing"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func TestServer(t *testing.T) {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}
