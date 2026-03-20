package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

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
		start := time.Now()
		proxy.ServeHTTP(w, r)
		t := time.Now()
		elapsed := t.Sub(start)
		log.Printf(
			"%s %s | client=%s | backend=%s | latency=%s",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			backendServerList.Servers[idx],
			elapsed.Round(time.Millisecond),
		)
	})

	log.Println("Proxy running on :8080")
	http.ListenAndServe(":8080", nil)
}
