#  Reverse Proxy Load Balancer (Go)

A simple yet powerful **reverse proxy + load balancer** built in Go with:

- Round-robin load balancing  
- Automatic failover (retry on backend failure)  
- Proxy reuse for better performance  
- Request logging (method, path, backend, latency, status)  
- Config-driven backend servers  

---

##  How it works

- Incoming requests hit the proxy
- Proxy selects backend using **round-robin**
- If a backend fails → automatically retries the next backend
- Returns the first successful response

---

## ✨ Features

###  Round Robin
Distributes traffic evenly across backend servers

###  Failover / Retry
If a backend is down:

### Proxy Reuse
- Proxies are created once at startup
- Avoids per-request overhead

### Logging
Example:
```
GET /api | client=127.0.0.1 | backend=http://localhost:9002 | status=200 | latency=5ms
```


---

## ⚙️ Setup & Run Locally

### 1️⃣ Clone the repository
```
git clone https://github.com/kaustubhhub/reverse-proxy.git
cd reverse-proxy
```
---
### 2️⃣ Install dependencies
```
go mod tidy
```

---

### 3️⃣ Configure backend servers
```
Edit: config/servers.yaml
```


Example:
```yaml
servers:
  - http://localhost:9001
  - http://localhost:9002
```

Run the proxy
```
go run main.go
```
Output:
```
Proxy running on :8080
```

Test the proxy
```
curl http://localhost:8080
```

Run multiple times → requests will alternate between backends
