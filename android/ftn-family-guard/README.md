# FTN Family Guard — Android

FTN Android client for authorized DNS policy, Family Guard controls, device health and privacy-preserving usage summaries.

## Architecture

```text
Android App
  ├─ Enrollment
  ├─ Policy Store
  ├─ DNS/Profile Manager
  ├─ Local Filter Engine
  ├─ Device Health
  ├─ Usage Aggregator
  └─ WebSocket Client
          │
          ▼
FTN Control Plane
  ├─ Family Policy API
  ├─ DNS Usage API
  ├─ Device API
  └─ Realtime Policy Stream
```

## Modules

- `app`: Android UI, settings and status
- `core/auth`: enrollment and token lifecycle
- `core/policy`: signed/versioned policy cache and rollback
- `core/dns`: resolver/profile state and DNS health
- `core/filter`: domain/IP/CIDR/category rule evaluation
- `core/family`: profile and schedule handling
- `core/usage`: local aggregate counters only
- `core/realtime`: WebSocket reconnect/heartbeat/policy updates
- `core/security`: secure local storage and certificate validation

## Filtering model

The app evaluates authorized FTN policy locally where supported. It may use Android's permitted local VPN/DNS mechanisms for filtering. It does not inspect application payloads, messages, passwords or private content.

Policy precedence:

1. emergency block
2. explicit allow
3. explicit block
4. family profile
5. category policy
6. reputation policy
7. default policy

## Device lifecycle

```text
unregistered
   ↓
enrollment
   ↓
active
   ↓
policy-sync
   ↓
healthy / degraded / offline
   ↓
revoked
```

## Local data

The client keeps the minimum state needed for filtering and operation. Usage is aggregated into counters such as allowed/blocked requests, category totals, resolver health and policy version. Raw DNS query logs are not uploaded to GitHub.

## Control-plane contract

Base API is supplied by the FTN deployment environment.

- `POST /api/v1/family/android/enroll`
- `GET /api/v1/family/devices/{device_id}`
- `GET /api/v1/family/policies/{device_id}`
- `POST /api/v1/family/policies/{device_id}/ack`
- `GET /api/v1/dns/usage/{device_id}/summary`
- `WS /api/v1/family/device/stream`

All mutations require authorization and audit logging.

## Build requirements

Use the repository's selected Android Gradle Plugin/Kotlin versions when implementation is added. Keep runtime endpoints, signing credentials and enrollment secrets outside Git.
