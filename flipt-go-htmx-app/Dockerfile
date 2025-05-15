FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o flipt-go-htmx-app

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/flipt-go-htmx-app .

EXPOSE 8080

CMD ["./flipt-go-htmx-app"]
