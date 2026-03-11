#!/usr/bin/env bash
set -euo pipefail

API_URL="${API_URL:-http://localhost:8000}"

echo "==> Registering admin user..."
REGISTER_RESP=$(curl -s -X POST "$API_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@ticketbox.com","password":"admin123","name":"Admin User"}')

echo "$REGISTER_RESP" | jq .

# Promote to admin
echo "==> Promoting to admin..."
docker exec backend-postgres-user-1 psql -U ticketbox -d ticketbox_user -c \
  "UPDATE users SET role='admin' WHERE email='admin@ticketbox.com';"

# Re-login to get admin token
echo "==> Logging in as admin..."
LOGIN_RESP=$(curl -s -X POST "$API_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@ticketbox.com","password":"admin123"}')

TOKEN=$(echo "$LOGIN_RESP" | jq -r '.access_token')

echo "==> Creating events..."

curl -s -X POST "$API_URL/api/events" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Neon Lights World Tour — Live in NYC",
    "description": "Experience the electrifying Neon Lights World Tour at Madison Square Garden.",
    "category": "concerts",
    "venue": "Madison Square Garden",
    "location": "New York, NY",
    "date": "2026-06-15T20:00:00Z",
    "image_url": "https://images.unsplash.com/photo-1459749411175-04bf5292ceea",
    "tiers": [
      {"name": "Floor", "price_cents": 35000, "total_quantity": 200},
      {"name": "Lower Bowl", "price_cents": 18500, "total_quantity": 1000},
      {"name": "Upper Deck", "price_cents": 8500, "total_quantity": 3000}
    ]
  }' | jq .

curl -s -X POST "$API_URL/api/events" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Champions League Final 2026",
    "description": "The biggest match in European football. Two titans clash for glory.",
    "category": "sports",
    "venue": "Wembley Stadium",
    "location": "London, UK",
    "date": "2026-05-30T21:00:00Z",
    "image_url": "https://images.unsplash.com/photo-1489944440615-453fc2b6a9a9",
    "tiers": [
      {"name": "Hospitality", "price_cents": 50000, "total_quantity": 100},
      {"name": "Category 1", "price_cents": 25000, "total_quantity": 5000},
      {"name": "Category 2", "price_cents": 15000, "total_quantity": 10000},
      {"name": "Category 3", "price_cents": 7500, "total_quantity": 20000}
    ]
  }' | jq .

curl -s -X POST "$API_URL/api/events" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Premiere — Starfall: A Space Odyssey",
    "description": "Be among the first to witness the most anticipated sci-fi epic of the decade.",
    "category": "films",
    "venue": "TCL Chinese Theatre",
    "location": "Los Angeles, CA",
    "date": "2026-08-20T19:30:00Z",
    "image_url": "https://images.unsplash.com/photo-1478720568477-152d9b164e26",
    "tiers": [
      {"name": "Red Carpet", "price_cents": 50000, "total_quantity": 50},
      {"name": "Premium", "price_cents": 15000, "total_quantity": 200},
      {"name": "Standard", "price_cents": 5000, "total_quantity": 500}
    ]
  }' | jq .

echo "==> Registering test user..."
curl -s -X POST "$API_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"user123","name":"Jane Smith"}' | jq .

echo "==> Done! Seed data created."
echo "    Admin: admin@ticketbox.com / admin123"
echo "    User:  user@example.com / user123"
