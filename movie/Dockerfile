# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/main ./movie/cmd

# Runtime stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /out/main ./
COPY movie/configs/default.yaml ./

EXPOSE 8083

CMD ["/app/main", "-config=default.yaml"]
