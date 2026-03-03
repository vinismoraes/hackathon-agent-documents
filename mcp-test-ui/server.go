package main

import (
	"embed"
	"fmt"
	"io"
	"log"
	"net/http"
)

//go:embed index.html
var content embed.FS

const mcpTarget = "http://localhost:18006"
const mockTarget = "http://localhost:8092"

func proxyTo(target string, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := target + path
		if r.URL.RawQuery != "" {
			url += "?" + r.URL.RawQuery
		}
		proxyReq, err := http.NewRequest(r.Method, url, r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		for k, vv := range r.Header {
			for _, v := range vv {
				proxyReq.Header.Add(k, v)
			}
		}

		resp, err := http.DefaultClient.Do(proxyReq)
		if err != nil {
			http.Error(w, "upstream unreachable: "+err.Error(), 502)
			return
		}
		defer resp.Body.Close()

		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/mcp", proxyTo(mcpTarget, "/mcp"))
	mux.HandleFunc("/doc-content", proxyTo(mockTarget, "/v1/documents/content"))

	mux.Handle("/", http.FileServer(http.FS(content)))

	port := 8091
	fmt.Printf("MCP Test UI: http://localhost:%d\n", port)
	fmt.Printf("Proxying /mcp to %s/mcp\n", mcpTarget)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}
