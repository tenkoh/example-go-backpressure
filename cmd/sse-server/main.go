package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	tokenBytes := flag.Int("token-bytes", 1024, "bytes per token payload")
	flag.Parse()

	handler := eventsHandler(*tokenBytes)
	mux := http.NewServeMux()
	mux.Handle("/events", handler)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("starting SSE server on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func startTokenStream(ctx context.Context, tokenBytes int) <-chan string {
	tokens := make(chan string, 100)
	go func() {
		defer close(tokens)
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				token := buildResponseToken(tokenBytes)
				select {
				case tokens <- token:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return tokens
}

func buildResponseToken(tokenBytes int) string {
	return strings.Repeat("x", tokenBytes)
}

func eventsHandler(tokenBytes int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, _ := w.(http.Flusher)

		ctx := r.Context()
		tokens := startTokenStream(ctx, tokenBytes)
		log.Printf("client connected (payload=%d bytes)", tokenBytes)
		defer log.Println("client disconnected")

		for {
			select {
			case <-ctx.Done():
				return
			case token, ok := <-tokens:
				if !ok {
					return
				}
				if _, err := fmt.Fprintf(w, "data: %s\n\n", token); err != nil {
					log.Printf("write failed: %v", err)
					return
				}
				flusher.Flush()
			}
		}
	})
}
