package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	mode := flag.String("mode", "naive", "Booking mode: naive, pessimistic, optimistic")
	users := flag.Int("users", 200, "Number of concurrent users")
	gateway := flag.String("gateway", "http://localhost:8000", "Gateway URL")
	tierID := flag.String("tier", "", "Tier ID to book (auto-detect if empty)")
	dbURL := flag.String("db", "", "Event database URL for verification")
	flag.Parse()

	if *dbURL == "" {
		*dbURL = "postgres://ticketbox:ticketbox_secret@localhost:5434/ticketbox_event?sslmode=disable"
	}

	fmt.Println("=== TicketBox Double Booking Load Test ===")
	fmt.Printf("Mode:             %s\n", *mode)
	fmt.Printf("Concurrent users: %d\n", *users)
	fmt.Printf("Gateway:          %s\n", *gateway)
	fmt.Println()

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        *users,
			MaxIdleConnsPerHost: *users,
			IdleConnTimeout:     30 * time.Second,
		},
		Timeout: 30 * time.Second,
	}

	// Step 1: Get events and find a tier
	if *tierID == "" {
		fmt.Println("Auto-detecting tier ID...")
		resp, err := client.Get(*gateway + "/api/events")
		if err != nil {
			log.Fatalf("Failed to list events: %v", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		var eventsResp struct {
			Events []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
				Tiers []struct {
					ID                string `json:"id"`
					Name              string `json:"name"`
					AvailableQuantity int32  `json:"available_quantity"`
				} `json:"tiers"`
			} `json:"events"`
		}
		if err := json.Unmarshal(body, &eventsResp); err != nil {
			log.Fatalf("Failed to parse events: %v", err)
		}

		// Find the first tier with limited tickets (Premium/Courtside/Cageside)
		for _, e := range eventsResp.Events {
			for _, t := range e.Tiers {
				if t.AvailableQuantity <= 100 && t.AvailableQuantity > 0 {
					*tierID = t.ID
					fmt.Printf("Selected: %s - %s (%d available)\n", e.Title, t.Name, t.AvailableQuantity)
					break
				}
			}
			if *tierID != "" {
				break
			}
		}

		if *tierID == "" {
			log.Fatal("No suitable tier found")
		}
	}

	// Also need the event ID for the tier
	var eventID string
	{
		resp, err := client.Get(*gateway + "/api/events")
		if err != nil {
			log.Fatalf("Failed to list events: %v", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		var eventsResp struct {
			Events []struct {
				ID    string `json:"id"`
				Tiers []struct {
					ID string `json:"id"`
				} `json:"tiers"`
			} `json:"events"`
		}
		json.Unmarshal(body, &eventsResp)
		for _, e := range eventsResp.Events {
			for _, t := range e.Tiers {
				if t.ID == *tierID {
					eventID = e.ID
					break
				}
			}
		}
	}

	// Step 2: Register test users and get tokens
	fmt.Printf("\nRegistering %d test users...\n", *users)
	tokens := make([]string, *users)
	var registerWg sync.WaitGroup
	for i := 0; i < *users; i++ {
		registerWg.Add(1)
		go func(idx int) {
			defer registerWg.Done()
			email := fmt.Sprintf("load-test-%d-%d@test.com", time.Now().UnixNano(), idx)
			reqBody, _ := json.Marshal(map[string]string{
				"email":    email,
				"password": "loadtest123",
				"name":     fmt.Sprintf("Load Test User %d", idx),
			})

			resp, err := client.Post(*gateway+"/api/auth/register", "application/json", bytes.NewReader(reqBody))
			if err != nil {
				log.Printf("Register failed for user %d: %v", idx, err)
				return
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			var authResp struct {
				AccessToken string `json:"access_token"`
			}
			json.Unmarshal(body, &authResp)
			tokens[idx] = authResp.AccessToken
		}(i)
	}
	registerWg.Wait()

	validTokens := 0
	for _, t := range tokens {
		if t != "" {
			validTokens++
		}
	}
	fmt.Printf("Registered: %d/%d users\n", validTokens, *users)

	// Step 3: Fire concurrent booking requests
	fmt.Printf("\nFiring %d concurrent booking requests (mode=%s)...\n", validTokens, *mode)
	var (
		confirmed int64
		failed    int64
		errors    int64
		wg        sync.WaitGroup
		startCh   = make(chan struct{})
	)

	for i := 0; i < *users; i++ {
		if tokens[i] == "" {
			continue
		}
		wg.Add(1)
		go func(token string) {
			defer wg.Done()
			<-startCh // Wait for all goroutines to be ready

			reqBody, _ := json.Marshal(map[string]interface{}{
				"event_id": eventID,
				"items": []map[string]interface{}{
					{"ticket_tier_id": *tierID, "quantity": 1},
				},
			})

			req, _ := http.NewRequest("POST", *gateway+"/api/bookings", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("X-Booking-Mode", *mode)

			resp, err := client.Do(req)
			if err != nil {
				atomic.AddInt64(&errors, 1)
				return
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			var bookingResp struct {
				Status string `json:"status"`
			}
			json.Unmarshal(body, &bookingResp)

			if resp.StatusCode == 201 && bookingResp.Status == "CONFIRMED" {
				atomic.AddInt64(&confirmed, 1)
			} else {
				atomic.AddInt64(&failed, 1)
			}
		}(tokens[i])
	}

	start := time.Now()
	close(startCh) // Fire all at once!
	wg.Wait()
	duration := time.Since(start)

	// Step 4: Check actual availability from database
	var finalAvailable int32 = -9999
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, *dbURL)
	if err == nil {
		defer pool.Close()
		pool.QueryRow(ctx,
			"SELECT available_quantity FROM ticket_tiers WHERE id = $1", *tierID).Scan(&finalAvailable)
	}

	// Step 5: Print results
	fmt.Println()
	fmt.Println("=== Results ===")
	fmt.Printf("Mode:              %s\n", *mode)
	fmt.Printf("Total requests:    %d\n", confirmed+failed+errors)
	fmt.Printf("Confirmed:         %d\n", confirmed)
	fmt.Printf("Failed:            %d\n", failed)
	fmt.Printf("Errors:            %d\n", errors)
	fmt.Printf("Duration:          %s\n", duration.Round(time.Millisecond))
	if finalAvailable != -9999 {
		fmt.Printf("Available after:   %d\n", finalAvailable)
	}
	fmt.Println()

	if *mode == "naive" && finalAvailable < 0 {
		fmt.Println("DOUBLE BOOKING DETECTED! Available quantity is NEGATIVE.")
		fmt.Println("This proves the naive approach is NOT safe under concurrency.")
	} else if *mode != "naive" && finalAvailable >= 0 {
		fmt.Println("No double booking. Locking strategy worked correctly.")
	}

	os.Exit(0)
}
