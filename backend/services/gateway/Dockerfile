FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY pkg/ ./pkg/
COPY services/gateway/ ./services/gateway/

RUN cd services/gateway && GOWORK=off go mod download && \
    CGO_ENABLED=0 GOOS=linux GOWORK=off go build -o /bin/gateway ./cmd

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/gateway /bin/gateway

ENTRYPOINT ["/bin/gateway"]
