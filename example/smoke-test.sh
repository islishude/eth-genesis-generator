#!/usr/bin/env bash
set -euo pipefail

example_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
workdir=$(mktemp -d /tmp/ethgen-smoke.XXXXXXXX)
workdir=$(cd "$workdir" && pwd -P)
project="ethgen-smoke-$$"

init_compose() {
  PWD="$workdir" docker compose -p "$project-init" -f "$example_dir/docker-compose.init.yml" "$@"
}

node_compose() {
  PWD="$workdir" docker compose -p "$project" -f "$example_dir/docker-compose.yml" "$@"
}

cleanup() {
  status=$?
  if (( status != 0 )); then
    init_compose logs --no-color || true
    node_compose logs --no-color || true
  fi
  node_compose down --timeout 5 || true
  init_compose down --timeout 5 || true
  # On Linux the clients may create directories owned by root.
  if ! rm -rf "$workdir" 2>/dev/null; then
    docker run --rm --mount "type=bind,src=$workdir,target=/smoke" \
      alpine:3.24.4 rm -rf /smoke/devnet
    rmdir "$workdir"
  fi
}
trap cleanup EXIT

init_compose up --abort-on-container-failure "${1:---build}"
node_compose up --detach --wait --wait-timeout 90

deadline=$((SECONDS + 180))
while (( SECONDS < deadline )); do
  if curl --silent --fail --max-time 3 http://127.0.0.1:8545 \
    --header 'Content-Type: application/json' \
    --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' |
    jq -e '.result | strings | test("^0x[1-9a-f][0-9a-f]*$")' >/dev/null; then
    echo "The example devnet produced a block after genesis."
    exit 0
  fi
  sleep 2
done

echo "Timed out waiting for the example devnet to produce a block." >&2
exit 1
