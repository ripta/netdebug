# `netdebug bench` — local recipes

`netdebug bench` drives load against `pkg.echo.v1.Echoer/Echo` and prints a
summary covering throughput, latency, server-vs-network time, per-backend
breakdown, error grouping, and compression effectiveness. Every recipe below
expects a local echo server running in another terminal:

```
netdebug echo --mode=grpc --port=8080
```

For TLS targets that use a self-signed certificate, add
`--tls-insecure-skip-verify` to the `bench` invocations; omit it to verify
against the system trust store.

For a cluster walkthrough, see [`bench-k8s.md`](bench-k8s.md).

## Payload-shape compare

Same wire size, different proto-decoding cost. With `--embedding-dim=1024`
both shapes carry 4 KiB of payload; `embedding-float` is a packed `repeated
float`, `embedding-bytes` is raw little-endian float32. The gap between the
two runs is the proto-decoding overhead.

```
netdebug bench --target=127.0.0.1:8080 --payload=embedding-float \
    --embedding-dim=1024 --concurrency=4 --duration=15s

netdebug bench --target=127.0.0.1:8080 --payload=embedding-bytes \
    --embedding-dim=1024 --concurrency=4 --duration=15s
```

Compare the `Latency (server)` and `Latency (network)` percentiles between
the two runs.

## Compression compare

Fix the payload and vary the codec. The summary's compression-effectiveness
rows report bytes-on-wire vs uncompressed for each direction; pairing that
with the server-time percentiles is how you spot the crossover where codec
CPU cost overtakes the wire-size win.

```
netdebug bench --target=127.0.0.1:8080 --payload=embedding-float \
    --embedding-dim=1024 --concurrency=4 --duration=15s --compression=identity

netdebug bench --target=127.0.0.1:8080 --payload=embedding-float \
    --embedding-dim=1024 --concurrency=4 --duration=15s --compression=gzip

netdebug bench --target=127.0.0.1:8080 --payload=embedding-float \
    --embedding-dim=1024 --concurrency=4 --duration=15s --compression=snappy

netdebug bench --target=127.0.0.1:8080 --payload=embedding-float \
    --embedding-dim=1024 --concurrency=4 --duration=15s --compression=zstd
```

## Conn-model compare

Without a service mesh, kube-proxy load-balances per connection. `shared`
pins every worker to one backend, `per-worker` spreads `--concurrency`
connections through the LB, and `per-request` exercises the full TCP+TLS+HTTP/2
handshake on every call. Read the per-backend breakdown to see the
distribution.

```
netdebug bench --target=127.0.0.1:8080 --concurrency=8 --duration=15s \
    --conn-model=per-worker

netdebug bench --target=127.0.0.1:8080 --concurrency=8 --duration=15s \
    --conn-model=shared

netdebug bench --target=127.0.0.1:8080 --concurrency=8 --duration=15s \
    --conn-model=per-request
```

## Paired mesh-on / mesh-off

Tag each run with `--label` so the JSON summaries are self-identifying when
read or diffed later. Use `--header` to inject any routing or tracing header
the mesh keys off; with `--output=json` the summaries diff cleanly under `jq`.

```
# Mesh-off baseline: talk straight to the service.
netdebug bench --target=echo.default.svc:8080 \
    --payload=embedding-float --embedding-dim=1024 \
    --concurrency=8 --duration=30s \
    --label=mesh=off --label=run=2026-05-17a \
    --output=json > mesh-off.json

# Mesh-on: same target, now via the local sidecar; --header drives
# any header-keyed mesh routing or tracing propagation.
netdebug bench --target=echo.default.svc:8080 \
    --payload=embedding-float --embedding-dim=1024 \
    --concurrency=8 --duration=30s \
    --header=x-trace-id=mesh-on-001 \
    --label=mesh=on --label=run=2026-05-17a \
    --output=json > mesh-on.json
```

The labels travel inside each summary, so the files are self-identifying.
Diff the per-backend breakdowns and the network-and-mesh percentiles to see
the cost the sidecar adds.
