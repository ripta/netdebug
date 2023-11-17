FROM golang:1.21-bookworm AS build

RUN mkdir /app
WORKDIR /app
COPY . .

RUN go test ./...
RUN go build ./cmd/netdebug

###

FROM gcr.io/distroless/base-debian12
COPY --from=build /app/netdebug /app/netdebug
RUN ["/app/netdebug", "version"]
ENTRYPOINT ["/app/netdebug"]
