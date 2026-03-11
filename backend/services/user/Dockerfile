FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY pkg/ ./pkg/
COPY services/user/ ./services/user/

RUN cd services/user && GOWORK=off go mod download && \
    CGO_ENABLED=0 GOOS=linux GOWORK=off go build -o /bin/user-service ./cmd

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/user-service /bin/user-service
COPY services/user/migrations /migrations

ENTRYPOINT ["/bin/user-service"]
