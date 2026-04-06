#!/usr/bin/env node
/**
 * Layers CLI entry — contract in docs/LAYERS_SPEC.md §7
 * Phase 0: health + version only.
 */
const VERSION = "0.0.0";

function main(): void {
  const argv = process.argv.slice(2);
  if (argv[0] === "v1" && argv[1] === "health") {
    console.log(JSON.stringify({ ok: true, version: VERSION, service: "layers" }));
    process.exit(0);
    return;
  }
  console.error(
    "Usage: layers v1 health\nSee docs/LAYERS_SPEC.md — implementation in progress."
  );
  process.exit(1);
}

main();
