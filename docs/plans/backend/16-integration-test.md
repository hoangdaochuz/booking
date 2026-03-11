# Task 16: Integration Smoke Test

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** End-to-end verification that the entire system works together.

---

### Step 1: Start everything

```bash
cd /Users/dev/work/booking/backend
docker-compose up -d --build
sleep 10
bash scripts/migrate.sh
go run scripts/seed.go
```

### Step 2: Test auth flow

```bash
# Register
curl -s -X POST http://localhost:8000/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123","name":"Test User"}' | jq

# Login and capture token
TOKEN=$(curl -s -X POST http://localhost:8000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}' | jq -r '.access_token')

echo "Token: $TOKEN"
```
Expected: Both return 200 with tokens and user profile.

### Step 3: Test event endpoints

```bash
# List events
curl -s http://localhost:8000/api/events | jq

# Get single event (use ID from list response)
EVENT_ID=$(curl -s http://localhost:8000/api/events | jq -r '.events[0].id')
curl -s http://localhost:8000/api/events/$EVENT_ID | jq
```
Expected: Returns 4 seeded events with ticket tiers.

### Step 4: Test booking flow

```bash
# Get a tier ID
TIER_ID=$(curl -s http://localhost:8000/api/events/$EVENT_ID | jq -r '.tiers[0].id')

# Create booking (pessimistic — safe mode)
curl -s -X POST http://localhost:8000/api/bookings \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Booking-Mode: pessimistic" \
  -d "{\"event_id\":\"$EVENT_ID\",\"items\":[{\"ticket_tier_id\":\"$TIER_ID\",\"quantity\":2}]}" | jq

# List my bookings
curl -s http://localhost:8000/api/bookings \
  -H "Authorization: Bearer $TOKEN" | jq
```
Expected: Booking created with status CONFIRMED.

### Step 5: Test token refresh and logout

```bash
# Refresh (use refresh_token from login response)
REFRESH_TOKEN=$(curl -s -X POST http://localhost:8000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}' | jq -r '.refresh_token')

curl -s -X POST http://localhost:8000/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}" | jq

# Logout
curl -s -X POST http://localhost:8000/api/auth/logout \
  -H "Authorization: Bearer $TOKEN" | jq

# Verify token is blacklisted
curl -s http://localhost:8000/api/users/me \
  -H "Authorization: Bearer $TOKEN" | jq
```
Expected: After logout, the old token returns 401.

### Step 6: Run double booking demo

```bash
# Reset seed data
go run scripts/seed.go

# Naive mode — should oversell
bash scripts/load-test.sh naive 200

# Reset and run safe mode
go run scripts/seed.go
bash scripts/load-test.sh pessimistic 200
```
Expected: Naive mode oversells, pessimistic mode does not.

### Step 7: Check Kafka UI

Open http://localhost:8080 in browser. Verify topics exist and messages are flowing:
- `booking.events`
- `user.events`
- `event.events`

### Step 8: Check notification logs

```bash
docker-compose logs notification-service | grep "NOTIFICATION SENT"
```
Expected: See logged notifications for bookings and user registration.

### Step 9: Final commit

```bash
git add -A
git commit -m "feat(backend): complete TicketBox backend microservice system"
```
