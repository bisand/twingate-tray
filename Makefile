.PHONY: build clean install uninstall test help

# Variables
BINARY_NAME=twingate-tray
GO=go
INSTALL_PATH=/usr/local/bin
CMD_PATH=./cmd/twingate-tray

help:
	@echo "Twingate Tray - Build Targets"
	@echo ""
	@echo "  make build              Build the binary"
	@echo "  make clean              Remove compiled binary"
	@echo "  make install            Install binary to /usr/local/bin"
	@echo "  make uninstall          Remove installed binary"
	@echo "  make test               Run tests (if any)"
	@echo "  make help               Show this help message"
	@echo ""

build:
	$(GO) build -o $(BINARY_NAME) $(CMD_PATH)
	@echo "Build complete: $(BINARY_NAME)"

clean:
	$(GO) clean
	rm -f $(BINARY_NAME)
	@echo "Clean complete"

install: build
	sudo cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installed to $(INSTALL_PATH)/$(BINARY_NAME)"

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
