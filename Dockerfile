FROM golang:1.21-bookworm AS build

RUN mkdir /app
WORKDIR /app
COPY . .

RUN go test ./...
RUN go build ./cmd/netdebug

###

FROM gcr.io/distroless/static-debian12
COPY --from=build /app/netdebug /app/netdebug
ENTRYPOINT ["/app/netdebug"]
