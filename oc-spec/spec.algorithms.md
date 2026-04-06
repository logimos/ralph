# OpenClaw Runtime Algorithms (TypeScript Pseudocode)

This document captures runtime behavior as TypeScript-style pseudocode.

Purpose:

- make control flow and invariants explicit
- support ports to any language without losing semantics
- complement `spec.md` and `spec.contracts.md`

---

## 1) WebSocket Handshake Pipeline

```ts
async function onSocketMessage(raw: unknown): Promise<void> {
  const parsed = parseJson(raw);

  if (!clientConnected) {
    assertRequestFrame(parsed);
    assert(parsed.method === "connect");
    assert(validateConnectParams(parsed.params));

    enforcePreauthPayloadLimits(raw);
    enforceProtocolVersion(parsed.params);
    enforceRoleAndScopeRules(parsed.params);
    enforceOriginPolicy(parsed.params, requestMeta);
    const auth = await authenticate(parsed.params, requestMeta);
    enforcePairingAndDevicePolicy(parsed.params, auth);

    const hello = buildHelloOk(parsed.params, auth);
    clientConnected = true;
    sendResponse(parsed.id, { ok: true, payload: hello });
    return;
  }

  assertRequestFrame(parsed);
  await handleGatewayRequest(parsed, clientContext);
}
```

Critical rules:

- Handshake is single-shot state transition.
- Unknown/invalid first frame is a protocol violation.
- Post-connect accepts only request envelopes.

---

## 2) Gateway Method Dispatch

```ts
async function handleGatewayRequest(
  req: GatewayRequestFrame,
  ctx: GatewayRequestContext,
): Promise<void> {
  const authzError = authorizeGatewayMethod(req.method, ctx.client);
  if (authzError) return respond(req.id, { ok: false, error: authzError });

  if (CONTROL_PLANE_WRITE_METHODS.has(req.method)) {
    const budget = consumeControlPlaneWriteBudget(ctx.client);
    if (!budget.allowed) {
      return respond(req.id, {
        ok: false,
        error: unavailableRateLimitError(req.method, budget.retryAfterMs),
      });
    }
  }

  const handler = extraHandlers[req.method] ?? coreHandlers[req.method];
  if (!handler) return respond(req.id, { ok: false, error: unknownMethodError(req.method) });

  await withPluginRuntimeGatewayRequestScope(ctx, async () => {
    await handler(req, ctx);
  });
}
```

---

## 3) Inbound Context Finalization

```ts
function finalizeInboundContext(ctx: MsgContext): FinalizedMsgContext {
  const out = { ...ctx };

  out.Body = sanitize(normalizeNewlines(out.Body ?? ""));
  out.RawBody = normalizeOptionalText(out.RawBody);
  out.CommandBody = normalizeOptionalText(out.CommandBody);
  out.BodyForAgent = sanitize(
    normalizeNewlines(out.BodyForAgent ?? out.CommandBody ?? out.RawBody ?? out.Body),
  );
  out.BodyForCommands = sanitize(
    normalizeNewlines(out.BodyForCommands ?? out.CommandBody ?? out.RawBody ?? out.Body),
  );

  out.ChatType = normalizeChatType(out.ChatType);
  out.ConversationLabel = out.ConversationLabel?.trim() || resolveConversationLabel(out);
  out.CommandAuthorized = out.CommandAuthorized === true;

  // media alignment
  if (hasMedia(out)) {
    out.MediaTypes = normalizeAndPadMediaTypes(out, "application/octet-stream");
    out.MediaType = out.MediaType?.trim() || out.MediaTypes[0];
  }

  return out as FinalizedMsgContext;
}
```

---

## 4) Route Resolution with Tiered Matching

