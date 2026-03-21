package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"sync"

	"github.com/kaustubhhub/reverse-proxy/config"
	"github.com/kaustubhhub/reverse-proxy/utils"
)

func main() {

	backendServerList, err2 := utils.LoadConfig(config.SERVERS_PATH)

	if err2 != nil {
		panic(err2)
	}

	currentBackendServer := -1
	var mu sync.Mutex
	proxies := make([]*httputil.ReverseProxy, len(backendServerList.Servers))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		idx := utils.GetNextBackendServer(backendServerList.Servers, &currentBackendServer)
		mu.Unlock()

		utils.ProxyWithFailover(w, r, backendServerList.Servers, idx, proxies)
	})

	log.Println("Proxy running on :8080")
	http.ListenAndServe(":8080", nil)
}
