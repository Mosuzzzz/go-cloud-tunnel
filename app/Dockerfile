#STATE 1: Build
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod init url-shortener || true
RUN go build -o main .

#STATE 2: RUN
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
# Verify that the run permissions file is correct.
RUN chmod +x main

EXPOSE 8080
CMD ["./main"]

