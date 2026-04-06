#!/usr/bin/env node
import { runCli } from "./run.js";

function main(): void {
  const argv = process.argv.slice(2);
  const { code, stdout, stderr } = runCli(argv);
  if (stdout) {
    process.stdout.write(stdout);
  }
  if (stderr) {
    process.stderr.write(stderr);
  }
  process.exit(code);
}

main();
