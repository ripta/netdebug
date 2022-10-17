FROM golang:1.19-bullseye AS build

RUN mkdir /app
WORKDIR /app
COPY . .

RUN go test ./...
RUN go build ./cmd/netdebug

###

FROM debian:bullseye-slim
COPY --from=build /app/netdebug /app/netdebug
ENTRYPOINT ["/app/netdebug"]
