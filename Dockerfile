FROM golang:1.21-bookworm AS build

RUN mkdir /app
WORKDIR /app
COPY . .

RUN go test ./...
RUN go build ./cmd/netdebug

###

FROM debian:bookworm-slim
COPY --from=build /app/netdebug /app/netdebug
ENTRYPOINT ["/app/netdebug"]
