FROM golang:1.25-alpine3.22 AS builder

RUN apk add --no-cache ca-certificates tzdata protobuf protobuf-dev

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN protoc --proto_path=. \
    --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/splitter/v1/splitter.proto && \
    mkdir -p internal/pb/splitter/v1 && \
    mv api/proto/splitter/v1/*.go internal/pb/splitter/v1/

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server ./cmd/server/main.go

FROM alpine:3.20

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder --chown=appuser:appgroup --chmod=755 /server /server

USER appuser

EXPOSE 50051

CMD ["/server"]
