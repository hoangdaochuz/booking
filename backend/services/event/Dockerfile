FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY pkg/ ./pkg/
COPY services/event/ ./services/event/

RUN cd services/event && GOWORK=off go mod download && \
    CGO_ENABLED=0 GOOS=linux GOWORK=off go build -o /bin/event-service ./cmd

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/event-service /bin/event-service
COPY services/event/migrations /migrations

ENTRYPOINT ["/bin/event-service"]
