package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const port = 8092

// --- Filters ---

var filtersResponse = map[string]any{
	"data": []map[string]any{
		{
			"id":   "filter-statements",
			"type": "document-filter",
			"attributes": map[string]any{
				"filterKey":  "statements",
				"filterName": "Monthly Statements",
				"filterValue": map[string]any{
					"fieldType": "single-select",
					"options": []map[string]any{
						{"optionKey": "year2026", "optionLabel": "2026", "defaultValue": true},
						{"optionKey": "year2025", "optionLabel": "2025", "defaultValue": false},
					},
				},
				"parentFilter": nil,
			},
		},
		{
			"id":   "filter-bills",
			"type": "document-filter",
			"attributes": map[string]any{
				"filterKey":  "bills",
				"filterName": "Premium Bills",
				"filterValue": map[string]any{
					"fieldType": "single-select",
					"options": []map[string]any{
						{"optionKey": "year2026", "optionLabel": "2026", "defaultValue": true},
						{"optionKey": "year2025", "optionLabel": "2025", "defaultValue": false},
					},
				},
				"parentFilter": nil,
			},
		},
		{
			"id":   "filter-plan-docs",
			"type": "document-filter",
			"attributes": map[string]any{
				"filterKey":  "plan_docs",
				"filterName": "Plan Documents",
				"filterValue": map[string]any{
					"fieldType": "single-select",
					"options":   []map[string]any{},
				},
				"parentFilter": nil,
			},
		},
		{
			"id":   "filter-uploads",
			"type": "document-filter",
			"attributes": map[string]any{
				"filterKey":  "uploaded_by_me",
				"filterName": "My Uploads",
				"filterValue": map[string]any{
					"fieldType": "single-select",
					"options":   []map[string]any{},
				},
				"parentFilter": nil,
			},
		},
	},
}

// --- Documents by category ---

var statementDocs = []map[string]any{
	{
		"id":   "doc-stmt-001",
		"type": "document",
		"attributes": map[string]any{
			"title":        "January 2026 Monthly Statement",
			"fileType":     "pdf",
			"createdDate":  "2026-01-31T00:00:00Z",
			"modifiedDate": "2026-01-31T00:00:00Z",
			"actions":      []string{"view", "download"},
		},
	},
	{
		"id":   "doc-stmt-002",
		"type": "document",
		"attributes": map[string]any{
			"title":        "December 2025 Monthly Statement",
			"fileType":     "pdf",
			"createdDate":  "2025-12-31T00:00:00Z",
			"modifiedDate": "2025-12-31T00:00:00Z",
			"actions":      []string{"view", "download"},
		},
	},
	{
		"id":   "doc-stmt-003",
		"type": "document",
		"attributes": map[string]any{
			"title":        "November 2025 Monthly Statement",
			"fileType":     "pdf",
			"createdDate":  "2025-11-30T00:00:00Z",
			"modifiedDate": "2025-11-30T00:00:00Z",
			"actions":      []string{"view", "download"},
		},
	},
}

var billDocs = []map[string]any{
	{
		"id":   "doc-bill-001",
		"type": "document",
		"attributes": map[string]any{
			"title":        "February 2026 Premium Bill",
			"fileType":     "pdf",
			"createdDate":  "2026-02-01T00:00:00Z",
			"modifiedDate": "2026-02-01T00:00:00Z",
			"actions":      []string{"view", "download"},
		},
	},
	{
		"id":   "doc-bill-002",
		"type": "document",
		"attributes": map[string]any{
			"title":        "January 2026 Premium Bill",
			"fileType":     "pdf",
			"createdDate":  "2026-01-01T00:00:00Z",
			"modifiedDate": "2026-01-01T00:00:00Z",
			"actions":      []string{"view", "download"},
		},
	},
}

