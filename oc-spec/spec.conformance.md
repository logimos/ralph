# OpenClaw Conformance Test Manifest

This manifest defines compatibility tests for any OpenClaw-style runtime implementation.

Use this with:

- `spec.md` (architecture + lifecycle)
- `spec.contracts.md` (wire/storage contracts)
- `spec.algorithms.md` (reference control flow)

---

## 0) Test Harness Requirements

### 0.1 Determinism

- Freeze clock for tests that compare timestamps.
- Seed UUID/random sources where IDs appear in expected outputs.
- Disable nondeterministic background jitter in test mode.

### 0.2 Fixture Format

Use fixture bundles per test case:

```ts
type ConformanceFixture = {
  id: string;
  setup?: {
    config?: Record<string, unknown>;
    sessionStore?: Record<string, unknown>;
    transcriptLines?: string[];
    queueState?: Record<string, unknown>;
    walEntries?: Record<string, unknown>[];
    memoryIndexRows?: Record<string, unknown>[];
  };
  input: unknown;
  expected: {
    output?: unknown;
    events?: unknown[];
    persisted?: Record<string, unknown>;
    invariants?: string[];
  };
};
```

### 0.3 Verdict Policy

- **Pass**: output + side effects + invariants all match.
- **Fail**: any mismatch in envelope shape, ordering, persistence semantics, or state machine outcome.

---

## 1) Gateway / Ingestion Conformance

## GATE-001: First frame must be connect

- Input: WS frame `{type:"req", method:"agent"}`
- Expected:
  - handshake rejected
  - error response (if id present)
  - socket closed with policy/protocol reason

## GATE-002: Invalid connect params rejected

- Input: malformed `connect.params`
- Expected:
  - validation error response
  - no client connected state transition

## GATE-003: Protocol mismatch rejected

- Input: `minProtocol/maxProtocol` excluding server protocol
- Expected:
  - response error `protocol mismatch`
  - close code for protocol mismatch path

## GATE-004: Unauthorized method rejected before handler

- Setup: connected client with insufficient scopes
- Input: protected method call
- Expected:
  - `ok:false` auth/scope error
  - handler not invoked

## GATE-005: Control-plane write rate limit

- Setup: same actor sends write methods repeatedly
- Input: 4th call within limit window
- Expected:
  - `retryable=true`
  - `retryAfterMs` present

---

## 2) Routing Conformance

## ROUTE-001: Tier precedence peer > guild > default

- Setup: bindings matching same message at multiple tiers
- Input: message with peer + guild metadata
- Expected:
  - highest tier selected (`binding.peer`)

## ROUTE-002: Parent peer inheritance

- Setup: child thread has no direct binding, parent has binding
- Input: `peer=thread`, `parentPeer=boundChannel`
- Expected:
  - matchedBy `binding.peer.parent`

## ROUTE-003: Guild+roles must require role intersection

- Setup: binding has `guildId + roles`
- Input: guild matches but no member role overlap
- Expected:
  - no guild+roles match
  - fall through to lower tiers

## ROUTE-004: Deterministic session key

- Input: same logical route with casing/spacing differences
- Expected:
  - same normalized `sessionKey`
  - stable `mainSessionKey`

---

## 3) Command Lane Queue Conformance

## LANE-001: FIFO order in lane

- Setup: enqueue tasks A,B,C into same lane
- Expected:
  - completion order A,B,C

## LANE-002: maxConcurrent enforced

- Setup: lane concurrency=2, enqueue long tasks
- Expected:
  - at most 2 active simultaneously

## LANE-003: clear lane rejects queued only

- Setup: one active + two queued tasks
- Action: clear lane
- Expected:
  - queued tasks reject with lane-cleared error
  - active task completes normally

## LANE-004: gateway draining rejects new enqueues

- Setup: mark draining
- Input: new enqueue
- Expected:
  - immediate reject with draining error

---

## 4) Followup Queue Conformance

## FQ-001: message-id dedupe

- Setup: same messageId + same routing metadata
- Input: enqueue duplicate followup
- Expected:
  - duplicate skipped

## FQ-002: collect mode aggregates prompts

- Setup: queue mode `collect`, N enqueues within debounce window
- Expected:
  - one synthetic collected prompt run
  - includes all queued entries in order

## FQ-003: cross-channel collect safety

- Setup: collect mode with different `(channel,to,account,thread)` tuples
- Expected:
  - no invalid cross-target aggregation
  - drained as individual runs when cross-target detected

## FQ-004: drop policy summarize

