package main

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/dgraph-io/ristretto"
)

var cache *ristretto.Cache

func main() {
	cache, _ = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30, // 1GB cap — never grows past this
		BufferItems: 64,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/get", handleGet)
	mux.HandleFunc("/set", handleSet)
	http.ListenAndServe(":8090", mux)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	val, found := cache.Get(key)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"value": val})
}

func handleSet(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key   string `json:"key"`
		Value any    `json:"value"`
		TTL   int    `json:"ttl_seconds"`
	}
	body, _ := io.ReadAll(r.Body)
	json.Unmarshal(body, &req)
	cache.SetWithTTL(req.Key, req.Value, 1, time.Duration(req.TTL)*time.Second)
	cache.Wait()
	w.WriteHeader(http.StatusOK)
}
