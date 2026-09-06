#!/usr/bin/env bash
set -euo pipefail

registry=${REGISTRY:?Set REGISTRY to the Docker registry hostname}
username=${REGISTRY_USERNAME:?Set REGISTRY_USERNAME}
password=${REGISTRY_PASSWORD:?Set REGISTRY_PASSWORD}
docker_config=${DOCKER_CONFIG:-"$HOME/.docker"}
config_path="$docker_config/config.json"
temp_path="$config_path.tmp"

# Tencent CCR rejects the scope-less token request used by `docker login`.
# Store the basic credential so push/pull can request a repository-scoped token.
umask 077
mkdir -p "$docker_config"
auth=$(printf '%s:%s' "$username" "$password" | base64 | tr -d '\n')

if [[ -f "$config_path" ]]; then
  jq --arg registry "$registry" --arg auth "$auth" \
    '.auths = (.auths // {}) | .auths[$registry] = {"auth": $auth}' \
    "$config_path" > "$temp_path"
else
  jq -n --arg registry "$registry" --arg auth "$auth" \
    '{"auths": {($registry): {"auth": $auth}}}' > "$temp_path"
fi

mv "$temp_path" "$config_path"
