netdebug -- a collection of network debug tooling

Subcommands:

  - `dns` performs a DNS query
  - `echo` runs an HTTP or gRPC echo server.
  - `listen` runs a TCP server.
  - `send` creates a TCP connection and sends a payload.

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

# Run gRPC with self-signed certificate, port 8443
netdebug echo --tls-autogenerate --port=8443 --mode=grpc
```

gRPC mode has server reflection enabled, and thus you can:

```
grpcurl -insecure localhost:8443 list
grpcurl -insecure -d '{"query":"deadbeef"}' -rpc-header 'x-foo-id: bar-123' localhost:8443 pkg.echo.v1.Echoer.Echo
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
