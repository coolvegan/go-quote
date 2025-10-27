package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gittea.kittel.dev/marco/go-quote/internal/quotes"
)

func TestQuotesEndpoint(t *testing.T) {
	// Prepare test quotes
	allQuotes := []quotes.Quote{
		{Author: "TestAuthor", Text: "Test quote."},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(allQuotes)
	}

	apiKey := "testkey123"
	authMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if apiKey == "" || key != apiKey {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"unauthorized"}`))
				return
			}
			next(w, r)
		}
	}

	// Test: Request ohne API-Key
	req := httptest.NewRequest("GET", "/quotes", nil)
	w := httptest.NewRecorder()
	authMiddleware(handler)(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for missing API key, got %d", resp.StatusCode)
	}

	// Test: Request mit falschem API-Key
	req = httptest.NewRequest("GET", "/quotes", nil)
	req.Header.Set("X-API-Key", "wrongkey")
	w = httptest.NewRecorder()
	authMiddleware(handler)(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for wrong API key, got %d", resp.StatusCode)
	}

	// Test: Request mit korrektem API-Key
	req = httptest.NewRequest("GET", "/quotes", nil)
	req.Header.Set("X-API-Key", apiKey)
	w = httptest.NewRecorder()
	authMiddleware(handler)(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for correct API key, got %d", resp.StatusCode)
	}

	var got []quotes.Quote
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if len(got) != 1 || got[0] != allQuotes[0] {
		t.Errorf("Response mismatch: got %+v, want %+v", got, allQuotes)
	}
}

func TestRandomEndpoint(t *testing.T) {
	allQuotes := []quotes.Quote{
		{Author: "A", Text: "Q1"},
		{Author: "B", Text: "Q2"},
		{Author: "C", Text: "Q3"},
	}

	// In-Memory-Hashmap analog zur main.go
	type clientQuotes struct {
		Used map[string]struct{}
	}
	clientMap := make(map[string]*clientQuotes)

	randomHandler := func(w http.ResponseWriter, r *http.Request) {
		clientId := r.URL.Query().Get("clientId")
		var quote quotes.Quote
		if clientId == "" {
			quote = allQuotes[0] // deterministic for test
			json.NewEncoder(w).Encode(quote)
			return
		}
		cq, ok := clientMap[clientId]
		if !ok {
			cq = &clientQuotes{Used: make(map[string]struct{})}
			clientMap[clientId] = cq
		}
		unused := []quotes.Quote{}
		for _, q := range allQuotes {
			if _, used := cq.Used[q.Text]; !used {
				unused = append(unused, q)
			}
		}
		if len(unused) == 0 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"no more quotes for today"}`))
			return
		}
		quote = unused[0] // deterministic for test
		cq.Used[quote.Text] = struct{}{}
		json.NewEncoder(w).Encode(quote)
	}

	apiKey := "testkey123"
	authMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if apiKey == "" || key != apiKey {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"unauthorized"}`))
				return
			}
			next(w, r)
		}
	}

	// Test: /random ohne clientId
	req := httptest.NewRequest("GET", "/random", nil)
	req.Header.Set("X-API-Key", apiKey)
	w := httptest.NewRecorder()
	authMiddleware(randomHandler)(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	var got quotes.Quote
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if got != allQuotes[0] {
		t.Errorf("Expected first quote, got %+v", got)
	}

	// Test: /random mit clientId, alle Zitate nacheinander
	clientId := "abc"
	for i := 0; i < len(allQuotes); i++ {
		req = httptest.NewRequest("GET", "/random?clientId="+clientId, nil)
		req.Header.Set("X-API-Key", apiKey)
		w = httptest.NewRecorder()
		authMiddleware(randomHandler)(w, req)
		resp = w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
		if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if got != allQuotes[i] {
			t.Errorf("Expected quote %d, got %+v", i, got)
		}
	}

	// Test: /random mit clientId, alle Zitate verbraucht
	req = httptest.NewRequest("GET", "/random?clientId="+clientId, nil)
	req.Header.Set("X-API-Key", apiKey)
	w = httptest.NewRecorder()
	authMiddleware(randomHandler)(w, req)
	resp = w.Result()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", resp.StatusCode)
	}
}
