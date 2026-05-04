# Build stage
# confluent-kafka-go requires CGO + librdkafka. We use the musl build tag
# to link against the librdkafka bundled with confluent-kafka-go, so the
# runtime image needs no extra packages.
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache build-base

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -tags musl -o /out/main ./rating/cmd

# Runtime stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /out/main ./
COPY rating/configs/default.yaml ./

EXPOSE 8082

CMD ["/app/main", "-config=default.yaml"]
