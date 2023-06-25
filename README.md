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
