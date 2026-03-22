#!/bin/bash
# Wrapper script for marut that sets paths explicitly
# This ensures the hook works even if environment variables aren't passed through

MARUT_BIN="${MARUT_BIN:-$HOME/Developer/projects/go/marut/marut}"
MARUT_CONFIG="${MARUT_CONFIG:-$HOME/Developer/projects/go/marut/config/default.yaml}"
MARUT_LOG="${MARUT_LOG:-$HOME/Developer/projects/go/marut/audit.log}"

# Pass stdin through to marut with the correct arguments
exec "$MARUT_BIN" --mode validate --platform claudecode --config "$MARUT_CONFIG" --log "$MARUT_LOG"
