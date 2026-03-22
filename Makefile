.PHONY: build install \
	install-opencode-local install-opencode-global \
	install-claude-global install-claude-local \
	uninstall-opencode-local uninstall-opencode-global \
	uninstall-claude-global uninstall-claude-local \
	build-claude clean test

# Optional target directory for local installs (defaults to current directory)
DIR ?= $(CURDIR)

# Build the marut binary
build:
	go build -o marut ./cmd/marut

# Install marut binary to /usr/local/bin (requires sudo)
install: build
	@echo "Installing marut binary to /usr/local/bin..."
	sudo cp marut /usr/local/bin/marut
	@echo "✓ marut installed"

# ---------------------------------------------------------------------------
# OpenCode
# ---------------------------------------------------------------------------

# Install OpenCode plugin locally (project-level, defaults to current dir, override with DIR=)
install-opencode-local: build
	@echo "Building OpenCode plugin..."
	cd opencode-plugin && npm install && npm run build
	@echo "Installing plugin to $(DIR)/.opencode/plugins/..."
	mkdir -p $(DIR)/.opencode/plugins
	cp opencode-plugin/index.js $(DIR)/.opencode/plugins/marut.js
	@echo "✓ Plugin installed to $(DIR)"
	@echo ""
	@echo "Make sure these are in your ~/.zshrc:"
	@echo "  export MARUT_BIN=\"$$(pwd)/marut\""
	@echo "  export MARUT_CONFIG=\"$$(pwd)/config/default.yaml\""
	@echo "  export MARUT_LOG=\"$$(pwd)/audit.log\""
	@echo "  # Optional: export MARUT_ARGS=\"--sim --agent-id myagent\""

# Install OpenCode plugin globally (all OpenCode sessions)
install-opencode-global: build
	@echo "Building OpenCode plugin..."
	cd opencode-plugin && npm install && npm run build
	@echo "Installing plugin to ~/.config/opencode/plugins/..."
	mkdir -p ~/.config/opencode/plugins
	cp opencode-plugin/index.js ~/.config/opencode/plugins/marut.js
	@echo "✓ Plugin installed globally"
	@echo ""
	@echo "Make sure these are in your ~/.zshrc:"
	@echo "  export MARUT_BIN=\"$$(pwd)/marut\""
	@echo "  export MARUT_CONFIG=\"$$(pwd)/config/default.yaml\""
	@echo "  export MARUT_LOG=\"$$(pwd)/audit.log\""
	@echo "  # Optional: export MARUT_ARGS=\"--sim --agent-id myagent\""

# Uninstall OpenCode plugin locally (defaults to current dir, override with DIR=)
uninstall-opencode-local:
	@echo "Removing $(DIR)/.opencode/plugins/marut.js..."
	rm -f $(DIR)/.opencode/plugins/marut.js
	@echo "✓ Done"

# Uninstall OpenCode plugin globally
uninstall-opencode-global:
	@echo "Removing ~/.config/opencode/plugins/marut.js..."
	rm -f ~/.config/opencode/plugins/marut.js
	@echo "✓ Done"

# ---------------------------------------------------------------------------
# Claude Code
# ---------------------------------------------------------------------------

# Build Claude Code plugin
build-claude: build
	@echo "Building Claude Code plugin..."
	cd claudecode-plugin && npm install && npm run build
	@echo "✓ Claude Code plugin built"

# Install Claude Code hook globally (~/.claude/settings.json)
install-claude-global: build-claude
	@echo "Installing Claude Code hook to ~/.claude/settings.json..."
	@WRAPPER="$$(pwd)/claudecode-plugin/marut-wrapper.sh"; \
	chmod +x "$$WRAPPER"; \
	mkdir -p ~/.claude; \
	printf 'import json, os\npath = os.path.expanduser("~/.claude/settings.json")\ns = json.load(open(path)) if os.path.exists(path) else {}\nhook = {"type": "command", "command": "%s"}\nentry = {"matcher": "*", "hooks": [hook]}\nexisting = s.setdefault("hooks", {}).setdefault("PreToolUse", [])\nif not any("marut-wrapper.sh" in str(e) for e in existing):\n    existing.append(entry)\njson.dump(s, open(path, "w"), indent=2)\nprint("Hook added to ~/.claude/settings.json")\n' "$$WRAPPER" > /tmp/marut_install.py; \
	python3 /tmp/marut_install.py && echo "✓ Done" || { echo "✘ Failed to update settings.json"; exit 1; }; \
	rm -f /tmp/marut_install.py
	@echo ""
	@echo "Make sure these are in your ~/.zshrc:"
	@echo "  export MARUT_BIN=\"$$(pwd)/marut\""
	@echo "  export MARUT_CONFIG=\"$$(pwd)/config/default.yaml\""
	@echo "  export MARUT_LOG=\"$$(pwd)/audit.log\""

