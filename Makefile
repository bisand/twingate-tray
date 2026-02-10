.PHONY: build clean install uninstall test help version

# Variables
BINARY_NAME=twingate-tray
GO=go
INSTALL_PATH=/usr/local/bin
CMD_PATH=./cmd/twingate-tray

# Version information from git
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Linker flags to inject version information
LDFLAGS := -X github.com/bisand/twingate-tray/internal/app.Version=$(VERSION) \
           -X github.com/bisand/twingate-tray/internal/app.GitCommit=$(GIT_COMMIT) \
           -X github.com/bisand/twingate-tray/internal/app.BuildDate=$(BUILD_DATE)

help:
	@echo "Twingate Tray - Build Targets"
	@echo ""
	@echo "  make build              Build the binary with version info"
	@echo "  make clean              Remove compiled binary"
	@echo "  make install            Install binary to /usr/local/bin"
	@echo "  make install-icon       Install application icon system-wide"
	@echo "  make uninstall          Remove installed binary"
	@echo "  make test               Run tests (if any)"
	@echo "  make version            Show version information"
	@echo "  make help               Show this help message"
	@echo ""

version:
	@echo "Version:     $(VERSION)"
	@echo "Git Commit:  $(GIT_COMMIT)"
	@echo "Build Date:  $(BUILD_DATE)"

build:
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) $(CMD_PATH)
	@echo "Build complete: $(BINARY_NAME) (version $(VERSION))"

clean:
	$(GO) clean
	rm -f $(BINARY_NAME)
	@echo "Clean complete"

install: build
	sudo cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installed to $(INSTALL_PATH)/$(BINARY_NAME)"

install-icon:
	@echo "Installing application icon..."
	@cd assets && sudo ./install-icon.sh
	@echo "Icon installation complete"

uninstall:
	sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Uninstalled"

test:
	$(GO) test -v ./...

fmt:
	$(GO) fmt ./...

lint:
	$(GO) vet ./...

deps:
	$(GO) mod tidy

all: clean build
	@echo "Build complete"
