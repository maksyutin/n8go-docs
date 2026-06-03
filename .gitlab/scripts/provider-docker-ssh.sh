#!/usr/bin/env bash
# Symlink-equivalent: identical contract to .github/scripts/provider-docker-ssh.sh
# Kept as a separate copy so each platform is self-contained.
set -euo pipefail

SSH_OPTS="-o StrictHostKeyChecking=no -o BatchMode=yes -o ConnectTimeout=15 \
  -i ~/.ssh/deploy_key -p ${SSH_PORT:-22}"
SSH_TARGET="${SSH_USER}@${SSH_HOST}"

ssh_exec() {
  ssh $SSH_OPTS "$SSH_TARGET" "bash -s" <<EOF
set -euo pipefail
$1
EOF
}

echo "==> 1/5 Authenticate"
ssh_exec "echo '${REGISTRY_PASS}' | docker login '${REGISTRY}' --username '${REGISTRY_USER}' --password-stdin"

echo "==> 2/5 Prepare (capture rollback tag)"
PREVIOUS_IMAGE=$(ssh_exec "
  cd '${COMPOSE_DIR:-/opt/app}'
  docker compose -f '${COMPOSE_FILE:-docker-compose.yml}' ps -q '${SERVICE:-app}' 2>/dev/null | \
    xargs -r docker inspect --format '{{.Config.Image}}' 2>/dev/null | head -1 || true
") || true
echo "Previous: ${PREVIOUS_IMAGE:-<none>}"

echo "==> 3/5 Deploy"
ssh_exec "
  cd '${COMPOSE_DIR:-/opt/app}'
  export IMAGE='${IMAGE}'
  docker compose -f '${COMPOSE_FILE:-docker-compose.yml}' pull '${SERVICE:-app}'
  docker compose -f '${COMPOSE_FILE:-docker-compose.yml}' up -d --no-deps --remove-orphans '${SERVICE:-app}'
"

echo "==> 4/5 Healthcheck"
HEALTHY=false
for i in $(seq 1 "${HC_RETRIES:-12}"); do
  echo "  attempt $i / ${HC_RETRIES:-12}"
  STATUS=$(ssh_exec "curl -fsS -o /dev/null -w '%{http_code}' '${HC_URL:-http://localhost:8080/health}'" 2>/dev/null || echo "000")
  if [ "$STATUS" = "200" ]; then
    HEALTHY=true
    break
  fi
  sleep "${HC_INTERVAL:-5}"
done

if [ "$HEALTHY" != "true" ]; then
  echo "ERROR: healthcheck failed — rolling back"
  if [ -n "${PREVIOUS_IMAGE:-}" ]; then
    ssh_exec "
      cd '${COMPOSE_DIR:-/opt/app}'
      export IMAGE='${PREVIOUS_IMAGE}'
      docker compose -f '${COMPOSE_FILE:-docker-compose.yml}' up -d --no-deps '${SERVICE:-app}'
    " || echo "WARNING: rollback also failed"
  fi
  exit 1
fi

echo "Deploy complete: ${IMAGE}"
