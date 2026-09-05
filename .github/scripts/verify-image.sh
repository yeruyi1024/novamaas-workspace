#!/usr/bin/env bash
set -euo pipefail

image=${1:?Pass the Docker image reference}
expected_version=${2:?Pass the expected application version}

test "$(docker run --rm "$image" --version)" = "$expected_version"
docker run -d --name novamaas-smoke -p 127.0.0.1:3000:3000 "$image"
trap 'docker logs novamaas-smoke; docker rm -f novamaas-smoke' EXIT

for attempt in $(seq 1 90); do
  if curl -fsS http://127.0.0.1:3000/api/status > /tmp/novamaas-status.json; then
    jq -e '.success == true' /tmp/novamaas-status.json
    curl -fsS http://127.0.0.1:3000/ > /tmp/novamaas-index.html
    grep -qi '<!doctype html' /tmp/novamaas-index.html
    exit 0
  fi
  sleep 2
done

echo 'Container did not become ready in 180 seconds.' >&2
exit 1
