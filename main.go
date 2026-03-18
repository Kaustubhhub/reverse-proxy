package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

func GetNextBackendServer(backendServerList []string, currentBackendServer *int) int {
	*currentBackendServer = (*currentBackendServer + 1) % len(backendServerList)
	return *currentBackendServer
}

func main() {

	backendServerList := []string{"http://localhost:9001", "http://localhost:9002"}

	currentBackendServer := -1
	var mu sync.Mutex

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		idx := GetNextBackendServer(backendServerList, &currentBackendServer)
		mu.Unlock()

		target, err := url.Parse(backendServerList[idx])
		if err != nil {
			http.Error(w, "Bad Backend url", http.StatusInternalServerError)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(target)
		log.Println("Forwarding request to:", backendServerList[idx])
		proxy.ServeHTTP(w, r)
	})

	log.Println("Proxy running on :8080")
	http.ListenAndServe(":8080", nil)
}
