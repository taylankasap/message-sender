FROM golang:1.24-alpine AS builder
# required for CGO
RUN apk add --no-cache --update go gcc g++
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# CGO is required for go-sqlite3
RUN CGO_ENABLED=1 go build -o message-sender .

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /app/message-sender ./message-sender
EXPOSE 8080
CMD ["./message-sender"]
