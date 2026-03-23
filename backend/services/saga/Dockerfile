FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY pkg ./pkg
COPY services/saga ./services/saga

RUN cd services/saga && GOWORK=off go mod download && \
    CGO_ENABLED=0 GOOS=linux GOWORK=off go build -o /bin/saga-service ./cmd

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/saga-service /bin/saga-service
COPY services/saga/migrations /migrations

ENTRYPOINT ["/bin/saga-service"]