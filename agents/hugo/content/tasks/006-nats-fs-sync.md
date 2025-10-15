---
title: "Task 006 — NATS FS Sync"
summary: "Design hot/cold filesystem sync across local disk, NATS, and S3."
draft: false
---

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
* Clients and services can **subscribe to change events** for local and NATS updates.

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
* **Local FS (ephemeral per job/agent)** – files materialized to `/tmp/job-<id>` or an app-chosen workspace; removed when idle.
* **NATS Object Store (Hot)** – active artifacts keyed by content hash: `objects/<sha>`.
* **S3 (Cold)** – durable, lifecycle-managed store keyed by the same `<sha>`.

### Metadata
* **KV Store (`fs.index`)**
  - `projects/<name>/current → { manifest_sha }`
  - `objects/<sha> → { size, mime, created, cold: { bucket, key, class }, refs: {...} }`
* **Manifests (content-addressed JSON)** list paths → sha/size.

---

## Sync Model

### Local FS ⇄ NATS (Hot)
- Real-time (event-driven).
- FS watchers → compute SHA → Put to NATS → publish events.
- Conflict resolution: Last-Writer-Wins for now.

### NATS (Hot) ⇄ S3 (Cold)
- Eventual consistency via archiver.
- Strict vs fast promotion modes.

### NATS (Hot) ⇄ Other Local FS
- Servers subscribe to events; fetch SHAs from NATS or rehydrate from S3.

### Eviction / Rehydration
- Evict only if cold copy exists.
- Rehydrate from S3 → NATS → local.

### Ordering & Guarantees
- Objects immutable by SHA.
- Manifests updated via CAS.
- Emit `ObjectCreated` before `PathUpdated`.

---

## Change Events

### Event Categories
- Object events: `ObjectCreated`, `ObjectArchived`, `ObjectEvicted`.
- Path events: `PathCreated`, `PathUpdated`, `PathDeleted`, `PathMoved`.
- Project events: `ProjectCurrentChanged`.

### Event Envelope
- id, ts, type, actor, payload, trace metadata.

### Transport
- NATS subjects `fs.events.*`.
- SSE/webhooks/pull API for downstream consumers.

### Local FS Watch
- Use OS-specific watchers (inotify/FSEvents/etc.).
- Debounce, filter, backoff.

---

## Consistency Model
- Objects immutable, manifests CAS.
- Last-Writer-Wins baseline; CRDT later.

---

## Lifecycle & Policies
- Eviction: LRU/time-based with “pin hot”.
- Promotion: manual or heuristic.
- Rehydration: NATS-first or direct S3 stream.

---

## Security & RBAC
- Tenant prefixes in NATS/S3.
- Scoped creds, KMS keys, pre-signed URLs.
- Immutable audit stream.

---

## Datastar GUI

### Views
- Projects summary, explorer tree, events feed, job history, metrics.

### Actions
- Promote/Evict, resync/repair, share links.

### UX
- Color-coded events, cursor-based navigation.

---

## Interfaces
- Upload intent, object put, path commit, list changes, promote/evict APIs.

---

## Performance & Sizing
- Chunked uploads, streaming SHA-256, optional compression, rate limiting.

---

## Observability
- Metrics for hot/cold size, object counts, event rate, rehydration latency.
- Structured logs with correlation IDs.
- Health checks per service; consumer lag dashboards.

---

## Open Questions & Tradeoffs
- System of record (S3 vs NATS), manifest granularity, eviction policy, ordering guarantees, rehydration strategy, conflict resolution, compression, delivery semantics, tenant isolation.

---

## Next Steps
1. Finalize event schemas and manifest spec.
2. Implement sync client (watcher + uploader).
3. Build archiver/evictor/rehydrator/promoter services.
4. Ship Datastar GUI MVP.
5. Decide eviction/promotion policies.
