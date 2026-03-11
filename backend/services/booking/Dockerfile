FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY pkg/ ./pkg/
COPY services/booking/ ./services/booking/

RUN cd services/booking && GOWORK=off go mod download && \
    CGO_ENABLED=0 GOOS=linux GOWORK=off go build -o /bin/booking-service ./cmd

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/booking-service /bin/booking-service
COPY services/booking/migrations /migrations

ENTRYPOINT ["/bin/booking-service"]
