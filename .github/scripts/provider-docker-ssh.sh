#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# Deploy Provider: docker-ssh
#
# Contract (env vars injected by the composite action):
#   SSH_HOST, SSH_USER, SSH_PORT
#   REGISTRY, REGISTRY_USER, REGISTRY_PASS
#   IMAGE           — full image:tag to deploy
#   COMPOSE_FILE    — path on remote host
#   COMPOSE_DIR     — working directory on remote host
#   SERVICE         — compose service name to restart (empty = all)
#   HC_URL          — healthcheck URL (checked on the remote host)
#   HC_RETRIES      — number of attempts
#   HC_INTERVAL     — seconds between attempts
#
# Stages (matching the provider contract):
#   1. authenticate  — docker login on remote
#   2. prepare       — resolve rollback tag before pull
#   3. deploy        — pull + compose up
#   4. healthcheck   — poll /health
#   5. rollback      — triggered automatically on healthcheck failure
#   6. (notify)      — handled by caller
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

SSH_OPTS="-o StrictHostKeyChecking=no -o BatchMode=yes -o ConnectTimeout=15 \
  -i ~/.ssh/deploy_key -p ${SSH_PORT:-22}"
SSH_TARGET="${SSH_USER}@${SSH_HOST}"

ssh_exec() {
  # Run a command on the remote host.
  # Arguments are passed as a single heredoc to avoid quoting hell.
  local script="$1"
  ssh $SSH_OPTS "$SSH_TARGET" "bash -s" <<EOF
set -euo pipefail
$script
EOF
}

echo "::group::1/5 Authenticate"
# Write registry password via stdin — never in env or command line
ssh_exec "
  echo '${REGISTRY_PASS}' | \
    docker login '${REGISTRY}' --username '${REGISTRY_USER}' --password-stdin
"
echo "::endgroup::"

echo "::group::2/5 Prepare (capture rollback tag)"
PREVIOUS_IMAGE=$(ssh_exec "
  cd '${COMPOSE_DIR}'
  # Read the current image tag from the running container (if any)
  docker compose -f '${COMPOSE_FILE}' ps -q '${SERVICE}' 2>/dev/null | \
    xargs -r docker inspect --format '{{.Config.Image}}' 2>/dev/null | head -1 || true
") || true
echo "Previous image: ${PREVIOUS_IMAGE:-<none>}"
echo "::endgroup::"

echo "::group::3/5 Deploy — pull + rolling restart"
ssh_exec "
  cd '${COMPOSE_DIR}'

  # Export the new image tag so docker-compose.yml can reference \${IMAGE}
  export IMAGE='${IMAGE}'

  # Pull the new image first (zero-downtime preparation)
  docker compose -f '${COMPOSE_FILE}' pull '${SERVICE}'

  # Rolling update: stop → start with new image
  # --no-deps prevents accidentally restarting dependent services
  docker compose -f '${COMPOSE_FILE}' up -d --no-deps --remove-orphans '${SERVICE}'
"
echo "::endgroup::"

echo "::group::4/5 Healthcheck"
HEALTHY=false
for i in $(seq 1 "${HC_RETRIES:-12}"); do
  echo "  attempt $i / ${HC_RETRIES:-12} — checking ${HC_URL}"
  STATUS=$(ssh_exec "curl -fsS -o /dev/null -w '%{http_code}' '${HC_URL}'" 2>/dev/null || echo "000")
  if [ "$STATUS" = "200" ]; then
    echo "  ✓ healthy (HTTP 200)"
    HEALTHY=true
    break
  fi
  echo "  ✗ got HTTP $STATUS — waiting ${HC_INTERVAL:-5}s"
  sleep "${HC_INTERVAL:-5}"
done
echo "::endgroup::"

# ── Rollback hook ─────────────────────────────────────────────────────────────
if [ "$HEALTHY" != "true" ]; then
  echo "::error::Healthcheck failed after ${HC_RETRIES} attempts — triggering rollback"

  if [ -n "${PREVIOUS_IMAGE:-}" ]; then
    echo "::group::5/5 Rollback → ${PREVIOUS_IMAGE}"
    ssh_exec "
      cd '${COMPOSE_DIR}'
      export IMAGE='${PREVIOUS_IMAGE}'
      docker compose -f '${COMPOSE_FILE}' pull '${SERVICE}'
      docker compose -f '${COMPOSE_FILE}' up -d --no-deps '${SERVICE}'
    " || echo "::warning::Rollback also failed — manual intervention required"
    echo "::endgroup::"
  else
    echo "::warning::No previous image recorded — cannot rollback automatically"
  fi

  exit 1
fi

echo "Deploy complete: ${IMAGE}"
