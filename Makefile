BINARY_NAME := rapidfeed
VERSION     := 1.0.6
SRC         := cmd/main.go
COMMIT := $(shell git rev-parse --short HEAD)

OS_LIST     := linux darwin
ARCH_LIST   := amd64 arm64

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

docker-latest:
	docker build --platform=linux/amd64 -t ghcr.io/georgijgrigoriev/rapidfeed:latest --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) -f build/Dockerfile .

docker-build-version: docker-amd64 docker-arm64

docker-arm64:
	docker build --platform=linux/arm64 -t ghcr.io/georgijgrigoriev/rapidfeed:$(VERSION)-arm64 --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) -f build/Dockerfile .

docker-amd64:
	docker build --platform=linux/amd64 -t ghcr.io/georgijgrigoriev/rapidfeed:$(VERSION)-amd64 --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) -f build/Dockerfile .

docker-push-arm64:
	docker push ghcr.io/georgijgrigoriev/rapidfeed:$(VERSION)-arm64

docker-push-amd64:
	docker push ghcr.io/georgijgrigoriev/rapidfeed:$(VERSION)-amd64

docker-push-all: docker-push-arm64 docker-push-amd64

clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)-$(VERSION)*
