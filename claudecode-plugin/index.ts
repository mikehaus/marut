// Copyright (c) 2026 Mike Hollingshaus
// Licensed under the MIT License
// See https://github.com/mikehollingshaus/marut/blob/main/LICENSE

#!/usr/bin/env node

/**
 * Marut hook wrapper for Claude Code
 * 
 * This script reads Claude Code hook input from stdin, passes it to the marut
 * binary, and handles the exit codes according to Claude Code's hook contract.
 * 
 * Claude Code hook contract:
 * - Exit 0: Allow the action to proceed
 * - Exit 2: Block the action (shows stderr to Claude)
 * - Other: Non-blocking error (execution continues)
 * 
 * Marut exit codes:
 * - 0: Command allowed
 * - 1: Internal error
 * - 2: Command blocked by policy
 */

import { spawn } from "child_process";
import { createReadStream } from "fs";

const MARUT_BIN = process.env.MARUT_BIN ?? "marut";
const MARUT_ARGS = (
  process.env.MARUT_ARGS ?? "--mode validate --platform claudecode"
).split(/\s+/);
const MARUT_TIMEOUT_MS = 5000;

async function main() {
  // Read JSON input from stdin
  const chunks: Buffer[] = [];
  
  for await (const chunk of process.stdin) {
    chunks.push(chunk);
  }
  
  const input = Buffer.concat(chunks).toString('utf-8');
  
  // Spawn marut with the input
  const proc = spawn(MARUT_BIN, MARUT_ARGS, {
    stdio: ["pipe", "inherit", "inherit"],
  });

  const timeout = setTimeout(() => {
    proc.kill();
    console.error("marut timeout: process did not respond within 5s");
    process.exit(1);
  }, MARUT_TIMEOUT_MS);

  proc.on("close", (code) => {
    clearTimeout(timeout);
    // Pass through marut's exit code directly
    // 0 = allow, 2 = block, 1 = error
    process.exit(code ?? 1);
  });

  proc.on("error", (err) => {
    clearTimeout(timeout);
    console.error(`marut spawn error: ${err.message}`);
    process.exit(1);
  });

  proc.stdin.write(input);
  proc.stdin.end();
}

main().catch((err) => {
  console.error(`Marut hook error: ${err.message}`);
  process.exit(1);
});