var planDocs = []map[string]any{
	{
		"id":   "doc-plan-001",
		"type": "document",
		"attributes": map[string]any{
			"title":        "2026 Health Plan Summary of Benefits",
			"fileType":     "pdf",
			"createdDate":  "2026-01-01T00:00:00Z",
			"modifiedDate": "2026-01-01T00:00:00Z",
			"actions":      []string{"view", "download"},
		},
	},
	{
		"id":   "doc-plan-002",
		"type": "document",
		"attributes": map[string]any{
			"title":        "Prescription Drug Formulary 2026",
			"fileType":     "pdf",
			"createdDate":  "2026-01-01T00:00:00Z",
			"modifiedDate": "2026-01-01T00:00:00Z",
			"actions":      []string{"view", "download"},
		},
	},
}

var allDocs = func() []map[string]any {
	var all []map[string]any
	all = append(all, statementDocs...)
	all = append(all, billDocs...)
	all = append(all, planDocs...)
	return all
}()

// --- Document content (the "wow" moment data) ---

var documentContents = map[string]string{
	"doc-stmt-001": `
HEALTH PLAN MONTHLY STATEMENT
==============================
Statement Period: January 1 - January 31, 2026
Member: Jane Doe
Member ID: HM-12345678
Plan: Gold PPO Health Plan

CLAIMS SUMMARY
--------------

1. Office Visit - Preventive Care
   Date of Service: January 15, 2026
   Provider: Dr. Sarah Smith, Family Medicine
   Facility: Lakewood Medical Center
   Billed Amount:    $150.00
   Plan Discount:    -$75.00
   Plan Paid:        -$50.00
   YOUR COPAY:       $25.00

2. Laboratory Services - Blood Work
   Date of Service: January 15, 2026
   Provider: Quest Diagnostics
   Billed Amount:    $285.00
   Plan Discount:    -$142.50
   Plan Paid:        -$142.50
   YOUR COST:        $0.00

3. Prescription - Lisinopril 10mg (30-day supply)
   Date of Service: January 16, 2026
   Pharmacy: CVS Pharmacy #4521
   Retail Price:     $45.00
   Plan Paid:        -$35.00
   YOUR COPAY:       $10.00

MONTHLY TOTALS
--------------
Total Billed:       $480.00
Plan Discounts:     -$217.50
Plan Paid:          -$227.50
YOUR TOTAL:         $35.00

DEDUCTIBLE STATUS
-----------------
Individual Deductible:  $500.00 / $1,500.00 used
Family Deductible:      $500.00 / $3,000.00 used
Out-of-Pocket Maximum:  $535.00 / $6,000.00 used
`,
	"doc-stmt-002": `
HEALTH PLAN MONTHLY STATEMENT
==============================
Statement Period: December 1 - December 31, 2025
Member: Jane Doe
Member ID: HM-12345678
Plan: Gold PPO Health Plan

CLAIMS SUMMARY
--------------

1. Specialist Visit - Dermatology
   Date of Service: December 5, 2025
   Provider: Dr. Michael Chen, Dermatology
   Facility: Downtown Dermatology Associates
   Billed Amount:    $225.00
   Plan Discount:    -$100.00
   Plan Paid:        -$75.00
   YOUR COPAY:       $50.00

MONTHLY TOTALS
--------------
Total Billed:       $225.00
Plan Discounts:     -$100.00
Plan Paid:          -$75.00
YOUR TOTAL:         $50.00
`,
	"doc-bill-001": `
PREMIUM BILL
=============
Billing Period: February 2026
Due Date: February 15, 2026
Member: Jane Doe
Member ID: HM-12345678

Plan: Gold PPO Health Plan

Premium Breakdown:
  Medical:    $425.00
  Dental:     $45.00
  Vision:     $15.00
  -------------------
  TOTAL DUE:  $485.00

Payment Status: UNPAID
`,
	"doc-bill-002": `
PREMIUM BILL
=============
Billing Period: January 2026
Due Date: January 15, 2026
Member: Jane Doe
Member ID: HM-12345678

Plan: Gold PPO Health Plan

Premium Breakdown:
  Medical:    $425.00
  Dental:     $45.00
  Vision:     $15.00
  -------------------
  TOTAL DUE:  $485.00

Payment Status: PAID (January 12, 2026)
`,
	"doc-plan-001": `
2026 HEALTH PLAN SUMMARY OF BENEFITS
=====================================
Plan: Gold PPO Health Plan
Effective: January 1, 2026

KEY BENEFITS:
- Individual Deductible: $1,500
- Family Deductible: $3,000
- Out-of-Pocket Maximum: $6,000 (individual) / $12,000 (family)
- Primary Care Visit Copay: $25
- Specialist Visit Copay: $50
- Urgent Care Copay: $75
- Emergency Room Copay: $250
- Generic Rx Copay: $10
- Brand Rx Copay: $35
- Preventive Care: $0 (fully covered)
`,
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func handleFilters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	log.Printf("[MOCK] GET /v1/document-filters")
	writeJSON(w, http.StatusOK, filtersResponse)
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body map[string]any
	json.NewDecoder(r.Body).Decode(&body)

	filterKey := extractFilterKey(body)
	log.Printf("[MOCK] POST /v1/documents/search (filter: %s)", filterKey)

	var docs []map[string]any
	switch filterKey {
	case "statements":
		docs = statementDocs
	case "bills":
		docs = billDocs
	case "plan_docs":
		docs = planDocs
	default:
		docs = allDocs
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": docs,
		"meta": map[string]any{"totalRecords": len(docs)},
	})
}

