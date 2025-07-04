FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .


COPY go.mod ./
RUN go mod tidy
RUN go mod download

RUN go build -o app
FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]