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

## CLI (Phase 0)

After `npm run build -w layers`:

```bash
node layers/dist/cli/main.js v1 health
```

Response: JSON with `ok`, `version` (from `layers/package.json`), and `service: "layers"`.

## Status

**Phase 0** complete: TypeScript scaffold, strict build, ESLint, Prettier, Vitest, CI workflow, versioned health command. Later phases implement storage and retrieval per `docs/LAYERS_SPEC.md`.

## Quick links

- OpenClaw reference (in-repo): `oc-spec/spec.md` §8, `spec.algorithms.md` §10
- Ralph memory today: `internal/memory/memory.go`
- Loop analysis: `docs/RALPH_LOOP_ENHANCEMENTS.md` §6.5–7.12
