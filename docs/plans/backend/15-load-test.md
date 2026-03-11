# Task 15: Load Test — Double Booking Demo

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create a Go-based load test that demonstrates the double booking problem by firing concurrent booking requests.

**Files:**
- Create: `backend/scripts/load-test.go`
- Create: `backend/scripts/load-test.sh`

---

### Step 1: Create load test script

`backend/scripts/load-test.go`:

```go
// Usage:
//   go run scripts/load-test.go -mode=naive -users=200 -tier=<tier-id> -gateway=http://localhost:8000
//   go run scripts/load-test.go -mode=pessimistic -users=200
//   go run scripts/load-test.go -mode=optimistic -users=200
//
// What it does:
//   1. Registers N test users (test-user-001@load.test, etc.)
//   2. Picks an event tier with limited tickets (e.g., 50 Premium tickets)
//   3. Fires N concurrent booking requests using sync.WaitGroup + goroutines
//   4. Each request includes X-Booking-Mode header
//   5. Collects results: confirmed count, failed count, final available count
//   6. Prints summary report
//
// Expected results:
//   naive:       ~200 confirmed (OVERSOLD! only 50 available) — available goes negative
//   pessimistic: exactly 50 confirmed, 150 failed — available = 0
//   optimistic:  exactly 50 confirmed, 150 failed (with retries) — available = 0
```

Key implementation:
- Use `net/http` with connection pooling
- `sync.WaitGroup` to coordinate goroutines
- `sync.Mutex` protected counters for results
- Query final `available_quantity` from Event Service after all requests complete
- Print formatted table with results

### Step 2: Create shell wrapper

`backend/scripts/load-test.sh`:
```bash
#!/bin/bash
set -e

MODE=${1:-naive}
USERS=${2:-200}

echo "=== TicketBox Double Booking Load Test ==="
echo "Mode: $MODE"
echo "Concurrent users: $USERS"
echo ""

cd "$(dirname "$0")/.."
go run scripts/load-test.go -mode="$MODE" -users="$USERS" -gateway="http://localhost:8000"
```

### Step 3: Run the test

```bash
cd /Users/dev/work/booking/backend
chmod +x scripts/load-test.sh

# First reseed data to reset availability
go run scripts/seed.go

# Test naive mode (should show double booking!)
bash scripts/load-test.sh naive 200

# Reseed, then test pessimistic
go run scripts/seed.go
bash scripts/load-test.sh pessimistic 200

# Reseed, then test optimistic
go run scripts/seed.go
bash scripts/load-test.sh optimistic 200
```

Expected output example:
```
=== Results ===
Mode:              naive
Total requests:    200
Confirmed:         187    ← OVERSOLD (only 50 tickets!)
Failed:            13
Available after:   -137   ← NEGATIVE — double booking proven!

=== Results ===
Mode:              pessimistic
Total requests:    200
Confirmed:         50     ← CORRECT
Failed:            150
Available after:   0      ← CORRECT
```

### Step 4: Commit

```bash
git add backend/scripts/load-test.go backend/scripts/load-test.sh
git commit -m "feat(backend): add load test script for double booking demo"
```
