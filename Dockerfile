FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/marketplace-api ./cmd/marketplace-api

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/marketplace-api /app/
COPY ./migrations /app/migrations
COPY ./pkg /app/pkg

EXPOSE ${GRPC_PORT}
EXPOSE ${GATEWAY_PORT}

CMD ["/app/marketplace-api"]
