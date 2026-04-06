# Layers

TypeScript service for **recording and retrieving** project memory for the [Ralph](https://github.com/logimos/ralph) loop. Ralph integrates via a **versioned CLI/HTTP contract** — see the canonical spec:

**[docs/LAYERS_SPEC.md](../docs/LAYERS_SPEC.md)**

## Requirements

- **Node.js 20+** (see `engines` in `package.json`)

## Development

From the **repository root** (npm workspaces):

```bash
npm ci
npm run build -w layers
npm run test -w layers
npm run lint -w layers
npm run format:check -w layers
```

Or use **Make**:

```bash
make layers-test
```

## CLI

After `npm run build -w layers`:

```bash
node layers/dist/cli/main.js v1 health
```

**Phase 1 — record** (stdin JSON, see `docs/LAYERS_SPEC.md` §7.3):

```bash
echo '{"projectRoot":"/abs/repo","entries":[{"type":"decision","content":"Use SQLite","source":"agent"}]}' \
  | node layers/dist/cli/main.js v1 record
```

Database: `<projectRoot>/.layers/memory.db` unless overridden by `dataDir` in the JSON request or the **`LAYERS_DATA_DIR`** environment variable (see `docs/LAYERS_SPEC.md` §7.1).

**Phase 1 — import** legacy Ralph memory (`.ralph-memory.json` from `internal/memory`):

```bash
echo '{"projectRoot":"/abs/repo"}' | node layers/dist/cli/main.js v1 import-ralph-memory
# or explicit file:
echo '{"projectRoot":"/abs/repo"}' | node layers/dist/cli/main.js v1 import-ralph-memory /path/to/.ralph-memory.json
```

If stdin is omitted, set `LAYERS_PROJECT_ROOT` and optional path as first argument.

## Status

- **Phase 0**: scaffold, health, lint, tests, CI.
- **Phase 1**: SQLite + FTS5, `v1 record`, `v1 import-ralph-memory` (idempotent). **Phase 2**: `retrieve` (hybrid search) per `docs/LAYERS_SPEC.md`.

## Quick links

- OpenClaw reference (in-repo): `oc-spec/spec.md` §8, `spec.algorithms.md` §10
- Ralph memory today: `internal/memory/memory.go`
- Loop analysis: `docs/RALPH_LOOP_ENHANCEMENTS.md` §6.5–7.12
