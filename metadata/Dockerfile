# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/main ./metadata/cmd

# Runtime stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /out/main ./
COPY metadata/configs/default.yaml ./

EXPOSE 8081

CMD ["/app/main", "-config=default.yaml"]
