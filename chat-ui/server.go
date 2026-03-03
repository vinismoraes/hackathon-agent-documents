package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
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

	// Proxy messaging attachment content (mock data, no real backend)
	mux.HandleFunc("/msg-content/", func(w http.ResponseWriter, r *http.Request) {
		docID := r.URL.Path[len("/msg-content/"):]
		if docID == "" {
			http.Error(w, "missing attachment ID", 400)
			return
		}
		mockContent := map[string]string{
			"attach-001": "PRE-VISIT QUESTIONNAIRE\n\nPatient: Demo User\nAppointment: March 15, 2026 at 10:00 AM\nProvider: Dr. Chen\n\n1. Current medications: _______________\n2. Known allergies: _______________\n3. Recent surgeries or hospitalizations: _______________\n4. Primary reason for visit: Annual Checkup\n\nPlease bring this completed form to your appointment.",
			"attach-002": "SUMMARY OF BENEFITS — Gold PPO Plan 2026\n\nMember: Demo User\nPlan Year: January 1 – December 31, 2026\n\nPhysiotherapy Coverage:\n  • Sessions per year: 20\n  • Coverage: 80% after deductible\n  • Deductible: $500 individual / $1,000 family\n  • In-network copay: $30 per session\n\nPreventive Care: Covered 100% in-network\nEmergency Services: $250 copay, then 80%\nPrescription Drugs (generic): $10 copay",
			"attach-003": "PRESCRIPTION REFILL RECEIPT\n\nPharmacy: HealthFirst Pharmacy\nDate: January 15, 2026\nPatient: Demo User\n\nMedication: Lisinopril 10mg\nQuantity: 90 tablets (90-day supply)\nRefill #: 3 of 5\n\nCost Summary:\n  Retail price: $45.00\n  Plan pays: $35.00\n  Your copay: $10.00\n\nNext refill available: April 15, 2026",
			"attach-004": "MEDICATION GUIDE: LISINOPRIL\n\nGeneric Name: Lisinopril\nBrand Names: Prinivil, Zestril\nDrug Class: ACE Inhibitor\n\nWhat is Lisinopril used for?\n  • High blood pressure (hypertension)\n  • Heart failure\n  • Post-heart attack recovery\n\nImportant Safety Information:\n  • Take once daily, with or without food\n  • Avoid potassium supplements unless directed\n  • Report dizziness, persistent cough, or swelling\n  • Do not use if pregnant\n\nCommon Side Effects:\n  Headache, dizziness, cough, fatigue\n\nContact your healthcare provider with questions.",
		}
		text, ok := mockContent[docID]
		if !ok {
			text = fmt.Sprintf("Mock content for messaging attachment %s.\nThis is placeholder content for the hackathon demo.", docID)
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s.txt"`, docID))
		fmt.Fprint(w, text)
	})

	// Temporary file store for upload flow.
	// The chat UI uploads the file here, gets a fileId. The MCP tool then
	// fetches the file bytes by fileId to forward to the extension backend.
	var fileStore sync.Map
	const maxUploadSize = 25 * 1024 * 1024

	mux.HandleFunc("/api/upload-temp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			http.Error(w, "file too large", 413)
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing file", 400)
			return
		}
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "read error", 500)
			return
		}
		idBytes := make([]byte, 16)
		rand.Read(idBytes)
		fileID := hex.EncodeToString(idBytes)
		fileStore.Store(fileID, data)
		log.Printf("[UPLOAD] Stored temp file %s (%d bytes)", fileID, len(data))

		// Auto-expire after 5 minutes
		go func() {
			time.Sleep(5 * time.Minute)
			fileStore.Delete(fileID)
			log.Printf("[UPLOAD] Expired temp file %s", fileID)
		}()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"fileId": fileID})
	})

	mux.HandleFunc("/api/upload-temp/", func(w http.ResponseWriter, r *http.Request) {
		fileID := strings.TrimPrefix(r.URL.Path, "/api/upload-temp/")
		if fileID == "" {
			http.Error(w, "missing fileId", 400)
			return
		}
		data, ok := fileStore.Load(fileID)
		if !ok {
			http.Error(w, "file not found or expired", 404)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(data.([]byte))
	})

	mux.Handle("/", http.FileServer(http.FS(content)))

	port := 8093
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("  Chat UI:  http://localhost:%d\n", port)
	fmt.Printf("  AOR:      %s\n", aorTarget)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}
