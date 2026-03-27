FROM golang:1.25.8-bookworm

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o rapidfeed cmd/main.go

CMD ["./rapidfeed"]
