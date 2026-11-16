// Command bench measures the HTTP handler over a real loopback socket.
//
//	go run ./cmd/bench -n 300 -c 8 -out docs/bench-results.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"go-service-template/internal/database"
	"go-service-template/internal/handler"
	"go-service-template/internal/repository"
	"go-service-template/internal/router"
	"go-service-template/internal/service"
	"gorm.io/gorm/logger"
)

type result struct {
	Runs        int     `json:"runs"`
	Concurrency int     `json:"concurrency"`
	Path        string  `json:"path"`
	P50ms       float64 `json:"p50_ms"`
	P95ms       float64 `json:"p95_ms"`
	P99ms       float64 `json:"p99_ms"`
	RequestsSec float64 `json:"requests_per_sec"`
}

func main() {
	runs := flag.Int("n", 300, "total requests")
	concurrency := flag.Int("c", 8, "parallel workers")
	out := flag.String("out", "", "optional JSON output path")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	dir, err := os.MkdirTemp("", "bench-*")
	if err != nil {
		log.Error("temp dir", "error", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	db, err := database.Open(filepath.Join(dir, "bench.db"))
	if err != nil {
		log.Error("database", "error", err)
		os.Exit(1)
	}
	// Keep the benchmark output machine-readable: no slow-query chatter.
	db.Logger = db.Logger.LogMode(logger.Silent)
	if err := database.Seed(db); err != nil {
		log.Error("seed", "error", err)
		os.Exit(1)
	}
	engine := router.New(
		handler.NewItemHandler(service.NewItems(repository.NewItemStore(db)), log),
		log, "bench",
	)
	server := httptest.NewServer(engine)
	defer server.Close()

	client := &http.Client{Timeout: 10 * time.Second}
	// Warm up the pool and the SQLite page cache.
	for index := 0; index < 10; index++ {
		response, err := client.Get(server.URL + "/api/v1/items?limit=20")
		if err != nil {
			log.Error("warm-up", "error", err)
			os.Exit(1)
		}
		response.Body.Close()
	}

	latencies := make([]time.Duration, 0, *runs)
	var mu sync.Mutex
	var wg sync.WaitGroup
	started := time.Now()
	jobs := make(chan int, *runs)
	for index := 0; index < *runs; index++ {
		jobs <- index
	}
	close(jobs)

	for worker := 0; worker < *concurrency; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				begin := time.Now()
				response, err := client.Get(server.URL + "/api/v1/items?limit=20")
				if err != nil {
					continue
				}
				response.Body.Close()
				elapsed := time.Since(begin)
				mu.Lock()
				latencies = append(latencies, elapsed)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	total := time.Since(started)

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	report := result{
		Runs:        len(latencies),
		Concurrency: *concurrency,
		Path:        "GET /api/v1/items?limit=20",
		P50ms:       percentile(latencies, 0.50),
		P95ms:       percentile(latencies, 0.95),
		P99ms:       percentile(latencies, 0.99),
		RequestsSec: float64(len(latencies)) / total.Seconds(),
	}
	encoded, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(encoded))
	if *out != "" {
		if err := os.WriteFile(*out, append(encoded, '\n'), 0o644); err != nil {
			log.Error("write report", "error", err)
			os.Exit(1)
		}
	}
}

func percentile(values []time.Duration, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	index := int(float64(len(values))*p) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(values) {
		index = len(values) - 1
	}
	return float64(values[index].Microseconds()) / 1000.0
}