```ts
function resolveAgentRoute(input: ResolveAgentRouteInput): ResolvedAgentRoute {
  const scope = normalizeScope(input);
  const bindings = getBindingsForChannelAccount(scope.channel, scope.accountId);

  const tiers: Array<{
    matchedBy: ResolvedAgentRoute["matchedBy"];
    candidates: Binding[];
    predicate: (b: Binding) => boolean;
  }> = [
    { matchedBy: "binding.peer", candidates: byPeer(scope.peer), predicate: isExactPeer },
    {
      matchedBy: "binding.peer.parent",
      candidates: byPeer(scope.parentPeer),
      predicate: isExactPeer,
    },
    {
      matchedBy: "binding.peer.wildcard",
      candidates: byPeerWildcard(scope.peer),
      predicate: isPeerWildcard,
    },
    {
      matchedBy: "binding.guild+roles",
      candidates: byGuildWithRoles(scope.guildId),
      predicate: hasGuildAndRole,
    },
    { matchedBy: "binding.guild", candidates: byGuild(scope.guildId), predicate: hasGuildOnly },
    { matchedBy: "binding.team", candidates: byTeam(scope.teamId), predicate: hasTeam },
    { matchedBy: "binding.account", candidates: accountBindings(), predicate: isAccountScoped },
    { matchedBy: "binding.channel", candidates: channelBindings(), predicate: isChannelWildcard },
  ];

  for (const tier of tiers) {
    const match = tier.candidates.find((b) => tier.predicate(b) && matchesScope(b, scope));
    if (match) return buildResolvedRoute(match.agentId, tier.matchedBy, scope);
  }

  return buildResolvedRoute(defaultAgentId(), "default", scope);
}
```

Determinism rules:

- preserve stable binding order
- first tier hit wins
- stable session key derivation and normalization

---

## 5) Command Lane Queue Pump

```ts
function enqueueCommandInLane<T>(lane: string, task: () => Promise<T>): Promise<T> {
  if (gatewayDraining) return Promise.reject(new GatewayDrainingError());

  const state = getOrCreateLane(lane);
  return new Promise<T>((resolve, reject) => {
    state.queue.push({ task, resolve, reject, enqueuedAt: Date.now() });
    drainLane(lane);
  });
}

function drainLane(lane: string): void {
  const state = getOrCreateLane(lane);
  if (state.draining) return;
  state.draining = true;

  try {
    while (state.active.size < state.maxConcurrent && state.queue.length > 0) {
      const entry = state.queue.shift()!;
      const taskId = nextTaskId++;
      const generation = state.generation;
      state.active.add(taskId);

      void (async () => {
        try {
          const result = await entry.task();
          if (completeTask(state, taskId, generation)) drainLane(lane);
          entry.resolve(result);
        } catch (err) {
          if (completeTask(state, taskId, generation)) drainLane(lane);
          entry.reject(err);
        }
      })();
    }
  } finally {
    state.draining = false;
  }
}
```

---

## 6) Followup Queue Drain (collect/followup/steer-backlog)

```ts
async function scheduleFollowupDrain(
  key: string,
  runFollowup: (run: FollowupRun) => Promise<void>,
): Promise<void> {
  const queue = beginQueueDrain(key);
  if (!queue) return;

  try {
    while (queue.items.length > 0 || queue.droppedCount > 0) {
      await waitForDebounce(queue.debounceMs, queue.lastEnqueuedAt);

      if (queue.mode === "collect") {
        const items = queue.items.splice(0);
        if (items.length === 0) break;
        const prompt = buildCollectPrompt(items, queue.summaryLines, queue.droppedCount);
        const run = items.at(-1)!.run;
        await runFollowup({ ...items.at(-1)!, prompt, run, enqueuedAt: Date.now() });
        clearQueueSummaryState(queue);
        continue;
      }

      const item = queue.items.shift();
      if (!item) break;
      await runFollowup(item);
    }
  } finally {
    queue.draining = false;
    if (queue.items.length === 0 && queue.droppedCount === 0) deleteQueue(key);
    else void scheduleFollowupDrain(key, runFollowup);
  }
}
```

---

## 7) Agent Turn Orchestration

