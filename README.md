# FTN GitHub Server

Central GitHub-based source and deployment registry for Family Time Network (FTN).

## Purpose

`ftn-github.server` is the control/source layer for managing existing FTN projects across multiple servers, Proxmox VMs/CTs, and geographic zones.

It does **not** copy every provider source into this repository. It records the official source location, deployment target, branch/release, and runtime metadata. Nodes can pull only the services assigned to them.

## Model

```text
GitHub
  │
  ├── FTN source repositories
  ├── approved provider source references
  └── deployment manifests
          │
          ▼
   FTN GitHub Server
          │
     ┌────┼────┐
     ▼    ▼    ▼
   Zone  Zone  Zone
     │    │    │
    VM   CT   Server
```

## Repository layout

```text
config/
  nodes.example.yaml
  sources.example.yaml
  services.example.yaml
scripts/
  validate.sh
  sync-node.sh
  health.sh
systemd/
  ftn-github-server.service.example
.env.example
.gitignore
LICENSE
README.md
```

## Rules

- Never commit passwords, API tokens, SSH private keys, `.env` files, certificates, or provider secrets.
- Keep provider source references pointed at their official repositories.
- Keep runtime secrets on the target node or in the existing secret-management system.
- Use immutable releases/tags for production when available.
- Deploy only to explicitly registered nodes and zones.
- Keep existing service runtimes intact; this repository is the source/deployment management layer.

## First workflow

```bash
git clone https://github.com/beparykamrul-dev/ftn-github.server.git
cd ftn-github.server
./scripts/validate.sh
```

Then copy the example manifests to local/private configuration storage and register the real nodes there.
