FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /url-shortener ./cmd/app

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /url-shortener /url-shortener
COPY --from=builder /app/web /web

EXPOSE 8080
ENTRYPOINT ["/url-shortener"]
