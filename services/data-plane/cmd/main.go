package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"goproxy/internal/balancer"
	"goproxy/internal/proxy"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type SafeRoutingTable struct {
	sync.RWMutex
	routes map[string]*balancer.RoundRobinBalancer
}

var table = SafeRoutingTable{
	routes: make(map[string]*balancer.RoundRobinBalancer),
}

var (
    httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "nexus_mesh_requests_total",
        Help: "Amount of HTTP requests processed",
    }, []string{"method", "endpoint", "status"})

    httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "nexus_mesh_request_duration_seconds",
        Help:    "Latency of requests in seconds",
        Buckets: prometheus.DefBuckets,
    }, []string{"endpoint"})
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	if _, err := rdb.Ping(ctx).Result(); err != nil {
        log.Printf("⚠️ Redis offline...")
    } else {
        loadRegistry(rdb)
        
        go watchConfiguration(rdb)
    }

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now() 

		var targetBalancer *balancer.RoundRobinBalancer
		var matchedPath string

		table.RLock()
		for path, b := range table.routes {
			if strings.HasPrefix(r.URL.Path, path) {
				targetBalancer = b
				matchedPath = path
				break
			}
		}
		table.RUnlock()

		if targetBalancer == nil {
			http.Error(w, "Service not found", http.StatusNotFound)
			return
		}

		target := targetBalancer.GetNext()
		if target == nil {
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}

		originalPath := r.URL.Path
		r.URL.Path = strings.TrimPrefix(r.URL.Path, matchedPath)
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}

		p := proxy.NewProxy(target.URL)
		r.Host = target.URL.Host
		
		log.Printf("%s %s -> %s%s", r.Method, originalPath, target.URL.Host, r.URL.Path)
		p.ServeHTTP(w, r)

		duration := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, "200").Inc()
        httpDuration.WithLabelValues(r.URL.Path).Observe(duration)
	})

	http.Handle("/metrics", promhttp.Handler())

	log.Println("Nexus-Mesh Data Plane listening on :8080 (Waiting for routes...)")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func startHealthCheck(lb *balancer.RoundRobinBalancer) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		for _, b := range lb.GetTargets() {
			b.SetStatus(pingServer(b))
		}
	}
}

func pingServer(b *balancer.Backend) bool {
	conn, err := net.DialTimeout("tcp", b.URL.Host, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

type ConfigUpdate struct {
	Path     string   `json:"path"`
	Backends []string `json:"backends"`
}

func watchConfiguration(rdb *redis.Client) {
	pubsub := rdb.Subscribe(ctx, "config_updates")
	defer pubsub.Close()

	for msg := range pubsub.Channel() {
		var update ConfigUpdate
		if err := json.Unmarshal([]byte(msg.Payload), &update); err != nil {
			log.Printf("Erro JSON: %v", err)
			continue
		}

		newLB := balancer.NewRoundRobinBalancer(update.Backends)
		
		table.Lock()
		table.routes[update.Path] = newLB
		table.Unlock()

		go startHealthCheck(newLB)
		log.Printf("🚀 Rota [%s] configurada dinamicamente", update.Path)
	}
}

func loadRegistry(rdb *redis.Client) {
    keys, err := rdb.Keys(ctx, "route:*").Result()
    if err != nil {
        log.Printf("❌ Erro ao ler Registry: %v", err)
        return
    }

    for _, key := range keys {
        val, _ := rdb.Get(ctx, key).Result()
        
        var update ConfigUpdate
        if err := json.Unmarshal([]byte(val), &update); err != nil {
            continue
        }

        newLB := balancer.NewRoundRobinBalancer(update.Backends)
        table.Lock()
        table.routes[update.Path] = newLB
        table.Unlock()

        go startHealthCheck(newLB)
        log.Printf("📥 [REGISTRY] Rota [%s] restaurada com sucesso", update.Path)
    }
}