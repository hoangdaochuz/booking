FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY pkg/ ./pkg/
COPY services/notification/ ./services/notification/

RUN cd services/notification && GOWORK=off go mod download && \
    CGO_ENABLED=0 GOOS=linux GOWORK=off go build -o /bin/notification-service ./cmd

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/notification-service /bin/notification-service
COPY services/notification/migrations /migrations

ENTRYPOINT ["/bin/notification-service"]
