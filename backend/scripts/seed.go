package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://ticketbox:ticketbox_secret@localhost:5434/ticketbox_event?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer pool.Close()

	// Clear existing data
	pool.Exec(ctx, "DELETE FROM ticket_tiers")
	pool.Exec(ctx, "DELETE FROM events_read_model")
	pool.Exec(ctx, "DELETE FROM events")

	fmt.Println("Seeding events...")

	events := []struct {
		ID       string
		Title    string
		Desc     string
		Category string
		Venue    string
		Location string
		Date     time.Time
		ImageURL string
		Tiers    []struct {
			ID       string
			Name     string
			Price    int64
			Quantity int32
		}
	}{
		{
			ID:       "a0000000-0000-0000-0000-000000000001",
			Title:    "The Weeknd — After Hours World Tour",
			Desc:     "Experience The Weeknd live on his After Hours World Tour. An unforgettable night of music featuring hits from his latest album.",
			Category: "concert",
			Venue:    "Madison Square Garden",
			Location: "New York, NY",
			Date:     time.Date(2026, 7, 15, 20, 0, 0, 0, time.UTC),
			ImageURL: "",
			Tiers: []struct {
				ID       string
				Name     string
				Price    int64
				Quantity int32
			}{
				{ID: "b0000000-0000-0000-0000-000000000001", Name: "General", Price: 13000, Quantity: 500},
				{ID: "b0000000-0000-0000-0000-000000000002", Name: "VIP", Price: 35000, Quantity: 200},
				{ID: "b0000000-0000-0000-0000-000000000003", Name: "Premium", Price: 55000, Quantity: 50},
			},
		},
		{
			ID:       "a0000000-0000-0000-0000-000000000002",
			Title:    "Billie Eilish — Hit Me Hard and Soft Tour",
			Desc:     "Billie Eilish brings her ethereal sound to the O2 Arena for a night of mesmerizing performances.",
			Category: "concert",
			Venue:    "O2 Arena",
			Location: "London, UK",
			Date:     time.Date(2026, 4, 8, 19, 30, 0, 0, time.UTC),
			ImageURL: "",
			Tiers: []struct {
				ID       string
				Name     string
				Price    int64
				Quantity int32
			}{
				{ID: "b0000000-0000-0000-0000-000000000004", Name: "General", Price: 8500, Quantity: 800},
				{ID: "b0000000-0000-0000-0000-000000000005", Name: "VIP", Price: 25000, Quantity: 300},
				{ID: "b0000000-0000-0000-0000-000000000006", Name: "Premium", Price: 45000, Quantity: 100},
			},
		},
		{
			ID:       "a0000000-0000-0000-0000-000000000003",
			Title:    "NBA Finals 2026 — Game 3",
			Desc:     "The NBA Finals come to San Francisco. Watch the best basketball players in the world compete for the championship.",
			Category: "sports",
			Venue:    "Chase Center",
			Location: "San Francisco, CA",
			Date:     time.Date(2026, 6, 12, 21, 0, 0, 0, time.UTC),
			ImageURL: "",
			Tiers: []struct {
				ID       string
				Name     string
				Price    int64
				Quantity int32
			}{
				{ID: "b0000000-0000-0000-0000-000000000007", Name: "Upper Level", Price: 12000, Quantity: 1000},
				{ID: "b0000000-0000-0000-0000-000000000008", Name: "Lower Level", Price: 35000, Quantity: 400},
				{ID: "b0000000-0000-0000-0000-000000000009", Name: "Courtside", Price: 120000, Quantity: 50},
			},
		},
		{
			ID:       "a0000000-0000-0000-0000-000000000004",
			Title:    "UFC 310 — Championship",
			Desc:     "UFC Championship night at T-Mobile Arena. Witness the most anticipated fights of the year.",
			Category: "sports",
			Venue:    "T-Mobile Arena",
			Location: "Las Vegas, NV",
			Date:     time.Date(2026, 5, 3, 22, 0, 0, 0, time.UTC),
			ImageURL: "",
			Tiers: []struct {
				ID       string
				Name     string
				Price    int64
				Quantity int32
			}{
				{ID: "b0000000-0000-0000-0000-000000000010", Name: "General", Price: 9500, Quantity: 600},
				{ID: "b0000000-0000-0000-0000-000000000011", Name: "Floor", Price: 30000, Quantity: 200},
				{ID: "b0000000-0000-0000-0000-000000000012", Name: "Cageside", Price: 80000, Quantity: 30},
			},
		},
	}

	for _, e := range events {
		now := time.Now()
		_, err := pool.Exec(ctx,
			`INSERT INTO events (id, title, description, category, venue, location, date, image_url, status, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'active', $9, $9)
			 ON CONFLICT (id) DO UPDATE SET title = $2, description = $3, category = $4, venue = $5, location = $6, date = $7`,
			e.ID, e.Title, e.Desc, e.Category, e.Venue, e.Location, e.Date, e.ImageURL, now)
		if err != nil {
			log.Printf("Failed to insert event %s: %v", e.Title, err)
			continue
		}

		for _, t := range e.Tiers {
			_, err := pool.Exec(ctx,
				`INSERT INTO ticket_tiers (id, event_id, name, price_cents, total_quantity, available_quantity, version, created_at)
				 VALUES ($1, $2, $3, $4, $5, $5, 1, $6)
				 ON CONFLICT (id) DO UPDATE SET available_quantity = $5, version = 1`,
				t.ID, e.ID, t.Name, t.Price, t.Quantity, now)
			if err != nil {
				log.Printf("Failed to insert tier %s: %v", t.Name, err)
			}
		}

		fmt.Printf("  Seeded: %s (%d tiers)\n", e.Title, len(e.Tiers))
	}

	fmt.Println("Seed complete!")
}
