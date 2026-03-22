#!/bin/bash
# Copyright (c) 2026 Mike Hollingshaus
# Licensed under the MIT License
# See https://github.com/mikehollingshaus/marut/blob/main/LICENSE

# Wrapper script for marut that sets paths explicitly
# This ensures the hook works even if environment variables aren't passed through

MARUT_BIN="${MARUT_BIN:-$HOME/Developer/projects/go/marut/marut}"
MARUT_CONFIG="${MARUT_CONFIG:-$HOME/Developer/projects/go/marut/config/default.yaml}"
MARUT_LOG="${MARUT_LOG:-$HOME/Developer/projects/go/marut/audit.log}"

# MARUT_ARGS is an optional escape hatch for extra flags (e.g. --sim, --agent-id).
# Keep to short flags only — no paths with spaces.
# shellcheck disable=SC2086
exec "$MARUT_BIN" --mode validate --platform claudecode --config "$MARUT_CONFIG" --log "$MARUT_LOG" $MARUT_ARGS
