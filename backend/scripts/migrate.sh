#!/bin/bash
set -e

SERVICES=("user:5433:ticketbox_user" "event:5434:ticketbox_event" "booking:5435:ticketbox_booking" "notification:5436:ticketbox_notification" "payment:5437:ticketbox_payment" "saga:5438:ticketbox_saga")

for entry in "${SERVICES[@]}"; do
    IFS=: read -r svc port db <<< "$entry"
    echo "Running migrations for $svc service..."
    migrate -path "services/$svc/migrations" \
        -database "postgres://ticketbox:${POSTGRES_PASSWORD:-ticketbox_secret}@localhost:$port/$db?sslmode=disable" \
        up
    echo "$svc migrations complete."
done
