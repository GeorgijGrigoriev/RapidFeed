BINARY_NAME := rapidfeed
VERSION     := 1.0.4
SRC         := cmd/main.go
COMMIT := $(shell git rev-parse --short HEAD)

OS_LIST     := linux darwin windows
ARCH_LIST   := amd64

.PHONY: all build bin clean

all: bin build

bin:
	@echo "Building $(BINARY_NAME)-$(VERSION) for current platform..."
	@go build -o $(BINARY_NAME)-$(VERSION) -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT)" $(SRC)

build:
	@echo "Cross‑building for $(OS_LIST)..."
	@for os in $(OS_LIST); do \
        for arch in $(ARCH_LIST); do \
            echo "  → $(BINARY_NAME)-$(VERSION)-$$os-$$arch"; \
            env GOOS=$$os GOARCH=$$arch \
                go build -o $(BINARY_NAME)-$(VERSION)-$$os-$$arch \
                -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT)" $(SRC); \
        done \
    done

docker:
	docker build -t ghcr.io/georgijgrigoriev/rapidfeed:latest --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) -f build/Dockerfile .

clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)-$(VERSION)*

