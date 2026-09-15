# FTN DDNS

FTN DDNS is a provider-neutral dynamic DNS layer. It updates only explicitly registered records when an approved node's public IPv4/IPv6 address changes.

## Flow

1. Node agent detects the current public address.
2. FTN validates the address and target record against an allowlist.
3. The provider adapter reads credentials from the runtime environment.
4. The adapter reads the current DNS record before changing it.
5. Only the changed A/AAAA record is updated.
6. The result is written to the FTN audit stream.
7. GeoFlow/monitoring can consume the resulting endpoint state.

## Supported targets

- Cloudflare DNS
- FTN/custom authoritative DNS
- Future provider adapters through the same interface

## Safety defaults

- Dry-run enabled for new targets.
- Allowlist-only record updates.
- No credentials in Git.
- No zone-wide deletion.
- No wildcard replacement by default.
- Preserve existing TTL unless the target explicitly requests another value.
- Audit every update with timestamp, node, record and old/new address.

## Dual-stack

A and AAAA are handled independently so an IPv4 change does not overwrite an IPv6 record, and vice versa.

## Cloudflare API

Cloudflare's DNS API supports listing, creating, updating, deleting and batching DNS records. FTN should use API tokens with only the permissions required by the registered DDNS targets.

Reference: https://developers.cloudflare.com/api/resources/dns/subresources/records/
