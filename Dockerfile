# -------- Stage 1: Build --------
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o urlshortener .

# -------- Stage 2: Minimal runtime --------
FROM alpine:edge

WORKDIR /app

COPY --from=builder /app/urlshortener .

EXPOSE 8080
ENTRYPOINT ["/app/urlshortener"]