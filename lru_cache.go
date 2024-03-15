	package main

	import (
		"encoding/json"
		"fmt"
		"github.com/gorilla/mux"
		"github.com/rs/cors"
		"sync"
		"net/http"
		"time"
	)

type Cache struct {
	Key        string    `json:"key"`
	Value      string    `json:"value"`
	Expiration time.Time `json:"expiration"`
}

type LRUCache struct {
	cache     map[string]*Cache
	order     []*Cache
	maxSize   int
	mutex     sync.Mutex
}


func NewLRUCache(maxSize int) *LRUCache {
	return &LRUCache{
		cache:   make(map[string]*Cache),
		order:   make([]*Cache, 0),
		maxSize: maxSize,
		mutex:   sync.Mutex{},
	}
}

func (lru *LRUCache) Get(key string) (string, bool) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	item, exists := lru.cache[key]
	if !exists || item.Expiration.Before(time.Now()) {
		return "", false
	}

	lru.updateOrder(item)
	return item.Value, true
}

func (lru *LRUCache) Set(key, value string, expiration time.Time) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if len(lru.cache) >= lru.maxSize {
		lru.deleteLeastRU()
	}

	item := &Cache{
		Key:        key,
		Value:      value,
		Expiration: expiration,
	}

	lru.cache[key] = item
	lru.order = append(lru.order, item)
}

func (lru *LRUCache) deleteLeastRU() {
	if len(lru.order) == 0 {
		return
	}

	item := lru.order[0]
	delete(lru.cache, item.Key)
	lru.order = lru.order[1:]
}

func (lru *LRUCache) updateOrder(item *Cache) {
	for i, cachedItem := range lru.order {
		if cachedItem == item {
			// Move item to the front of the order
			lru.order = append(lru.order[:i], lru.order[i+1:]...)
			lru.order = append([]*Cache{item}, lru.order...)
			break
		}
	}
}

	func main() {
		cache := NewLRUCache(1024)

		r := mux.NewRouter()

		r.HandleFunc("/get/{key}", func(w http.ResponseWriter, r *http.Request) {
			params := mux.Vars(r)
			key := params["key"]
		
			value, exists := cache.Get(key)
			if !exists {
				http.Error(w, "Key not found in cache", http.StatusNotFound)
				return
			}
		
			response := map[string]string{"key": key, "value": value}
			json.NewEncoder(w).Encode(response)
		}).Methods("GET")
		
		r.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
			var cacheData struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			}
		
			err := json.NewDecoder(r.Body).Decode(&cacheData)
			if err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
		
			cache.Set(cacheData.Key, cacheData.Value, time.Now().Add(5*time.Second))
		
			response := map[string]string{"message": "Value set in cache"}
			json.NewEncoder(w).Encode(response)
		}).Methods("POST")

		corsHandler := cors.Default().Handler(r)

		
		fmt.Println("Server started at :8080")
		http.ListenAndServe(":8080", corsHandler)
	}
