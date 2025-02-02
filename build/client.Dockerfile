FROM golang:1.23 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /app/cmd/client
RUN CGO_ENABLED=0 GOOS=linux go build -a -o client .

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /app/cmd/client/client .
CMD ["./client"]
