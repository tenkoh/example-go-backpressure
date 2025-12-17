package main

import (
	"bufio"
	"context"
	"flag"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	targetURL := flag.String("url", "http://localhost:8080/events", "SSE endpoint URL")
	delay := flag.Duration("delay", time.Second, "delay between reading events")
	limit := flag.Int("limit", 0, "number of events to read before exiting (0 for unlimited)")
	requestTimeout := flag.Duration("request-timeout", 5*time.Minute, "overall request timeout (0 for none)")
	flag.Parse()

	ctx := context.Background()
	var cancel context.CancelFunc
	if *requestTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, *requestTimeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, *targetURL, nil)
	if err != nil {
		log.Fatalf("build request: %v", err)
	}
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("unexpected status: %s", resp.Status)
	}

	log.Printf("connected to %s (delay=%s)", *targetURL, delay.String())
	readEvents(resp.Body, *delay, *limit)
}

func readEvents(body io.Reader, delay time.Duration, limit int) {
	reader := bufio.NewReader(body)
	var dataLines []string
	events := 0

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("server closed connection")
				return
			}
			log.Fatalf("read error: %v", err)
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			if len(dataLines) > 0 {
				events++
				payload := strings.Join(dataLines, "\n")
				log.Printf("event %d received (payload length=%d)", events, len(payload))
				dataLines = dataLines[:0]

				if limit > 0 && events >= limit {
					log.Printf("limit reached (%d), exiting", limit)
					return
				}
				if delay > 0 {
					time.Sleep(delay)
				}
			}
			continue
		}

		if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
}
