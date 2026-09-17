# FTN Git Center API

The Family Guard service exposes a minimal registry projection for the FTN Control Panel.

## Endpoint

`GET /api/v1/git/registry`

Returns registered repositories, node/service bindings, build/deploy policy, and a generation timestamp.

## Safety contract

- Registered repositories only.
- Registered nodes and services only.
- Registered build profiles only.
- Deploy and rollback require approval.
- No arbitrary shell-command API.
- Git credentials/tokens stay in runtime secrets or environment variables.
- Raw network traffic, payloads, secrets, and private runtime data are not exported to GitHub.

This endpoint is intentionally read-only. The existing deployment scripts remain the execution layer; future write actions must validate the registry and create an auditable approval record before execution.
