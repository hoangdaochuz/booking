#!/bin/bash
set -e

MODE=${1:-naive}
USERS=${2:-200}

echo "=== TicketBox Double Booking Load Test ==="
echo "Mode: $MODE"
echo "Concurrent users: $USERS"
echo ""

cd "$(dirname "$0")"
go run load-test.go -mode="$MODE" -users="$USERS" -gateway="http://localhost:8000"
