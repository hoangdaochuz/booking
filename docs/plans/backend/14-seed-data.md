# Task 14: Seed Data

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create a seed script that populates the Event Service database with the same events from the frontend mock data.

**Files:**
- Create: `backend/scripts/seed.go`

---

### Step 1: Create seed script

`backend/scripts/seed.go`:

The script connects directly to `postgres-event` and inserts:

```go
// Events to seed (matching frontend lib/mock-data.ts):
//
// 1. The Weeknd — After Hours World Tour
//    Category: Concert, Venue: MSG, Location: New York, Date: Jul 15 2026
//    Tiers: General ($130, 500), VIP ($350, 200), Premium ($550, 50)
//
// 2. Billie Eilish — Hit Me Hard and Soft Tour
//    Category: Concert, Venue: O2 Arena, Location: London, Date: Apr 8 2026
//    Tiers: General ($85, 800), VIP ($250, 300), Premium ($450, 100)
//
// 3. NBA Finals 2026 — Game 3
//    Category: Sports, Venue: Chase Center, Location: San Francisco, Date: Jun 12 2026
//    Tiers: Upper Level ($120, 1000), Lower Level ($350, 400), Courtside ($1200, 50)
//
// 4. UFC 310 — Championship
//    Category: Sports, Venue: T-Mobile Arena, Location: Las Vegas, Date: May 3 2026
//    Tiers: General ($95, 600), Floor ($300, 200), Cageside ($800, 30)
```

Key implementation details:
- Uses `pgx` to connect directly to the event database
- Inserts events with fixed UUIDs (deterministic for testing)
- Sets `available_quantity = total_quantity` for fresh state
- Prices stored in cents (e.g., $130 = 13000)

### Step 2: Run seed

```bash
cd /Users/dev/work/booking/backend
docker-compose up -d
go run scripts/seed.go
```

### Step 3: Verify via API

```bash
curl http://localhost:8000/api/events | jq
```
Expected: Returns 4 events with their ticket tiers.

### Step 4: Commit

```bash
git add backend/scripts/seed.go
git commit -m "feat(backend): add seed script with mock event data"
```