- Setup: cap reached, dropPolicy summarize
- Input: additional enqueue
- Expected:
  - oldest dropped
  - dropped count + summary lines updated
  - summary prompt emitted in subsequent drain

---

## 5) Session Store + Transcript Conformance

## STORE-001: lock-reload-mutate-write

- Setup: concurrent writers
- Action: two updates race
- Expected:
  - no lost update due to stale pre-lock snapshot
  - final store reflects both mutations correctly

## STORE-002: normalized session key migration behavior

- Setup: legacy mixed-case key entries
- Action: update by normalized key
- Expected:
  - canonical key persisted
  - legacy aliases removed as defined

## TX-001: transcript append ordering invariant

- Action: append user, then assistant message
- Expected:
  - transcript includes messages in append order
  - append API emits update event

## TX-002: idempotent mirror append

- Setup: same `idempotencyKey` repeated
- Expected:
  - only one mirrored assistant append

---

## 6) Outbound + WAL Conformance

## WAL-001: enqueue before send

- Setup: instrumentation on send path
- Expected:
  - WAL entry exists before adapter send attempt begins

## WAL-002: ack removes replay eligibility

- Setup: successful delivery then simulated restart
- Expected:
  - no replay of already-acked entry

## WAL-003: fail increments retry metadata

- Setup: adapter throws retriable error
- Expected:
  - `retryCount` increments
  - `lastAttemptAt` + `lastError` set

## WAL-004: partial best-effort failure handling

- Setup: multi-payload bestEffort send where one payload fails
- Expected:
  - delivery returns partial success
  - WAL entry marked failed (not acked)

---

## 7) Processing / Runner Conformance

## RUN-001: active-run queue decision

- Setup: active run true, mode followup
- Input: new inbound message
- Expected:
  - enqueued followup (not run immediately)

## RUN-002: heartbeat while active drops

- Setup: active run + heartbeat input
- Expected:
  - no run execution
  - no queued followup

## RUN-003: fallback transition persisted

- Setup: selected model fails, fallback model succeeds
- Expected:
  - fallback notice state persisted in session entry
  - lifecycle fallback event emitted

## RUN-004: usage accounting persisted

- Setup: run returns nonzero usage
- Expected:
  - session usage/tokens fields updated
  - response usage line behavior matches configuration mode

---

## 8) Memory Layer Conformance

## MEM-001: history layer durability

- Action: create/update session + transcript entries
- Expected:
  - `sessions.json` metadata and transcript JSONL remain consistent

## MEM-002: hybrid weighted merge

- Setup:
  - vector result `vScore=0.9`
  - keyword result `tScore=0.5`
  - weights `vectorWeight=0.7`, `textWeight=0.3`
- Expected:
  - merged score `0.7*0.9 + 0.3*0.5 = 0.78`

## MEM-003: temporal decay affects ordering

- Setup: same base score, different file ages
- Expected:
  - more recent item ranks higher when temporal decay enabled

## MEM-004: MMR diversity rerank

- Setup: near-duplicate top snippets
- Expected:
  - reranked set includes more diverse snippets when MMR enabled

## MEM-005: FTS-only degradation

- Setup: embedding provider unavailable
- Input: query over indexed text
- Expected:
  - keyword results still returned
  - no hard failure

## MEM-006: readonly DB recovery

- Setup: simulate readonly write failure during sync
- Expected:
  - manager reopens DB
  - retries sync once
  - status exposes recovery counters

---

## 9) End-to-End Golden Scenarios

## E2E-001: WS connect -> agent run -> outbound -> mirror

- Validate full sequence:
  - connect handshake success
  - request dispatch success
  - queued execution
  - outbound send
  - transcript mirror
  - session metadata touch

## E2E-002: Busy-session burst with collect mode

- Validate:
  - first run executes
  - burst inputs enqueue
  - collect prompt generated and drained once
  - no dropped routing metadata

## E2E-003: Restart recovery with pending WAL

- Validate:
  - pending deliveries replayed once
  - acked deliveries not replayed
  - permanent failures moved to failed directory

---

## 10) CI Gate Recommendation

Minimum gate:

1. GATE-001..005
2. ROUTE-001..004
3. LANE-001..004
4. FQ-001..004
5. STORE-001, TX-001..002
6. WAL-001..004
7. RUN-001..004
8. MEM-001..006
9. E2E-001..003

Promotion rule:

- No compatibility release if any case above fails.
- Changes that intentionally alter semantics must update:
  - this manifest
  - expected fixtures
  - versioned compatibility notes.
