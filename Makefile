BINARY_NAME=rapidfeed
VERSION=1.0.0

SRC=cmd/main.go

OS=linux darwin windows
ARCH=amd64

build:
	@for os in $(OS); do \
        export GOOS=$$os; \
        go build -o ${BINARY_NAME}-${VERSION}-$${GOOS}-${ARCH} $$SRC; \
    done
bin:
	go build -o rapidfeed cmd/main.go