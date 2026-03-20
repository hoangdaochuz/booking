FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY pkg ./pkg
COPY services/payment ./services/payment

RUN cd services/payment && GOWORK=off go mod download && \
    CGO_ENABLED=0 GOOS=linux GOWORK=off go build -o /bin/payment-service ./cmd

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/payment-service /bin/payment-service
COPY services/payment/migrations /migrations

ENTRYPOINT ["/bin/payment-service"]