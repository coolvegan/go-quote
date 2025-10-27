package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"gittea.kittel.dev/marco/go-quote/internal/quotes"
)

func main() {
	quotesFile := "./misc/author-quote.txt"
	allQuotes, err := quotes.LoadQuotes(quotesFile)
	if err != nil {
		log.Fatalf("Failed to load quotes: %v", err)
	}

	apiKey := os.Getenv("API_KEY")

	// In-Memory-Hashmap für clientId und genutzte Quotes pro Tag
	type clientQuotes struct {
		Used map[string]time.Time // quote text -> timestamp
	}

	var (
		clientMap   = make(map[string]*clientQuotes)
		clientMapMu sync.Mutex
	)

	// Goroutine: Täglich um Mitternacht Used-Map leeren
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			time.Sleep(time.Until(next))
			clientMapMu.Lock()
			for _, cq := range clientMap {
				cq.Used = make(map[string]time.Time)
			}
			clientMapMu.Unlock()
		}
	}()

	// Middleware für API-Key-Auth
	quotesHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(allQuotes)
	}

	randomHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		clientId := r.URL.Query().Get("clientId")
		var quote quotes.Quote
		if clientId == "" {
			// Kein clientId: beliebiges zufälliges Zitat
			quote = allQuotes[rand.Intn(len(allQuotes))]
			json.NewEncoder(w).Encode(quote)
			return
		}
		// Mit clientId: Zitat nicht doppelt pro Tag
		clientMapMu.Lock()
		cq, ok := clientMap[clientId]
		if !ok {
			cq = &clientQuotes{Used: make(map[string]time.Time)}
			clientMap[clientId] = cq
		}
		// Filtere ungenutzte Quotes
		unused := []quotes.Quote{}
		for _, q := range allQuotes {
			if _, used := cq.Used[q.Text]; !used {
				unused = append(unused, q)
			}
		}
		if len(unused) == 0 {
			clientMapMu.Unlock()
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"no more quotes for today"}`))
			return
		}
		quote = unused[rand.Intn(len(unused))]
		cq.Used[quote.Text] = time.Now()
		clientMapMu.Unlock()
		json.NewEncoder(w).Encode(quote)
	}

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

	http.HandleFunc("/quotes", authMiddleware(quotesHandler))
	http.HandleFunc("/random", authMiddleware(randomHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
