FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o api-docs-portal ./cmd/server/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/api-docs-portal .

EXPOSE 8080
VOLUME ["/app/data"]

ENV DB_PATH=/app/data/portal.db

ENTRYPOINT ["./api-docs-portal"]
