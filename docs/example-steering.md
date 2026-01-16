# Example: Steering stack choices mid-run

This example shows how to steer Ralph to:
- Start a "fake" app with Kotlin + Spring 3.3 + OpenSearch + SQS
- Add RabbitMQ after work has begun
- Decide against Kafka after briefly considering it
- Add tests for the resulting endpoints

The workflow uses **notes**, **plan generation**, and **nudges** for mid-run guidance.

## 1) Initial requirements (notes)

Save the following as `notes.md`:

```md
Project: WarehousePulse

Requirements:
- Kotlin + Spring Boot 3.3
- OpenSearch for search and analytics
- SQS for async job processing
- REST API: create shipment, get shipment, search shipments
- Add integration tests for all endpoints
```

## 2) Generate the initial plan

```bash
ralph -generate-plan -notes notes.md -output plan.json -verbose
```

## 3) Start iterations

```bash
ralph -iterations 10 -verbose
```

## 4) Steer the stack after the run starts

Ralph can be guided mid-run via `nudges.json`.
Create `nudges.json` (or edit it while the run is active):

```json
[
  {
    "id": "stack-kotlin-spring",
    "type": "constraint",
    "message": "Use Kotlin and Spring Boot 3.3 for all backend services."
  },
  {
    "id": "stack-search",
    "type": "constraint",
    "message": "Use OpenSearch for search indexes; avoid Elasticsearch."
  },
  {
    "id": "stack-queue-sqs",
    "type": "constraint",
    "message": "Use AWS SQS for async jobs; build queue client abstractions."
  }
]
```

If the run is already in progress, Ralph will pick these up on the next iteration.

### Add RabbitMQ after work starts

Append a new nudge:

```json
[
  {
    "id": "stack-rabbitmq",
    "type": "enhancement",
    "message": "Add RabbitMQ support in addition to SQS. Make it configurable."
  }
]
```

Effect: Ralph will add an abstraction layer to allow SQS or RabbitMQ,
and update configuration and docs accordingly.

### Consider Kafka, then remove it

If you briefly want Kafka, add this:

```json
[
  {
    "id": "stack-kafka",
    "type": "experiment",
    "message": "Explore adding Kafka support behind the same queue interface."
  }
]
```

If you later decide against Kafka, add a follow-up nudge:

```json
[
  {
    "id": "stack-remove-kafka",
    "type": "constraint",
    "message": "Do not implement Kafka. Remove any Kafka references from the plan."
  }
]
```

Ralph will reconcile the active guidance on the next iteration and drop
Kafka-related steps or plan items.

## 5) Request endpoint tests

Add a test-focused nudge:

```json
[
  {
    "id": "endpoint-tests",
    "type": "quality",
    "message": "Add integration tests for all shipment endpoints. Use the real HTTP server."
  }
]
```

## 6) Run with nudges enabled

```bash
ralph -iterations 10 -verbose -nudge-file nudges.json
```

## 7) Optional: make it persistent with memory

If you want these constraints to apply across future runs, add a memory file:

```json
{
  "project_stack": "Kotlin, Spring Boot 3.3, OpenSearch, SQS, RabbitMQ",
  "queue_strategy": "Configurable queue interface; no Kafka",
  "test_policy": "Integration tests for all public endpoints"
}
```

```bash
ralph -iterations 10 -memory-file project-memory.json
```
