	package main

	import (
		"encoding/json"
		"fmt"
		"github.com/gorilla/mux"
		"github.com/loganrk/go-heap-cache"
		"github.com/rs/cors"
		"sync"
		"net/http"
	)

	type LRUCache struct {
		cache cache.Cache
		mutex sync.Mutex
	}

	func NewLRUCache() *LRUCache {
		config := cache.Config{
			Capacity: 100,               
			Expire: 5, // 5 sec               
			EvictionPolicy: cache.EVICTION_POLICY_LRU, 
		}

		cacheInstance := cache.New(&config)
		return &LRUCache{cache: cacheInstance}
	}

	func (lru *LRUCache) Get(key string) (string, bool) {
		lru.mutex.Lock()
		defer lru.mutex.Unlock()

		data, err := lru.cache.Get(key)
		if err != nil {
			return "", false
		}

		if value, ok := data.(string); ok {
			return value, true
		}
	
		return "", false
	}

	func (lru *LRUCache) Set(key, value string) {
		lru.mutex.Lock()
		defer lru.mutex.Unlock()
		
		lru.cache.Set(key, value)
	}


	func main() {
		mycache := NewLRUCache()

		r := mux.NewRouter()

		r.HandleFunc("/get/{key}", func(w http.ResponseWriter, r *http.Request) {
			params := mux.Vars(r)
			key := params["key"]
			
			value, exists := mycache.Get(key)
			if !exists {
				http.Error(w, "Key not found in cache", http.StatusNotFound)
				return
			}
			
			fmt.Println("Key: "+ key + ", Value: "+value)
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
			
			mycache.Set(cacheData.Key, cacheData.Value)
			
			fmt.Println("Value set in cache")
			response := map[string]string{"message": "Value set in cache"}
			json.NewEncoder(w).Encode(response)
		}).Methods("POST")

		corsHandler := cors.Default().Handler(r)

		
		fmt.Println("Server started at :8080")
		http.ListenAndServe(":8080", corsHandler)
	}
