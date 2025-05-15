FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o flipt-go-htmx-app

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/flipt-go-htmx-app .

EXPOSE 8081

CMD ["./flipt-go-htmx-app"]
