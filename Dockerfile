FROM golang:1.24-alpine AS base

WORKDIR /app

RUN apk add --no-cache make

COPY go.mod go.sum ./

RUN go mod download

FROM base AS development

RUN go install github.com/air-verse/air@latest 

COPY . .

CMD ["air", "-c", ".air.toml"]

FROM base AS builder

COPY . .

RUN CGO_ENABLED=0 go build -o /app/importer -ldflags="-s -w" ./cmd/app/main.go

FROM alpine:latest AS production

WORKDIR /app

COPY --from=builder /app/importer /usr/local/bin/importer

RUN chmod +x /usr/local/bin/importer

CMD ["/usr/local/bin/importer"]
