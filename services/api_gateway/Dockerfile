FROM golang:1.25-alpine3.22 AS builder

RUN apk add --no-cache ca-certificates tzdata protobuf protobuf-dev

RUN go install github.com/swaggo/swag/cmd/swag@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN swag init -g services/api_gateway/cmd/server/main.go \
              -o services/api_gateway/docs \
              --parseDependency

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server ./services/api_gateway/cmd/server/main.go

FROM alpine:3.20

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder --chown=appuser:appgroup --chmod=755 /server /server

USER appuser

EXPOSE 8080

CMD ["/server"]
