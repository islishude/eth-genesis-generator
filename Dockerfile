# syntax=docker/dockerfile:1
FROM golang:1.25.3-alpine AS compiler
RUN apk add --no-cache make gcc musl-dev linux-headers git ca-certificates
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build go install

FROM alpine:latest
RUN apk add --no-cache jq
COPY --from=compiler /go/bin/* /usr/local/bin/
ENTRYPOINT [ "eth-genesis-generator" ]
