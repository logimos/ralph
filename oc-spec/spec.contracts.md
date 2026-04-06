# OpenClaw TypeScript Contracts (Reference)

This document provides TypeScript-facing contract shapes for the runtime described in `spec.md`.

Goal:

- keep behavior portable across implementations
- make edge-case expectations explicit
- separate **wire contracts** from implementation details

These are reference contracts distilled from current code paths; consumers should preserve semantics even if internal representations differ.

---

## 1) Gateway Wire Contracts

```ts
type GatewayFrameType = "req" | "res" | "event";

type GatewayRequestFrame<TParams = unknown> = {
  type: "req";
  id: string;
  method: string;
  params?: TParams;
};

type GatewayErrorShape = {
  code: string;
  message: string;
  retryable?: boolean;
  retryAfterMs?: number;
  details?: Record<string, unknown>;
};

type GatewayResponseFrame<TPayload = unknown> = {
  type: "res";
  id: string;
  ok: boolean;
  payload?: TPayload;
  error?: GatewayErrorShape;
};

type GatewayEventFrame<TPayload = unknown> = {
  type: "event";
  event: string;
  payload?: TPayload;
};
```

### 1.1 Connect handshake request

```ts
type GatewayRole = "operator" | "node";

type ConnectParams = {
  minProtocol: number;
  maxProtocol: number;
  role?: GatewayRole;
  scopes?: string[];
  client: {
    id: string;
    displayName?: string;
    version?: string;
    mode?: string;
    platform?: string;
    deviceFamily?: string;
    instanceId?: string;
    modelIdentifier?: string;
  };
  auth?: {
    token?: string;
    password?: string;
  };
  device?: {
    id: string;
    publicKey: string;
    signedAt: number;
    nonce: string;
    signature?: string;
  };
};
```

### 1.2 Connect handshake success

```ts
type HelloOkPayload = {
  type: "hello-ok";
  protocol: number;
  server: {
    version: string;
    connId: string;
  };
  features: {
    methods: string[];
    events: string[];
  };
  snapshot: Record<string, unknown>;
  policy: {
    maxPayload: number;
    maxBufferedBytes: number;
    tickIntervalMs: number;
  };
  auth?: {
    deviceToken: string;
    role: GatewayRole;
    scopes: string[];
    issuedAtMs: number;
    deviceTokens?: Array<{
      deviceToken: string;
      role: GatewayRole;
      scopes: string[];
      issuedAtMs: number;
    }>;
  };
  canvasHostUrl?: string;
};
```

---

## 2) Inbound Message Context Contracts

`finalizeInboundContext` should produce a normalized object that downstream processing can trust.

```ts
type MsgContext = {
  Body?: string;
  RawBody?: string;
  CommandBody?: string;
  BodyForAgent?: string;
  BodyForCommands?: string;

  Provider?: string;
  Surface?: string;
  AccountId?: string;
  ChatType?: string;
  ConversationLabel?: string;

  From?: string;
  To?: string;
  SenderId?: string;
  MessageSid?: string;
  MessageSidFull?: string;
  MessageThreadId?: string | number;

  OriginatingChannel?: string;
  OriginatingTo?: string;
  OriginatingAccountId?: string;
  OriginatingThreadId?: string | number;

  MediaPath?: string;
  MediaUrl?: string;
  MediaPaths?: string[];
  MediaUrls?: string[];
  MediaType?: string;
  MediaTypes?: string[];

  SessionKey?: string;
  CommandSource?: string;
  CommandTargetSessionKey?: string;
  CommandAuthorized?: boolean;
};

type FinalizedMsgContext = MsgContext & {
  Body: string;
  BodyForAgent: string;
  BodyForCommands: string;
  CommandAuthorized: boolean;
};
```

---

## 3) Route Resolution Contracts

```ts
type RoutePeer = {
  kind: "direct" | "group" | "channel";
  id: string;
};

type ResolveAgentRouteInput = {
  channel: string;
  accountId?: string | null;
  peer?: RoutePeer | null;
  parentPeer?: RoutePeer | null;
  guildId?: string | null;
  teamId?: string | null;
  memberRoleIds?: string[];
};

type ResolvedAgentRoute = {
  agentId: string;
  channel: string;
  accountId: string;
  sessionKey: string;
  mainSessionKey: string;
  lastRoutePolicy: "main" | "session";
  matchedBy:
    | "binding.peer"
    | "binding.peer.parent"
    | "binding.peer.wildcard"
    | "binding.guild+roles"
    | "binding.guild"
    | "binding.team"
    | "binding.account"
    | "binding.channel"
    | "default";
};
```

---

## 4) Queue Contracts

## 4.1 Command lane queue

```ts
type CommandLaneTask<T> = () => Promise<T>;

type EnqueueCommandOptions = {
  warnAfterMs?: number;
  onWait?: (waitMs: number, queuedAhead: number) => void;
};
```

Contract semantics:

