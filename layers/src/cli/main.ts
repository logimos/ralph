#!/usr/bin/env node
import { readStdinSync } from "./stdin.js";
import { runCli } from "./run.js";

function main(): void {
  const argv = process.argv.slice(2);
  const stdin = readStdinSync();
  const { code, stdout, stderr } = runCli(argv, process.env, stdin);
  if (stdout) {
    process.stdout.write(stdout);
  }
  if (stderr) {
    process.stderr.write(stderr);
  }
  process.exit(code);
}

main();
