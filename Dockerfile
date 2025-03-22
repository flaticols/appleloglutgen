FROM golang:1.24.1-alpine3.21 AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o lutgen .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/lutgen .
ENTRYPOINT ["./lutgen"]
