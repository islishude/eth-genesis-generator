# syntax=docker/dockerfile:1
FROM golang:1.26.2 AS compiler
RUN apt-get update \
	&& apt-get install -y --no-install-recommends make gcc g++ git ca-certificates \
	&& rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build make install

FROM gcr.io/distroless/cc-debian13:latest
COPY --from=compiler /go/bin/eth-genesis-generator /usr/local/bin/
ENTRYPOINT [ "/usr/local/bin/eth-genesis-generator" ]
