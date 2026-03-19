package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/kaustubhhub/reverse-proxy/utils"
)

func main() {

	// backendServerList := []string{"http://localhost:9001", "http://localhost:9002"}
	backendServerList, err2 := utils.LoadConfig("./config/servers.yaml")

	if err2 != nil {
		panic(err2)
	}

	fmt.Println("list of backends : ", backendServerList.Servers)

	currentBackendServer := -1
	var mu sync.Mutex

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		idx := utils.GetNextBackendServer(backendServerList.Servers, &currentBackendServer)
		mu.Unlock()

		target, err := url.Parse(backendServerList.Servers[idx])
		if err != nil {
			http.Error(w, "Bad Backend url", http.StatusInternalServerError)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(target)
		log.Println("Forwarding request to:", backendServerList.Servers[idx])
		proxy.ServeHTTP(w, r)
	})

	log.Println("Proxy running on :8080")
	http.ListenAndServe(":8080", nil)
}