- FIFO within lane
- lane-local concurrency (`maxConcurrent`)
- reject new tasks during gateway draining
- queue clear rejects pending tasks with lane-cleared error

## 4.2 Followup queue

```ts
type QueueMode = "steer" | "followup" | "collect" | "steer-backlog" | "interrupt" | "queue";
type QueueDropPolicy = "old" | "new" | "summarize";

type QueueSettings = {
  mode: QueueMode;
  debounceMs?: number;
  cap?: number;
  dropPolicy?: QueueDropPolicy;
};

type FollowupRun = {
  prompt: string;
  messageId?: string;
  summaryLine?: string;
  enqueuedAt: number;
  originatingChannel?: string;
  originatingTo?: string;
  originatingAccountId?: string;
  originatingThreadId?: string | number;
  run: {
    sessionId: string;
    sessionKey?: string;
    sessionFile: string;
    workspaceDir: string;
    provider: string;
    model: string;
    timeoutMs: number;
    config: Record<string, unknown>;
  };
};
```

---

## 5) Outbound Delivery and WAL Contracts

```ts
type OutboundChannel = string;

type ReplyPayload = {
  text?: string;
  mediaUrl?: string;
  mediaUrls?: string[];
  audioAsVoice?: boolean;
  replyToId?: string;
  interactive?: Record<string, unknown>;
  channelData?: Record<string, unknown>;
  isReasoning?: boolean;
};

type OutboundDeliveryResult = {
  channel: OutboundChannel;
  messageId: string;
  timestamp?: number;
  meta?: Record<string, unknown>;
};
```

### 5.1 WAL queue entry

```ts
type QueuedDeliveryPayload = {
  channel: OutboundChannel;
  to: string;
  accountId?: string;
  payloads: ReplyPayload[];
  threadId?: string | number | null;
  replyToId?: string | null;
  bestEffort?: boolean;
  silent?: boolean;
};

type QueuedDelivery = QueuedDeliveryPayload & {
  id: string;
  enqueuedAt: number;
  retryCount: number;
  lastAttemptAt?: number;
  lastError?: string;
};
```

WAL semantics:

- enqueue before send attempt
- ack on success via two-phase rename/unlink marker
- fail updates retry metadata
- recovery replays pending `.json` entries

---

## 6) Session Persistence Contracts

```ts
type DeliveryContext = {
  channel?: string;
  to?: string;
  accountId?: string;
  threadId?: string | number;
};

type SessionEntry = {
  sessionId: string;
  sessionFile?: string;
  updatedAt: number;
  systemSent?: boolean;
  abortedLastRun?: boolean;
  lastChannel?: string;
  lastTo?: string;
  lastAccountId?: string;
  lastThreadId?: string | number;
  deliveryContext?: DeliveryContext;

  // model/accounting
  modelProvider?: string;
  model?: string;
  inputTokens?: number;
  outputTokens?: number;
  totalTokens?: number;
  estimatedCostUsd?: number;

  // optional behavior/session controls
  queueMode?: QueueMode;
  queueDebounceMs?: number;
  queueCap?: number;
  queueDrop?: QueueDropPolicy;
  ttsAuto?: "off" | "tool" | "final";

  // optional ACP/runtime metadata
  acp?: Record<string, unknown>;
};
```

### 6.1 Transcript append contract

Use `SessionManager.appendMessage` semantic equivalent.

```ts
type TranscriptRole = "user" | "assistant" | "system";

type TranscriptMessage = {
  role: TranscriptRole;
  content: string | Array<{ type: "text"; text: string }>;
  timestamp: number;
  api?: string;
  provider?: string;
  model?: string;
  usage?: Record<string, number | Record<string, number>>;
  stopReason?: string;
  idempotencyKey?: string;
};
```

Invariant: append API must preserve message parent/ordering model used by compaction/history.

---

## 7) Memory Search Contracts

```ts
type MemorySource = "memory" | "sessions";

type MemorySearchResult = {
  path: string;
  startLine: number;
  endLine: number;
  snippet: string;
  score: number;
  source: MemorySource;
};

type MemoryProviderStatus = {
  backend: "builtin";
  files: number;
  chunks: number;
  provider: string;
  model?: string;
  requestedProvider?: string;
  sources: MemorySource[];
  dirty: boolean;
  vector: {
    enabled: boolean;
    available?: boolean;
    dims?: number;
    loadError?: string;
  };
  fts: {
    enabled: boolean;
    available: boolean;
    error?: string;
  };
};
```

---

## 8) Non-Negotiable Invariants

1. First WS request must be `connect`.
2. Per-session execution must be serialized (session lane).
3. Route precedence must remain deterministic.
4. Session store updates must be lock-protected and re-read inside lock.
5. Transcript appends must use session-manager append semantics.
6. Delivery must be WAL-backed for replay resilience.
7. Memory search must support graceful degradation (hybrid -> FTS-only) without crashing runtime paths.
