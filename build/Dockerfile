# Build stage
FROM golang:1.25-alpine AS builder

ARG VERSION=latest
ARG COMMIT=ffffff

# Set build arguments as labels for image metadata
LABEL org.opencontainers.image.version=${VERSION}
LABEL org.opencontainers.image.revision=${COMMIT}
LABEL org.opencontainers.image.ref.name="RapidFeed"
LABEL org.opencontainers.image.authors="gg <grigorievga@yandex.ru>"
# For cgo support, for sqlite3
RUN apk add --no-cache git make build-base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -ldflags "-X main.Version=${VERSION} -X main.Commit=${COMMIT}" -o rapidfeed cmd/main.go

# Runtime stage
FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/rapidfeed .

RUN chmod +x /app/rapidfeed

CMD ["/app/rapidfeed"]