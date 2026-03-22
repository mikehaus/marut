// Copyright (c) 2026 Mike Hollingshaus
// Licensed under the MIT License
// See https://github.com/mikehollingshaus/marut/blob/main/LICENSE

import { spawn } from "child_process";

const MARUT_BIN = process.env.MARUT_BIN ?? "marut";
// MARUT_ARGS is split on whitespace. Do NOT use quoted arguments with spaces
// (e.g. --config "/path with spaces/file.yaml" will break). Use paths without
// spaces or set --config at invocation time instead of via env var.
const MARUT_ARGS = (
  process.env.MARUT_ARGS ?? "--mode validate --platform opencode"
).split(/\s+/);

const MARUT_TIMEOUT_MS = 5000; // kill marut if it doesn't respond within 5s

function invokeMarut(payload: unknown): Promise<number> {
  return new Promise((resolve) => {
    const proc = spawn(MARUT_BIN, MARUT_ARGS, {
      stdio: ["pipe", "inherit", "inherit"],
    });

    // Timeout guard: if marut hangs, kill it and treat as internal error.
    const timeout = setTimeout(() => {
      proc.kill();
      console.error(
        `marut timeout: process did not respond within ${MARUT_TIMEOUT_MS}ms`
      );
      resolve(1);
    }, MARUT_TIMEOUT_MS);

    proc.on("close", (code) => {
      clearTimeout(timeout);
      resolve(code ?? 1);
    });

    // Spawn failure: binary not found, not executable, etc.
    proc.on("error", (err) => {
      clearTimeout(timeout);
      console.error(`marut spawn error: ${err.message}`);
      resolve(1);
    });

    proc.stdin.write(JSON.stringify(payload));
    proc.stdin.end();
  });
}

export const MarutPlugin = async ({
  directory,
  worktree,
}: {
  directory: string;
  worktree: string;
}) => {
  // DEBUG: Uncomment to verify plugin loaded
  // console.log("[marut] Plugin loaded successfully");
  // console.log(`[marut] directory: ${directory}`);
  // console.log(`[marut] worktree: ${worktree}`);
  // console.log(`[marut] MARUT_BIN: ${MARUT_BIN}`);
  // console.log(`[marut] MARUT_ARGS: ${MARUT_ARGS.join(" ")}`);

  return {
    "tool.execute.before": async (input: unknown, output: unknown) => {
      // DEBUG: Uncomment to inspect hook payloads
      // console.log("[marut] tool.execute.before triggered");
      // console.log(`[marut] input: ${JSON.stringify(input)}`);
      // console.log(`[marut] output.args: ${JSON.stringify((output as any).args)}`);

      // Validate payload shape before constructing normalized form.
      if (!input || typeof (input as any).tool !== "string") {
        throw new Error(
          "marut: invalid hook payload — input.tool missing or not a string"
        );
      }
      if (!output || typeof (output as any).args !== "object") {
        throw new Error(
          "marut: invalid hook payload — output.args missing or not an object"
        );
      }

      const toolName = (input as { tool: string }).tool;
      const args = (output as { args: Record<string, unknown> }).args;

      // Extract raw_input: command > filePath > JSON.stringify(args).
      // If all are missing/undefined, raw_input becomes "undefined" which
      // will fail validation in the Go normalizer.
      const rawInput =
        (args.command as string | undefined) ??
        (args.filePath as string | undefined) ??
        JSON.stringify(args);

      const normalized = {
        tool: toolName,
        raw_input: rawInput,
        cwd: directory,
        worktree: worktree, // unused by OpenCode but sent for audit log consistency
      };

      // DEBUG: Uncomment to inspect normalized payload and exit codes
      // console.log(`[marut] normalized payload: ${JSON.stringify(normalized)}`);
      const code = await invokeMarut(normalized);
      // console.log(`[marut] marut exit code: ${code}`);

      // Exit code contract: 0 = pass, 1 = internal error, 2 = blocked.
      // Anything else (crash, wrong binary, etc.) is treated as an error.
      if (code === 2) {
        throw new Error("Blocked by Marut: forbidden command");
      } else if (code !== 0) {
        throw new Error(`Marut internal error (exit code ${code})`);
      }
    },
    "tool.execute.after": async (_input: unknown, _output: unknown) => {
      // TODO: monitor mode invocation goes here (Block 11).
      // Contract TBD: fire-and-forget or await? Does it block the tool?
    },
  };
};
