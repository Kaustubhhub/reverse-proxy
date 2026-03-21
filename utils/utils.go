package utils

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
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

func ProxyWithFailover(
	w http.ResponseWriter,
	r *http.Request,
	servers []string,
	startIdx int,
	proxies []*httputil.ReverseProxy,
) {
	if len(servers) == 0 {
		http.Error(w, "No backends configured", http.StatusInternalServerError)
		return
	}

	// Read body once (for retries)
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	attempts := len(servers)

	for i := 0; i < attempts; i++ {
		backendIdx := (startIdx + i) % len(servers)
		proxy := proxies[backendIdx]

		var transportErr error

		// Capture transport-level errors (backend down, timeout, etc.)
		proxy.ErrorHandler = func(_ http.ResponseWriter, _ *http.Request, err error) {
			transportErr = err
		}

		// Clone request
		reqCopy := r.Clone(r.Context())
		reqCopy.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		// Capture response
		rec := httptest.NewRecorder()

		start := time.Now()
		proxy.ServeHTTP(rec, reqCopy)
		elapsed := time.Since(start)

		// If backend failed → try next
		if transportErr != nil {
			log.Printf(
				"attempt=%d | backend=%s | error=%v",
				i+1,
				servers[backendIdx],
				transportErr,
			)
			continue
		}

		// Success → copy response to client
		for k, values := range rec.Header() {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}

		w.WriteHeader(rec.Code)
		w.Write(rec.Body.Bytes())

		log.Printf(
			"%s %s | client=%s | backend=%s | status=%d | latency=%s",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			servers[backendIdx],
			rec.Code,
			elapsed.Round(time.Millisecond),
		)

		return
	}

	// All backends failed
	log.Println("all backends failed")
	http.Error(w, "All backends unavailable", http.StatusBadGateway)
}