# Install Claude Code hook locally (project-level, defaults to current dir, override with DIR=)
install-claude-local: build-claude
	@echo "Installing Claude Code hook to $(DIR)/.claude/settings.json..."
	@WRAPPER="$$(pwd)/claudecode-plugin/marut-wrapper.sh"; \
	chmod +x "$$WRAPPER"; \
	mkdir -p $(DIR)/.claude; \
	printf 'import json, os\npath = os.path.expanduser("%s/.claude/settings.json")\ns = json.load(open(path)) if os.path.exists(path) else {}\nhook = {"type": "command", "command": "%s"}\nentry = {"matcher": "*", "hooks": [hook]}\nexisting = s.setdefault("hooks", {}).setdefault("PreToolUse", [])\nif not any("marut-wrapper.sh" in str(e) for e in existing):\n    existing.append(entry)\njson.dump(s, open(path, "w"), indent=2)\nprint("Hook added to %s/.claude/settings.json")\n' "$(DIR)" "$$WRAPPER" "$(DIR)" > /tmp/marut_install.py; \
	python3 /tmp/marut_install.py && echo "✓ Done" || { echo "✘ Failed to update settings.json"; exit 1; }; \
	rm -f /tmp/marut_install.py
	@echo ""
	@echo "Make sure these are in your ~/.zshrc:"
	@echo "  export MARUT_BIN=\"$$(pwd)/marut\""
	@echo "  export MARUT_CONFIG=\"$$(pwd)/config/default.yaml\""
	@echo "  export MARUT_LOG=\"$$(pwd)/audit.log\""

# Uninstall Claude Code hook globally (~/.claude/settings.json)
uninstall-claude-global:
	@echo "Removing marut hook from ~/.claude/settings.json..."
	@printf 'import json, os\npath = os.path.expanduser("~/.claude/settings.json")\nif not os.path.exists(path):\n    print("Nothing to uninstall")\n    exit(0)\ns = json.load(open(path))\nhooks = s.get("hooks", {}).get("PreToolUse", [])\ns.setdefault("hooks", {})["PreToolUse"] = [e for e in hooks if "marut-wrapper.sh" not in str(e)]\nif not s["hooks"]["PreToolUse"]:\n    del s["hooks"]["PreToolUse"]\nif not s["hooks"]:\n    del s["hooks"]\njson.dump(s, open(path, "w"), indent=2)\nprint("Done")\n' > /tmp/marut_uninstall.py; \
	python3 /tmp/marut_uninstall.py && echo "✓ Done" || { echo "✘ Failed"; exit 1; }; \
	rm -f /tmp/marut_uninstall.py

# Uninstall Claude Code hook locally (defaults to current dir, override with DIR=)
uninstall-claude-local:
	@echo "Removing marut hook from $(DIR)/.claude/settings.json..."
	@printf 'import json, os\npath = os.path.expanduser("%s/.claude/settings.json")\nif not os.path.exists(path):\n    print("Nothing to uninstall")\n    exit(0)\ns = json.load(open(path))\nhooks = s.get("hooks", {}).get("PreToolUse", [])\ns.setdefault("hooks", {})["PreToolUse"] = [e for e in hooks if "marut-wrapper.sh" not in str(e)]\nif not s["hooks"]["PreToolUse"]:\n    del s["hooks"]["PreToolUse"]\nif not s["hooks"]:\n    del s["hooks"]\njson.dump(s, open(path, "w"), indent=2)\nprint("Done")\n' "$(DIR)" > /tmp/marut_uninstall.py; \
	python3 /tmp/marut_uninstall.py && echo "✓ Done" || { echo "✘ Failed"; exit 1; }; \
	rm -f /tmp/marut_uninstall.py

# ---------------------------------------------------------------------------
# Misc
# ---------------------------------------------------------------------------

# Clean build artifacts
clean:
	rm -f marut
	rm -f audit.log
	cd opencode-plugin && npm run clean
	cd claudecode-plugin && npm run clean

# Run tests
test:
	go test ./...