func extractFilterKey(body map[string]any) string {
	included, ok := body["included"].([]any)
	if !ok || len(included) == 0 {
		return ""
	}
	for _, inc := range included {
		item, ok := inc.(map[string]any)
		if !ok {
			continue
		}
		attrs, ok := item["attributes"].(map[string]any)
		if !ok {
			continue
		}
		if key, ok := attrs["filterKey"].(string); ok {
			return key
		}
	}
	return ""
}

func handleContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	docID := r.URL.Query().Get("documentId")
	log.Printf("[MOCK] GET /v1/documents/content (documentId: %s)", docID)

	content, exists := documentContents[docID]
	if !exists {
		content = "Document content not available for this document ID."
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.pdf", docID))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(content))
}

func handleDocuments(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case path == "/v1/document-filters":
		handleFilters(w, r)
	case path == "/v1/documents/search":
		handleSearch(w, r)
	case path == "/v1/documents/content":
		handleContent(w, r)
	case path == "/v1/documents" && r.Method == http.MethodPost:
		log.Printf("[MOCK] POST /v1/documents (upload)")
		writeJSON(w, http.StatusAccepted, map[string]any{"data": statementDocs[:1]})
	case path == "/v1/documents" && r.Method == http.MethodDelete:
		log.Printf("[MOCK] DELETE /v1/documents")
		w.WriteHeader(http.StatusNoContent)
	case path == "/v1/documents" && r.Method == http.MethodPatch:
		log.Printf("[MOCK] PATCH /v1/documents")
		writeJSON(w, http.StatusOK, map[string]any{"data": statementDocs[:1]})
	case path == "/" || path == "/health":
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  "ok",
			"service": "hackathon-demo-mock-documents",
		})
	default:
		log.Printf("[MOCK] 404 %s %s", r.Method, path)
		http.NotFound(w, r)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		handleDocuments(w, r)
	})

	addr := fmt.Sprintf(":%d", port)
	log.Printf("=== Hackathon Demo Mock Documents Server ===")
	log.Printf("Listening on %s", addr)
	log.Printf("")
	log.Printf("Endpoints:")
	log.Printf("  GET  /v1/document-filters     → 4 filter categories")
	log.Printf("  POST /v1/documents/search     → documents by filter")
	log.Printf("  GET  /v1/documents/content     → document text content")
	log.Printf("")
	log.Printf("Demo documents:")
	log.Printf("  doc-stmt-001  January 2026 Statement  (copay: $25)")
	log.Printf("  doc-stmt-002  December 2025 Statement (copay: $50)")
	log.Printf("  doc-bill-001  February 2026 Premium Bill ($485)")
	log.Printf("  doc-bill-002  January 2026 Premium Bill ($485, PAID)")
	log.Printf("  doc-plan-001  2026 Summary of Benefits")
	log.Printf("")
	log.Printf("Try: curl http://localhost%s/v1/document-filters | jq", addr)
	log.Printf("Try: curl http://localhost%s/v1/documents/content?documentId=doc-stmt-001", addr)

	banner := strings.Repeat("=", 50)
	log.Printf(banner)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