```ts
async function runReplyAgent(input: RunReplyInput): Promise<ReplyPayload | ReplyPayload[] | undefined> {
  const queueAction = resolveActiveRunQueueAction({
    isActive: input.isActive,
    isHeartbeat: input.isHeartbeat,
    shouldFollowup: input.shouldFollowup,
    queueMode: input.queue.mode,
  });

  if (queueAction === "drop") return undefined;
  if (queueAction === "enqueue-followup") {
    enqueueFollowupRun(input.queueKey, input.followupRun, input.queue);
    return undefined;
  }

  await signalTypingStart();
  await maybeRunPreflightCompaction();
  await maybeRunMemoryFlush();

  const outcome = await runAgentTurnWithFallback(input);
  const payloads = normalizeReplyPayloads(outcome.payloads);
  const payloadsWithUsage = appendUsageLineIfEnabled(payloads, outcome.usage);

  return finalizeWithFollowup(payloadsWithUsage, input.queueKey);
}
```

---

## 8) Outbound WAL Delivery

```ts
async function deliverOutboundPayloads(params: DeliverParams): Promise<OutboundDeliveryResult[]> {
  const queueId = params.skipQueue ? null : await enqueueDelivery(params);
  let partialFailure = false;

  try {
    const results = await deliverOutboundPayloadsCore({
      ...params,
      onError: (err, payload) => {
        partialFailure = true;
        params.onError?.(err, payload);
      },
    });

    if (queueId) {
      if (partialFailure) await failDelivery(queueId, "partial delivery failure");
      else await ackDelivery(queueId);
    }
    return results;
  } catch (err) {
    if (queueId) {
      if (isAbortError(err)) await ackDelivery(queueId);
      else await failDelivery(queueId, String(err));
    }
    throw err;
  }
}
```

Recovery loop expectations:

- scan pending WAL files
- skip backoff-ineligible entries
- retry eligible entries
- move permanently failed entries to `failed/`

---

## 9) Session Store Locked Update

```ts
async function updateSessionStore<T>(
  storePath: string,
  mutator: (store: Record<string, SessionEntry>) => Promise<T> | T,
): Promise<T> {
  return withSessionStoreLock(storePath, async () => {
    const store = loadSessionStore(storePath, { skipCache: true });
    const previousAcp = snapshotAcpMetadata(store);
    const result = await mutator(store);
    preserveExistingAcpMetadata(previousAcp, store);
    await saveSessionStoreUnlocked(storePath, store);
    return result;
  });
}
```

Hard guarantees:

- load-inside-lock before mutate
- atomic writes
- maintenance and archival pass before commit

---

## 10) Hybrid Memory Search

```ts
async function memorySearch(query: string, opts: SearchOpts): Promise<MemorySearchResult[]> {
  const cleaned = query.trim();
  if (!cleaned) return [];

  if (syncOnSearch && indexDirty) void syncInBackground();

  const maxResults = opts.maxResults ?? defaults.maxResults;
  const minScore = opts.minScore ?? defaults.minScore;

  if (!embeddingProviderAvailable) {
    const keywordOnly = await keywordSearchWithExtraction(cleaned);
    return selectByScore(keywordOnly, maxResults, minScore, /* relaxed */ textWeight);
  }

  const vector = await vectorSearch(cleaned);
  const keyword = ftsAvailable ? await keywordSearch(cleaned) : [];
  if (!hybridEnabled || !ftsAvailable) return vector.filter((r) => r.score >= minScore).slice(0, maxResults);

  const merged = mergeHybridResults({
    vector,
    keyword,
    vectorWeight,
    textWeight,
    temporalDecay,
    mmr,
  });
  return selectByScore(merged, maxResults, minScore);
}
```

---

## 11) Algorithmic Invariants Checklist

1. Queue pumps must be re-entrant safe (`draining` guard).
2. Session lane + main lane constraints must prevent same-session concurrent runs.
3. Followup queue dedupe should key on routing metadata + message identity.
4. Store mutation must not overwrite newer state from parallel writers.
5. Transcript append must preserve ordering/chain invariants expected by compaction.
6. Delivery WAL must never replay already-acked messages.
7. Hybrid memory retrieval must remain available when vector backend is down (FTS-only fallback).
