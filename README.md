netdebug -- a collection of network debug tooling

Subcommands:

  - `dns` performs a DNS query
  - `echo` runs an HTTP or gRPC echo server.
  - `listen` runs a TCP server.
  - `send` creates a TCP connection and sends a payload.
  - `bench` benchmarks a gRPC echo server.

## `dns`

```
netdebug dns -t mx github.com
netdebug dns -d 1.1.1.1:53 -t mx github.com
```

## `echo`

```
# Run with default settings: HTTP/1, port 8080
netdebug echo

# Listen only on 127.0.0.1
netdebug echo --host=127.0.0.1

# Run HTTP/2 with self-signed certificate, port 8443
netdebug echo --tls-autogenerate --port=8443
```

gRPC mode has server reflection enabled, and thus you can:

```
# Run gRPC with self-signed certificate, port 8443
netdebug echo --tls-autogenerate --port=8443 --mode=grpc

# and then...
grpcurl -insecure localhost:8443 list
```

### `pkg.echo.v1.Echoer/Echo`

```
# Run gRPC with self-signed certificate, port 8443
netdebug echo --tls-autogenerate --port=8443 --mode=grpc

grpcurl -insecure -d '{"query":"deadbeef"}' -rpc-header 'x-foo-id: bar-123' localhost:8443 pkg.echo.v1.Echoer.Echo
```

There's also a grpc and HTTP mode:

```
# Run gRPC and HTTP with self-signed certificate, port 8443
netdebug echo --tls-autogenerate --port=8443 --mode=grpc+http

# gRPCurl will still work:
grpcurl -insecure localhost:8443 list

# as will regular curl with HTTP/1 and HTTP/2
curl --http1.1 -k https://localhost:8443/
curl --http2 -k https://localhost:8443/
```

In fact, with the help of `protoc`, `grpcto`, and `curl`, you can hand-craft a gRPC request over HTTPS:

```
# Run gRPC and HTTP with self-signed certificate, port 8443
netdebug echo --tls-autogenerate --port=8443 --mode=grpc+http

echo 'query:"fizzbuzz"' \
    | protoc --encode pkg.echo.v1.EchoRequest ./pkg/echo/v1/echo.proto \
    | grpcto frame \
    | curl -X POST -k -v --data-binary @- -H "Content-Type: application/grpc" --raw https://localhost:8443/pkg.echo.v1.Echoer/Echo \
    | grpcto unframe \
    | protoc --decode_raw
```

or HTTP:

```
# Run gRPC and HTTP
netdebug echo --port=8080 --mode=grpc+http

echo 'query:"fizzbuzz"' \
    | protoc --encode pkg.echo.v1.EchoRequest ./pkg/echo/v1/echo.proto \
    | grpcto frame \
    | curl -X POST -v --data-binary @- -H "Content-Type: application/grpc" --raw --http2 --http2-prior-knowledge http://localhost:8443/pkg.echo.v1.Echoer/Echo \
    | grpcto unframe \
    | protoc --decode_raw
```

### `pkg.echo.v1.Echoer/Status`

There's also a `Status` procedure that allows you to request that the server return arbitrary errors:

```
# Run gRPC
netdebug echo --port=8080 --mode=grpc

grpcurl -v -plaintext -d '{"force_grpc_status":3,"message":"oopsie!"}' localhost:8080 pkg.echo.v1.Echoer/Status
```

which might return something like:

```
Resolved method descriptor:
rpc Status ( .pkg.echo.v1.StatusRequest ) returns ( .pkg.echo.v1.StatusResponse );

Request metadata to send:
(empty)

Response headers received:
content-type: application/grpc
trailer: Grpc-Status
trailer: Grpc-Message
trailer: Grpc-Status-Details-Bin

Response trailers received:
(empty)
Sent 1 request and received 0 responses
ERROR:
  Code: InvalidArgument
  Message: oopsie!
```

## `listen` and `send`

Start listening in one terminal:

```
netdebug listen --port=20202
```

and send stuff to it in a different terminal:

```
date | netdebug send --address=localhost:20202
```

## `bench`

`netdebug bench` drives load against `pkg.echo.v1.Echoer/Echo` and prints a
summary covering throughput, latency, server-vs-network time, per-backend
breakdown, error grouping, and compression effectiveness. Start a target with
`netdebug echo --mode=grpc` on one side and run bench on the other. For TLS
targets that use a self-signed certificate, pass `--tls-insecure-skip-verify`;
omit it to verify against the system trust store.

### Payload-shape compare

Same wire size, different proto-decoding cost. With `--embedding-dim=1024`
both shapes carry 4 KiB of payload; `embedding-float` is a packed `repeated
float`, `embedding-bytes` is raw little-endian float32. The gap between the
two runs is the proto-decoding overhead.

```
# Start an echo server in one terminal:
netdebug echo --mode=grpc --port=8080

# In another terminal, compare proto-float decoding vs raw bytes:
netdebug bench --target=127.0.0.1:8080 --payload=embedding-float \
    --embedding-dim=1024 --concurrency=4 --duration=15s

netdebug bench --target=127.0.0.1:8080 --payload=embedding-bytes \
    --embedding-dim=1024 --concurrency=4 --duration=15s
```

Compare the `Latency (server)` and `Latency (network)` percentiles between
the two runs.

### Compression compare

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

### Conn-model compare

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

### Paired mesh-on / mesh-off

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
