

# 006 nats-fs-sync

## Problem

Server nodes need a file system (FS) to work off.
We have many servers, so the FS between servers needs to be kept in sync.

We also need to have cold data in S3 that can be synced into the FS to become hot data.

We need the ability to remove or deactivate a user’s FS when it is not in use, to save hot storage resources.

---

## Solution

**Local FS ⇄ NATS (Hot) ⇄ S3 (Cold)**

---

## Purpose

Provide a general, application-agnostic file platform where:

* Applications always read/write **real files on local disk**.
* The platform synchronizes artifacts to a **hot tier (NATS Object Store)** and **cold tier (S3)**.
* Clients and services can **subscribe to change events** (like Google Drive “changes”) for local and NATS updates.

No FUSE mounts, no symlinks—explicit staging and explicit APIs.

---

## Goals

* **Universal file types** (binary/text/archives/media).
* **Deterministic addressing** (SHA-256) + versioning.
* **Low-latency hot tier**, **durable cold tier**.
* **Change Events** on local and NATS mutations.
* **Simple, observable flows** for humans and agents.

---

## Architecture Overview

### Storage Tiers

* **Local FS (ephemeral per job/agent)**
  Files materialized to `/tmp/job-<id>` or an app-chosen workspace.
  Removed when not in use.

* **NATS Object Store (Hot)**
  Active artifacts keyed by content hash: `objects/<sha>`.
  Globally replicated for fast sync across servers.

* **S3 (Cold)**
  Durable, lifecycle-managed store keyed by the **same** `<sha>` (or mirrored path).
  Cheap long-term retention; slower to access.

### Metadata

* **KV Store (`fs.index`)**

  * `projects/<name>/current → { manifest_sha }`
  * `objects/<sha> → { size, mime, created, cold: { bucket, key, class }, refs: {...} }`

* **Manifests (content-addressed JSON)**
  Example:
  { "version":1, "entries":\[ {"path":"docs/spec.md","sha":"<sha1>","size":1234}, {"path":"images/logo.png","sha":"<sha2>","size":20480} ] }

---

## Sync Model

### Local FS ⇄ NATS (Hot)

* Real-time (event-driven).
* FS watchers → compute SHA → Put to NATS → publish `ObjectCreated`/`PathUpdated`.
* Latency: sub-second to a few seconds.
* Conflict resolution: Last-Writer-Wins; optional CRDT later.

### NATS (Hot) ⇄ S3 (Cold)

* Near-real-time mirroring (eventual consistency).
* Archiver uploads new SHAs to S3; updates KV with `cold`.
* Latency: seconds to tens of seconds.
* Durability gates: Strict mode waits for S3 confirm before path flip; Fast mode flips on NATS write, evict only after S3 confirm.

### NATS (Hot) ⇄ Other Local FS

* Real-time fan-out via events.
* Servers subscribe to events; fetch SHAs from NATS or rehydrate from S3.

### Eviction / Rehydration

* Evict only if cold copy exists.
* Rehydrate from S3 into NATS (or stream directly S3 → local FS).
* Emit events: `ObjectEvicted`, `ObjectRehydrated`.

### Ordering & Guarantees

* Objects immutable by SHA.
* Manifests updated via CAS.
* Publish `ObjectCreated` before `PathUpdated`.
* Consumers reconcile with manifest `rev` or `changes?since=<cursor>`.

---

## Change Events

### Event Categories

* **Object events:** `ObjectCreated`, `ObjectArchived`, `ObjectEvicted`.
* **Path events:** `PathCreated`, `PathUpdated`, `PathDeleted`, `PathMoved`.
* **Project events:** `ProjectCurrentChanged`.

### Event Envelope

* id: unique event id
* ts: timestamp
* type: event type
* actor: user/system/agent id
* payload: event-specific fields
* trace: job id, region, request id

### Transport

* NATS subjects: `fs.events.object.*`, `fs.events.path.*`, `fs.events.project.*`
* Fan-out: SSE for GUIs, webhooks for tenants, pull API with cursors.

### Local FS Watch

* Use inotify (Linux), FSEvents (macOS), ReadDirectoryChangesW (Windows).
* Debounce, filter, backoff on errors.

---

## Consistency Model

* Objects: immutable, content-addressed.
* Manifests: atomic CAS updates.
* Conflict resolution: Last-Writer-Wins (default); CRDT possible later.

---

## Lifecycle & Policies

* Eviction: LRU/time-based with “pin hot” override.
* Promotion: manual or heuristic.
* Rehydration: NATS hot-first (default) or direct S3 stream.

---

## Security & RBAC

* Tenant prefixes in NATS and S3.
* Scoped NATS creds; per-tenant S3 KMS keys.
* Pre-signed S3 URLs for external downloads.
* Audit stream of immutable events.

---

## Datastar GUI

### Views

* Projects: list + current manifest rev + hot/cold status.
* Explorer: tree of paths → SHA, size, Hot/Cold/Local.
* Events: live SSE feed with filters.
* Jobs: status/history + output links.
* Metrics: hot usage, cold usage, hit/miss ratio, latency.

### Actions

* Promote/Evict hot state.
* Resync/Repair.
* Generate share links.

### UX

* Color-coded events: green=update, red=delete, yellow=eviction, blue=promotion.
* Cursor-based scroll; bookmarkable cursors.

---

## Interfaces

* Submit Upload Intent → returns whether upload is needed.
* Put Object (if needed) → stream to NATS or presigned URL, emits `ObjectCreated`.
* Commit Path Update → flips manifest entry, emits Path events.
* List Changes → returns events since a cursor.
* Promote/Evict → mark artifacts or projects as hot/cold.

---

## Performance & Sizing

* Chunked, resumable uploads for large files.
* Streaming SHA-256 during upload.
* Store raw by default; compression optional.
* Coalesce bursts; rate-limit events.

---

## Observability

* Metrics: hot size, cold size, objects count, manifest count, event rate, rehydration latency.
* Structured logs with correlation IDs.
* Health checks per service; consumer lag dashboards.

---

## Open Questions & Tradeoffs

* System of record: S3 as truth vs NATS as hot authority.
* Manifest granularity: single manifest vs nested.
* Eviction policy: LRU vs time-based vs manual pinning.
* Write ordering: require S3 confirm vs optimistic.
* Rehydration: re-hot into NATS vs direct S3 stream.
* Conflict resolution: Last-Writer-Wins vs CRDT.
* Compression: raw vs compressed.
* Event delivery: exactly-once vs at-least-once with idempotency.
* Tenant isolation: per-tenant buckets vs shared with ACLs.

---

## Next Steps

1. Finalize event schemas and subject naming.
2. Define manifest v1 and job spec v1 JSON Schemas.
3. Implement Sync Client (FS watcher + uploader + consumer).
4. Build Archiver, Evictor, Rehydrator, Promoter microservices.
5. Build Datastar GUI MVP (project list, artifact states, live events).
6. Decide eviction/promotion defaults and lifecycle policies.


