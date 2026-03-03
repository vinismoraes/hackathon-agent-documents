package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

//go:embed index.html
var content embed.FS

const aorTarget = "http://localhost:8080"
const mockTarget = "http://localhost:8092"

func buildJWT() string {
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	now := time.Now().Unix()
	payload := map[string]any{
		"sub":                 "demo-user",
		"iat":                 now,
		"exp":                 now + 3600,
		"iss":                 "https://id.league.com",
		"https://el/tenant_id": "league",
		"https://el/typ":       "user",
		"https://el/user_id":   "demo-user",
	}

	hJSON, _ := json.Marshal(header)
	pJSON, _ := json.Marshal(payload)

	b64 := func(b []byte) string {
		return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
	}

	signingInput := b64(hJSON) + "." + b64(pJSON)
	mac := hmac.New(sha256.New, []byte("dev-secret"))
	mac.Write([]byte(signingInput))
	sig := b64(mac.Sum(nil))

	return signingInput + "." + sig
}

func main() {
	aorURL, _ := url.Parse(aorTarget)
	proxy := httputil.NewSingleHostReverseProxy(aorURL)
	proxy.FlushInterval = -1 // flush immediately for SSE streaming

	mux := http.NewServeMux()

	// Serve a valid JWT to the browser
	mux.HandleFunc("/api/jwt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"token":"%s"}`, buildJWT())
	})

	// Proxy all /v1/* requests to AOR with streaming support
	mux.HandleFunc("/v1/", func(w http.ResponseWriter, r *http.Request) {
		r.Host = aorURL.Host
		// Override User-Agent so AOR accepts X-Skip-Guardrails
		r.Header.Set("User-Agent", "Local-CLI")
		r.Header.Set("X-Skip-Guardrails", "true")
		w.Header().Set("X-Accel-Buffering", "no")
		w.Header().Set("Cache-Control", "no-cache")
		proxy.ServeHTTP(w, r)
	})

	// Proxy healthz
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(aorTarget + "/healthz")
		if err != nil {
			http.Error(w, "AOR unreachable", 502)
			return
		}
		defer resp.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	// Proxy document downloads to the mock server
	mux.HandleFunc("/doc-content/", func(w http.ResponseWriter, r *http.Request) {
		docID := r.URL.Path[len("/doc-content/"):]
		if docID == "" {
			http.Error(w, "missing document ID", 400)
			return
		}
		url := mockTarget + "/v1/documents/content?documentId=" + docID + "&documentType=document"
		resp, err := http.Get(url)
		if err != nil {
			http.Error(w, "mock server unreachable", 502)
			return
		}
		defer resp.Body.Close()
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s.txt"`, docID))
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	mux.Handle("/", http.FileServer(http.FS(content)))

	port := 8093
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("  Chat UI:  http://localhost:%d\n", port)
	fmt.Printf("  AOR:      %s\n", aorTarget)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}
