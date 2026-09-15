# FTN GeoFlow

GeoFlow is the FTN geographic and edge-flow layer for correlating service traffic with region, edge/colo, node health and latency.

## Inputs

- Cloudflare Worker request.cf signals
- CF-IPCountry / visitor-location headers when enabled
- FTN node and zone metadata
- DNS endpoint state
- Health-check results
- RTT/latency measurements

## Core fields

`country`, `continent`, `region`, `city`, `latitude`, `longitude`, `timezone`, `colo`, `clientTcpRtt`, and `clientQuicRtt` when available.

## Processing

GeoFlow normalizes location and network observations into:

`visitor -> country/region -> colo -> provider -> FTN zone -> node -> service`

It can then feed:

- global service map
- latency map
- edge map
- DNS endpoint map
- provider/zone status
- service availability summary

## Routing principle

GeoFlow is an observation and routing-input layer. It should not permanently bind a service to one geographic location. Health, latency and service availability remain independent dimensions.

## Privacy

Use the minimum location data required for the selected feature. Do not export API tokens, authentication material or unrelated request payloads into telemetry.

## Cloudflare source

Cloudflare Workers expose geolocation and edge properties through `request.cf`, including country, city, latitude, longitude, timezone and colo. Cloudflare also documents IP geolocation headers for origins.

References:

- https://developers.cloudflare.com/workers/runtime-apis/request/
- https://developers.cloudflare.com/workers/examples/geolocation-hello-world/
- https://developers.cloudflare.com/network/ip-geolocation/
