#!/usr/bin/env node
import { readStdinSync } from "./stdin.js";
import { runCli } from "./run.js";

async function main(): Promise<void> {
  const argv = process.argv.slice(2);
  const stdin = readStdinSync();
  const { code, stdout, stderr } = await runCli(argv, process.env, stdin);
  if (stdout) {
    process.stdout.write(stdout);
  }
  if (stderr) {
    process.stderr.write(stderr);
  }
  process.exit(code);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
