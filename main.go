package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {

	backend_servers := [...]string{"http://localhost:9001", "http://localhost:9002"}

	target, _ := url.Parse()

	proxy := httputil.NewSingleHostReverseProxy(target)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Forwarding request to backend server")
		proxy.ServeHTTP(w, r)
	})

	log.Println("Proxy running on :8080")
	http.ListenAndServe(":8080", nil)
}
