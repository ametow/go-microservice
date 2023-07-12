FROM golang:1.20-alpine AS builder

WORKDIR /usr/src/app

ENV CGO_ENABLED 0

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src .
RUN mkdir -p /usr/local/bin/
RUN go mod tidy
RUN go build -ldflags="-s -w" -o /usr/local/bin/app cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

RUN mkdir -p /app/config /app/migrations /usr/local/bin

COPY --from=builder /usr/local/bin/app /usr/local/bin/app
COPY --from=builder /usr/src/app/config /app/config
COPY --from=builder /usr/src/app/migrations /app/migrations

EXPOSE 8080

CMD ["app", "docker"]
