.PHONY: build install clean help

# Default target
all: build

# Build bitshare executable
build:
	@echo "Building BitShare..."
	@go build -o bitshare .

# Install bitshare
install: build
	@echo "Installing BitShare..."
	@if [ "$(OS)" = "Windows_NT" ]; then \
		powershell -ExecutionPolicy Bypass -File install.ps1; \
	else \
		bash install.sh; \
	fi

# Clean up
clean:
	@echo "Cleaning up..."
	@rm -f bitshare
	@rm -f bitshare.exe

# Display help
help:
	@echo "BitShare Makefile Commands:"
	@echo "  make build    - Build the BitShare executable"
	@echo "  make install  - Build and install BitShare"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make help     - Show this help message"
