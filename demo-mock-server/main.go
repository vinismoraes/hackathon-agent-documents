package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
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
	{
		"id":   "doc-stmt-004",
		"type": "document",
		"attributes": map[string]any{
			"title":        "October 2025 Monthly Statement",
			"fileType":     "pdf",
			"createdDate":  "2025-10-31T00:00:00Z",
			"modifiedDate": "2025-10-31T00:00:00Z",
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
	{
		"id":   "doc-bill-003",
		"type": "document",
		"attributes": map[string]any{
			"title":        "December 2025 Premium Bill",
			"fileType":     "pdf",
			"createdDate":  "2025-12-01T00:00:00Z",
			"modifiedDate": "2025-12-01T00:00:00Z",
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
	{
		"id":   "doc-plan-003",
		"type": "document",
		"attributes": map[string]any{
			"title":        "Provider Network Directory 2026",
			"fileType":     "pdf",
			"createdDate":  "2026-01-01T00:00:00Z",
			"modifiedDate": "2026-01-01T00:00:00Z",
			"actions":      []string{"view", "download"},
		},
	},
}

var (
	uploadedDocs []map[string]any
	mu           sync.RWMutex
)

func getAllDocs() []map[string]any {
	mu.RLock()
	defer mu.RUnlock()
	var all []map[string]any
	all = append(all, statementDocs...)
	all = append(all, billDocs...)
	all = append(all, planDocs...)
	all = append(all, uploadedDocs...)
	return all
}

// --- Document content (the "wow" moment data) ---

var documentContents = map[string]string{
	"doc-stmt-001": `
HEALTH PLAN MONTHLY STATEMENT — EXPLANATION OF BENEFITS
=========================================================
Statement Period: January 1 - January 31, 2026
Generated: February 3, 2026
Member: Jane Doe | DOB: 04/15/1985
Member ID: HM-12345678 | Group: GRP-ACME-2026
Plan: Gold PPO Health Plan (In-Network Tier 1)
PCP: Dr. Sarah Smith, MD — Lakewood Medical Center

IMPORTANT NOTICES
-----------------
• Your annual wellness visit is due. Schedule before March 31, 2026 to
  avoid the $150 preventive care gap.
• Prescription benefit changes effective February 1, 2026: Lisinopril
  moves to Tier 1 ($5 copay). Atorvastatin remains Tier 2 ($15 copay).
• You have unused FSA funds of $340.00 expiring December 31, 2026.

CLAIMS DETAIL — 7 CLAIMS PROCESSED
------------------------------------

1. Office Visit — Annual Physical & Preventive Screening
   Date of Service:   January 8, 2026
   Provider:          Dr. Sarah Smith, MD (Family Medicine)
   Facility:          Lakewood Medical Center
   Procedure Codes:   99395 (preventive visit), 36415 (venipuncture)
   Diagnosis:         Z00.00 (General adult medical exam)
   Billed Amount:     $285.00
   Network Discount:  -$127.50
   Plan Paid:         -$157.50
   YOUR COST:         $0.00    ← Covered 100% as preventive care

2. Laboratory Services — Comprehensive Metabolic Panel + Lipid Panel
   Date of Service:   January 8, 2026
   Provider:          Quest Diagnostics (Reference Lab)
   Procedure Codes:   80053 (CMP), 80061 (Lipid panel), 85025 (CBC)
   Diagnosis:         Z00.00, E78.5 (Hyperlipidemia screening)
   Billed Amount:     $412.00
   Network Discount:  -$247.20
   Plan Paid:         -$164.80
   YOUR COST:         $0.00    ← Covered as part of preventive visit

3. Specialist Visit — Cardiology Follow-up
   Date of Service:   January 14, 2026
   Provider:          Dr. Robert Nguyen, MD (Cardiology)
   Facility:          Heart & Vascular Institute
   Procedure Codes:   99214 (office visit, moderate complexity)
   Diagnosis:         I10 (Essential hypertension), E78.5 (Hyperlipidemia)
   Billed Amount:     $350.00
   Network Discount:  -$140.00
   Plan Paid:         -$170.00
   YOUR COPAY:        $40.00   ← Specialist copay

4. Diagnostic Imaging — Echocardiogram
   Date of Service:   January 14, 2026
   Provider:          Heart & Vascular Institute (Imaging Dept.)
   Procedure Codes:   93306 (Echocardiography, complete)
   Diagnosis:         I10 (Essential hypertension)
   Pre-authorization: AUTH-2026-00312 (Approved 01/10/2026)
   Billed Amount:     $1,850.00
   Network Discount:  -$1,017.50
   Plan Paid (80%):   -$665.00
   YOUR COINSURANCE:  $167.50  ← 20% coinsurance after deductible

5. Prescription — Lisinopril 10mg (90-day mail order)
   Date of Service:   January 16, 2026
   Pharmacy:          CVS Caremark Mail Order
   NDC:               00591-0407-01
   Quantity:          90 tablets
   Days Supply:       90
   Retail Price:      $67.50
   Plan Paid:         -$57.50
   YOUR COPAY:        $10.00   ← Tier 1 generic

6. Prescription — Atorvastatin 20mg (30-day supply)
   Date of Service:   January 16, 2026
   Pharmacy:          CVS Pharmacy #4521
   NDC:               00378-3952-77
   Quantity:          30 tablets
   Days Supply:       30
   Retail Price:      $128.00
   Plan Paid:         -$113.00
   YOUR COPAY:        $15.00   ← Tier 2 preferred brand

7. Urgent Care Visit — Acute Sinusitis
   Date of Service:   January 22, 2026
   Provider:          MinuteClinic (Walk-in)
   Facility:          CVS MinuteClinic #4521
   Procedure Codes:   99203 (new patient, low complexity)
   Diagnosis:         J01.90 (Acute sinusitis, unspecified)
   Prescription:      Amoxicillin 500mg (10-day course) — $4.00 generic
   Billed Amount:     $195.00
   Network Discount:  -$78.00
   Plan Paid:         -$82.00
   YOUR COPAY:        $35.00   ← Urgent care copay
   Rx Copay:          $4.00    ← Tier 1 generic

MONTHLY TOTALS
--------------
Total Billed:               $3,287.50
Network Discounts:          -$1,610.20
Plan Paid:                  -$1,405.80
YOUR OUT-OF-POCKET TOTAL:   $271.50

   Breakdown:
   • Copays:        $104.00  (specialist $40 + Rx $10 + Rx $15 + urgent $35 + Rx $4)
   • Coinsurance:   $167.50  (echocardiogram 20%)
   • Deductible:    $0.00    (already met)

YEAR-TO-DATE ACCUMULATOR STATUS
---------------------------------
                           Used        Limit       Remaining
Individual Deductible:     $1,500.00   $1,500.00   $0.00 (MET)
Family Deductible:         $2,150.00   $3,000.00   $850.00
Out-of-Pocket Maximum:     $2,421.50   $6,000.00   $3,578.50
Preventive Visits Used:    1 of 1 annual
Specialist Visits (YTD):   3

NOTES & EXPLANATIONS
---------------------
• Claim #4 (Echocardiogram): Pre-authorization was obtained. After the
  negotiated rate, you owe 20% coinsurance ($167.50) because your
  individual deductible has already been met for the year. Without the
  network discount, this procedure would have cost $1,850.00.
• Claim #7 (Urgent Care): Consider using Telehealth ($0 copay) for
  non-emergency conditions like sinusitis to save on urgent care copays.
• Your prescription costs could be reduced by switching Atorvastatin to
  mail-order (90-day supply for $30 vs $45 for three 30-day fills).

APPEALS & QUESTIONS
--------------------
If you believe a claim was processed incorrectly, you may file an appeal
within 180 days. Contact Member Services: 1-800-555-PLAN (7526)
Online: www.goldppohealth.com/members | App: Gold Health Mobile
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
	"doc-stmt-003": `
HEALTH PLAN MONTHLY STATEMENT
==============================
Statement Period: November 1 - November 30, 2025
Member: Jane Doe
Member ID: HM-12345678
Plan: Gold PPO Health Plan

CLAIMS SUMMARY
--------------

1. Physical Therapy - Shoulder Rehabilitation
   Date of Service: November 8, 2025
   Provider: Peak Physical Therapy
   Billed Amount:    $180.00
   Plan Discount:    -$60.00
   Plan Paid:        -$80.00
   YOUR COPAY:       $40.00

2. Office Visit - Primary Care Follow-up
   Date of Service: November 22, 2025
   Provider: Dr. Sarah Smith, Family Medicine
   Facility: Lakewood Medical Center
   Billed Amount:    $150.00
   Plan Discount:    -$75.00
   Plan Paid:        -$50.00
   YOUR COPAY:       $25.00

MONTHLY TOTALS
--------------
Total Billed:       $330.00
Plan Discounts:     -$135.00
Plan Paid:          -$130.00
YOUR TOTAL:         $65.00
`,
	"doc-stmt-004": `
HEALTH PLAN MONTHLY STATEMENT
==============================
Statement Period: October 1 - October 31, 2025
Member: Jane Doe
Member ID: HM-12345678
Plan: Gold PPO Health Plan

CLAIMS SUMMARY
--------------

1. Urgent Care Visit
   Date of Service: October 3, 2025
   Provider: MinuteClinic
   Facility: CVS MinuteClinic #8821
   Billed Amount:    $195.00
   Plan Discount:    -$70.00
   Plan Paid:        -$50.00
   YOUR COPAY:       $75.00

2. Prescription - Amoxicillin 500mg (10-day supply)
   Date of Service: October 3, 2025
   Pharmacy: CVS Pharmacy #4521
   Retail Price:     $25.00
   Plan Paid:        -$15.00
   YOUR COPAY:       $10.00

3. Laboratory Services - Flu Test
   Date of Service: October 3, 2025
   Provider: MinuteClinic Lab
   Billed Amount:    $85.00
   Plan Discount:    -$42.50
   Plan Paid:        -$42.50
   YOUR COST:        $0.00

MONTHLY TOTALS
--------------
Total Billed:       $305.00
Plan Discounts:     -$112.50
Plan Paid:          -$107.50
YOUR TOTAL:         $85.00
`,
	"doc-bill-003": `
PREMIUM BILL
=============
Billing Period: December 2025
Due Date: December 15, 2025
Member: Jane Doe
Member ID: HM-12345678

Plan: Gold PPO Health Plan

Premium Breakdown:
  Medical:    $425.00
  Dental:     $45.00
  Vision:     $15.00
  -------------------
  TOTAL DUE:  $485.00

Payment Status: PAID (December 10, 2025)
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
	"doc-plan-002": `
PRESCRIPTION DRUG FORMULARY 2026
==================================
Plan: Gold PPO Health Plan
Effective: January 1, 2026

TIER 1 — GENERIC (Copay: $10)
  Lisinopril, Metformin, Amlodipine, Omeprazole,
  Amoxicillin, Atorvastatin, Levothyroxine, Losartan

TIER 2 — PREFERRED BRAND (Copay: $35)
  Eliquis, Jardiance, Ozempic, Xarelto, Humira

TIER 3 — NON-PREFERRED BRAND (Copay: $60)
  Entresto, Stelara, Dupixent, Skyrizi

TIER 4 — SPECIALTY (Copay: 20%, max $250)
  Keytruda, Opdivo, Revlimid

NOTES:
- Prior authorization required for Tier 3 and Tier 4
- Mail-order 90-day supply: 2.5x copay
- Step therapy applies to select medications
`,
	"doc-plan-003": `
PROVIDER NETWORK DIRECTORY 2026
=================================
Plan: Gold PPO Health Plan
Effective: January 1, 2026

PRIMARY CARE PHYSICIANS:
  Dr. Sarah Smith, Family Medicine — Lakewood Medical Center
  Dr. James Wilson, Internal Medicine — Riverside Health Group
  Dr. Maria Garcia, Family Medicine — Community Care Clinic

SPECIALISTS:
  Dr. Michael Chen, Dermatology — Downtown Dermatology Associates
  Dr. Lisa Park, Cardiology — Heart Care Specialists
  Dr. Robert Kim, Orthopedics — Orthopedic Sports Medicine Center

URGENT CARE:
  MinuteClinic (CVS locations) — In-network
  MedExpress Urgent Care — In-network

HOSPITALS:
  Lakewood Regional Medical Center — In-network
  St. Mary's Hospital — In-network
  University Health System — In-network

LABS:
  Quest Diagnostics — In-network (preferred)
  LabCorp — In-network

For a complete provider directory, visit www.example-healthplan.com/providers
`,
}

var uploadCounter int

func handleUpload(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	json.NewDecoder(r.Body).Decode(&body)

	uploadCounter++
	docID := fmt.Sprintf("doc-upload-%03d", uploadCounter)

	fileName := "uploaded-document"
	fileType := "pdf"
	var contentText string

	if docs, ok := body["documents"].([]any); ok && len(docs) > 0 {
		if doc, ok := docs[0].(map[string]any); ok {
			if fn, ok := doc["fileName"].(string); ok {
				fileName = fn
			}
			if ft, ok := doc["fileType"].(string); ok {
				fileType = ft
			}
			if contentB64, ok := doc["content"].(string); ok {
				if decoded, err := base64.StdEncoding.DecodeString(contentB64); err == nil {
					contentText = string(decoded)
				}
			}
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	newDoc := map[string]any{
		"id":   docID,
		"type": "document",
		"attributes": map[string]any{
			"title":        fileName,
			"fileType":     fileType,
			"createdDate":  now,
			"modifiedDate": now,
			"actions":      []string{"view", "download"},
		},
	}

	mu.Lock()
	uploadedDocs = append(uploadedDocs, newDoc)
	if contentText == "" {
		contentText = fmt.Sprintf("Uploaded document: %s\n(Content uploaded via chat on %s)", fileName, now)
	}
	documentContents[docID] = contentText
	mu.Unlock()

	log.Printf("[MOCK] POST /v1/documents (upload) → %s (%s) — now searchable under 'uploaded_by_me' and all docs", docID, fileName)

	writeJSON(w, http.StatusAccepted, map[string]any{"data": []map[string]any{newDoc}})
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
	case "uploaded_by_me":
		mu.RLock()
		docs = uploadedDocs
		mu.RUnlock()
	default:
		docs = getAllDocs()
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
		handleUpload(w, r)
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
