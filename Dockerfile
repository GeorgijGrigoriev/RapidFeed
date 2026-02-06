FROM golang:1.25

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o rapidfeed cmd/main.go

EXPOSE 8000

CMD ["./rapidfeed"]

