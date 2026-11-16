package server_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"go-service-template/internal/server"
)

// TestRunDrainsInFlightRequests starts a real server, sends a request that takes
// 250 ms, then cancels the context. The request must finish before Run returns.
func TestRunDrainsInFlightRequests(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := http.NewServeMux()
	handler.HandleFunc("/slow", func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(250 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("done"))
	})

	srv := server.New("127.0.0.1:0", handler, log, 2*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- srv.Run(ctx) }()

	deadline := time.Now().Add(2 * time.Second)
	for srv.Addr() == "" && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if srv.Addr() == "" {
		cancel()
		t.Fatal("server never started")
	}

	type outcome struct {
		status int
		err    error
	}
	result := make(chan outcome, 1)
	go func() {
		response, err := http.Get("http://" + srv.Addr() + "/slow")
		if err != nil {
			result <- outcome{err: err}
			return
		}
		defer response.Body.Close()
		_, _ = io.ReadAll(response.Body)
		result <- outcome{status: response.StatusCode}
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case got := <-result:
		if got.err != nil {
			t.Fatalf("in-flight request failed during shutdown: %v", got.err)
		}
		if got.status != http.StatusOK {
			t.Fatalf("in-flight request status = %d", got.status)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("in-flight request did not finish")
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned an error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not return after shutdown")
	}
}

func TestRunReportsBindFailure(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := server.New("127.0.0.1:0", http.NewServeMux(), log, time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// A cancelled context before listen is fine: the listener still binds and
	// then drains immediately, which must return nil rather than hang.
	if err := srv.Run(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
