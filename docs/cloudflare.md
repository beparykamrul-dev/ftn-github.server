# FTN Cloudflare Integration

FTN keeps the Cloudflare documentation/source tree as a versioned reference and maps selected Cloudflare capabilities into the FTN control/deployment registry.

## Managed capability groups

- DNS records and bulk DNS operations
- DNSSEC and multi-signer DNSSEC references
- Zones and zone settings
- Workers and Workers platform projects
- KV, D1 and R2 integration references
- Pages/build references
- Tunnels and private-origin connectivity references
- Load balancing and health checks
- Rules, caching and traffic controls
- IP geolocation and edge/colo information

## FTN rule

Cloudflare is a provider integration. Provider source is referenced; proprietary provider code is not copied into FTN service repositories.

Credentials, API tokens, zone IDs and account identifiers stay outside Git.

## DNS API model

The FTN DDNS layer can use Cloudflare DNS record APIs for A/AAAA updates. The registry should identify the zone and record at runtime; secrets are injected by the deployment environment.

## GeoFlow

Cloudflare Workers can expose request geolocation and edge information such as country, city, latitude, longitude, timezone and colo. FTN GeoFlow consumes these signals as routing/observability inputs rather than storing unnecessary raw visitor data.

## References

- https://developers.cloudflare.com/dns/
- https://developers.cloudflare.com/dns/manage-dns-records/
- https://developers.cloudflare.com/api/resources/dns/subresources/records/
- https://developers.cloudflare.com/api/resources/dns/
- https://developers.cloudflare.com/workers/runtime-apis/request/
- https://developers.cloudflare.com/workers/examples/geolocation-hello-world/
- https://developers.cloudflare.com/network/ip-geolocation/
