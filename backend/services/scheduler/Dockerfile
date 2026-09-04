FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY pkg ./pkg
COPY services/scheduler ./services/scheduler

RUN cd services/scheduler && GOWORK=off go mod download && \
    CGO_ENABLED=0 GOOS=linux GOWORK=off go build -o /bin/scheduler-service ./cmd

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/scheduler-service /bin/scheduler-service
COPY services/scheduler/migrations /migrations

ENTRYPOINT ["/bin/scheduler-service"]