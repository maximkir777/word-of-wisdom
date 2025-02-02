FROM golang:1.23 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /app/cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -a -o server .

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /app/cmd/server/server .
EXPOSE 9000
CMD ["./server"]
