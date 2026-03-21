.PHONY: build install install-plugin install-global clean test

# Build the marut binary
build:
	go build -o marut ./cmd/marut

# Install marut binary to /usr/local/bin (requires sudo)
install: build
	@echo "Installing marut binary to /usr/local/bin..."
	sudo cp marut /usr/local/bin/marut
	@echo "✓ marut installed"

# Install OpenCode plugin locally (project-level)
install-plugin: build
	@echo "Building OpenCode plugin..."
	cd opencode-plugin && npm install && npm run build
	@echo "Installing plugin to .opencode/plugins/..."
	mkdir -p .opencode/plugins
	cp opencode-plugin/index.js .opencode/plugins/marut.js
	@echo "✓ Plugin installed locally"
	@echo ""
	@echo "Set environment variables in your shell:"
	@echo "  export MARUT_BIN=\"$$(pwd)/marut\""
	@echo "  export MARUT_ARGS=\"--config $$(pwd)/config/default.yaml --log $$(pwd)/audit.log\""

# Install OpenCode plugin globally (all OpenCode sessions)
install-global: build
	@echo "Building OpenCode plugin..."
	cd opencode-plugin && npm install && npm run build
	@echo "Installing plugin to ~/.config/opencode/plugins/..."
	mkdir -p ~/.config/opencode/plugins
	cp opencode-plugin/index.js ~/.config/opencode/plugins/marut.js
	@echo "✓ Plugin installed globally"
	@echo ""
	@echo "Set environment variables in your shell:"
	@echo "  export MARUT_BIN=\"marut\""
	@echo "  export MARUT_ARGS=\"--config /path/to/config.yaml --log /path/to/audit.log\""

# Clean build artifacts
clean:
	rm -f marut
	rm -f audit.log
	cd opencode-plugin && npm run clean

# Run tests
test:
	go test ./...
