package utils

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/kaustubhhub/reverse-proxy/config"
	"gopkg.in/yaml.v3"
)

func GetNextBackendServer(backendServerList []string, currentBackendServer *int) int {
	*currentBackendServer = (*currentBackendServer + 1) % len(backendServerList)
	return *currentBackendServer
}

func LoadConfig(path string) (*config.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg config.Config

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func ProxyWithFailover(w http.ResponseWriter, r *http.Request, servers []string, startIdx int) {
	if len(servers) == 0 {
		http.Error(w, "No backends configured", http.StatusInternalServerError)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	attempts := len(servers)
	var lastErr error
	var lastStatus int
	var lastBody []byte

	for i := range attempts {
		backendIdx := (startIdx + i) % len(servers)
		targetURL, err := url.Parse(servers[backendIdx])
		if err != nil {
			log.Printf("invalid backend URL %s: %v", servers[backendIdx], err)
			lastErr = err
			continue
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		var transportErr error
		proxy.ErrorHandler = func(_ http.ResponseWriter, _ *http.Request, err error) {
			transportErr = err
		}

		rec := httptest.NewRecorder()
		reqCopy := r.Clone(r.Context())
		reqCopy.URL.Scheme = targetURL.Scheme
		reqCopy.URL.Host = targetURL.Host
		reqCopy.Host = targetURL.Host
		reqCopy.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		start := time.Now()
		proxy.ServeHTTP(rec, reqCopy)
		elapsed := time.Since(start)

		if transportErr != nil {
			lastErr = transportErr
			log.Printf("backend failure %s: %v (try next)", servers[backendIdx], transportErr)
			continue
		}

		for k, values := range rec.Header() {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(rec.Code)
		w.Write(rec.Body.Bytes())

		log.Printf("%s %s | client=%s | backend=%s | latency=%s | status=%d",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			servers[backendIdx],
			elapsed.Round(time.Millisecond),
			rec.Code,
		)
		return
	}

	if lastErr != nil {
		log.Printf("all backends failed: %v", lastErr)
		http.Error(w, "All backends unavailable", http.StatusBadGateway)
		return
	}

	// fallback if no explicit transport error was captured
	if lastStatus != 0 {
		w.WriteHeader(lastStatus)
		w.Write(lastBody)
		return
	}

	http.Error(w, "All backends unavailable", http.StatusBadGateway)
}
