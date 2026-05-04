# syntax=docker/dockerfile:1
FROM golang:1.26.2-alpine AS compiler
RUN apk add --no-cache make gcc musl-dev linux-headers git ca-certificates g++ libstdc++
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build make install

FROM gcr.io/distroless/base-debian13:latest
COPY --from=compiler /go/bin/eth-genesis-generator /usr/local/bin/
ENTRYPOINT [ "eth-genesis-generator" ]
