package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/ticketbox/event/internal/domain"
	"github.com/ticketbox/event/internal/repository"
	"github.com/ticketbox/pkg/config"
	"github.com/ticketbox/pkg/database"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	ctx := context.Background()
	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL, logger)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()

	seatRepo := repository.NewPostgresSeatRepository(pool)
	eventRepo := repository.NewPostgresEventRepository(pool)

	// Get all events
	events, _, err := eventRepo.List(ctx, "", "", 1, 10)
	if err != nil {
		log.Fatal("Failed to list events:", err)
	}

	fmt.Printf("Found %d events\n", len(events))

	for _, event := range events {
		fmt.Printf("\nProcessing event: %s (category: %s)\n", event.Title, event.Category)

		// Delete existing seats for this event to regenerate
		existingSeats, err := seatRepo.GetByEventID(ctx, event.ID, nil)
		if err == nil && len(existingSeats) > 0 {
			fmt.Printf("  - Deleting %d existing seats\n", len(existingSeats))
			if err := seatRepo.DeleteByEventID(ctx, event.ID); err != nil {
				log.Printf("Failed to delete seats for event %s: %v", event.ID, err)
				continue
			}
		}

		// Generate seats for each tier based on event category
		var allSeats []*domain.Seat

		// Sort tiers by price (most expensive first) to assign sections correctly
		sortedTiers := make([]domain.TicketTier, len(event.Tiers))
		copy(sortedTiers, event.Tiers)

		// Simple sort by price (most expensive first)
		for i := 0; i < len(sortedTiers); i++ {
			for j := i + 1; j < len(sortedTiers); j++ {
				if sortedTiers[i].PriceCents < sortedTiers[j].PriceCents {
					sortedTiers[i], sortedTiers[j] = sortedTiers[j], sortedTiers[i]
				}
			}
		}

		for idx, tier := range sortedTiers {
			sections := getSectionsForTier(idx, event.Category)
			fmt.Printf("  - Tier %d: %s ($%d) -> sections: %v\n", idx, tier.Name, tier.PriceCents/100, sections)

			// Distribute seats across sections for this tier
			seatsPerSection := int(tier.TotalQuantity) / len(sections)
			if seatsPerSection < 10 {
				seatsPerSection = 10 // Minimum 10 seats per section
			}

			for sectionIdx, sectionID := range sections {
				// Generate section name combining tier and section
				sectionName := fmt.Sprintf("%s-%s", tier.Name, sectionID)

				seats := generateSeatsForSection(event.ID, tier.ID, sectionName, seatsPerSection, idx, sectionIdx)
				allSeats = append(allSeats, seats...)
				fmt.Printf("    - Section %s: %d seats\n", sectionName, len(seats))
			}
		}

		if len(allSeats) > 0 {
			if err := seatRepo.CreateBatch(ctx, allSeats); err != nil {
				log.Printf("Failed to create seats for event %s: %v", event.ID, err)
				continue
			}
			fmt.Printf("  ✓ Created %d total seats\n", len(allSeats))
		}
	}

	fmt.Println("\nDone!")
}

// getSectionsForTier returns the section IDs for a tier based on its price rank (0 = most expensive)
func getSectionsForTier(tierRank int, category string) []string {
	category = strings.ToLower(category)

	if category == "sports" {
		// Sports layout templates
		switch tierRank {
		case 0: // Most expensive - courtside
			return []string{"north", "south"}
		case 1: // Medium - lower bowl
			return []string{"lower-north", "lower-south", "lower-west", "lower-east"}
		case 2: // Cheapest - upper deck
			return []string{"upper-north", "upper-south"}
		default:
			return []string{fmt.Sprintf("section-%d", tierRank)}
		}
	}

	// Concert layout templates (default)
	switch tierRank {
	case 0: // Most expensive - center front
		return []string{"center"}
	case 1: // Medium - sides
		return []string{"left", "right"}
	case 2: // Cheapest - back
		return []string{"back-center", "back-left", "back-right"}
	default:
		return []string{fmt.Sprintf("section-%d", tierRank)}
	}
}

// generateSeatsForSection generates seats for a specific section
func generateSeatsForSection(eventID, tierID uuid.UUID, sectionID string, totalSeats int, tierRank, sectionIdx int) []*domain.Seat {
	seats := make([]*domain.Seat, 0, totalSeats)

	// Reduced seat count: max 8 seats per row, max 4 rows per section
	rows := (totalSeats + 7) / 8
	if rows > 4 {
		rows = 4
	}
	if rows < 1 {
		rows = 1
	}

	seatsPerRow := totalSeats / rows
	if seatsPerRow > 8 {
		seatsPerRow = 8
	}
	if seatsPerRow < 1 {
		seatsPerRow = 1
	}

	// Adjust last row if needed
	remainingSeats := totalSeats

	for row := 0; row < rows && remainingSeats > 0; row++ {
		rowLabel := string(rune('A' + row))
		seatsInRow := seatsPerRow
		if row == rows-1 {
			seatsInRow = remainingSeats
		}

		for i := 0; i < seatsInRow; i++ {
			seats = append(seats, &domain.Seat{
				ID:           uuid.New(),
				EventID:      eventID,
				TicketTierID: tierID,
				Status:       domain.SeatStatusAvailable,
				BookingID:    nil,
				OrderID:      nil,
				Position: domain.Position{
					SectionID: sectionID,
					Row:       rowLabel,
					Seat:      i + 1,
					X:         0, // Not used - SVG templates define positions
					Y:         0,
				},
			})
		}
		remainingSeats -= seatsInRow
	}

	return seats
}
