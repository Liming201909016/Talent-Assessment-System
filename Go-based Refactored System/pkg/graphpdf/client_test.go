package graphpdf

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestClientConvertUploadsConvertsAndDeletes(t *testing.T) {
	var mu sync.Mutex
	requests := make([]string, 0, 4)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests = append(requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		mu.Unlock()
		switch {
		case r.URL.Path == "/tenant/oauth2/v2.0/token":
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"access_token":"test-token","expires_in":3600}`)
		case r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/content"):
			if r.Header.Get("Authorization") != "Bearer test-token" {
				t.Fatalf("authorization header=%q", r.Header.Get("Authorization"))
			}
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"id":"item-1"}`)
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/items/item-1/content"):
			w.Header().Set("Content-Type", "application/pdf")
			io.WriteString(w, "%PDF-1.7\n"+strings.Repeat("x", 2048))
		case r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/items/item-1"):
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "unexpected", http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient(Config{TenantID: "tenant", ClientID: "client", ClientSecret: "secret", DriveID: "drive", Folder: "reports", TimeoutSeconds: 10}, server.Client())
	client.tokenBaseURL = server.URL
	client.graphBaseURL = server.URL
	pdf, err := client.Convert(context.Background(), "report.docx", []byte("docx"))
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if !strings.HasPrefix(string(pdf), "%PDF-") {
		t.Fatalf("not a PDF: %q", pdf[:8])
	}
	mu.Lock()
	joined := strings.Join(requests, "\n")
	mu.Unlock()
	for _, expected := range []string{"POST /tenant/oauth2/v2.0/token?", "PUT /drives/drive/root:/reports/", "GET /drives/drive/items/item-1/content?format=pdf", "DELETE /drives/drive/items/item-1?"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("request %q missing from:\n%s", expected, joined)
		}
	}
}

func TestClientConvertRejectsNonPDFAndStillDeletes(t *testing.T) {
	deleted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/token"):
			io.WriteString(w, `{"access_token":"test-token","expires_in":3600}`)
		case r.Method == http.MethodPut:
			io.WriteString(w, `{"id":"item-1"}`)
		case r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "text/html")
			io.WriteString(w, "not a pdf")
		case r.Method == http.MethodDelete:
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()
	client := NewClient(Config{TenantID: "tenant", ClientID: "client", ClientSecret: "secret", DriveID: "drive", TimeoutSeconds: 10}, server.Client())
	client.tokenBaseURL = server.URL
	client.graphBaseURL = server.URL
	if _, err := client.Convert(context.Background(), "report.docx", []byte("docx")); err == nil {
		t.Fatal("non-PDF response accepted")
	}
	if !deleted {
		t.Fatal("temporary Graph item was not deleted after conversion failure")
	}
}

func TestClientConvertRetriesTransientConversionStatus(t *testing.T) {
	getCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/token"):
			io.WriteString(w, `{"access_token":"test-token","expires_in":3600}`)
		case r.Method == http.MethodPut:
			io.WriteString(w, `{"id":"item-1"}`)
		case r.Method == http.MethodGet:
			getCount++
			if getCount == 1 {
				http.Error(w, "not ready", http.StatusServiceUnavailable)
				return
			}
			w.Header().Set("Content-Type", "application/pdf")
			io.WriteString(w, "%PDF-1.7\n"+strings.Repeat("x", 2048))
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()
	client := NewClient(Config{TenantID: "tenant", ClientID: "client", ClientSecret: "secret", DriveID: "drive", TimeoutSeconds: 10}, server.Client())
	client.tokenBaseURL = server.URL
	client.graphBaseURL = server.URL
	client.retryDelay = time.Millisecond
	if _, err := client.Convert(context.Background(), "report.docx", []byte("docx")); err != nil {
		t.Fatalf("transient conversion status was not retried: %v", err)
	}
	if getCount != 2 {
		t.Fatalf("conversion GET count=%d, want 2", getCount)
	}
}
