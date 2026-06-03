#!/usr/bin/env bash
# deploy/tests/test_provider_docker_ssh.sh
#
# Contract tests for .github/scripts/provider-docker-ssh.sh and its
# GitLab mirror .gitlab/scripts/provider-docker-ssh.sh.
#
# Run: bash deploy/tests/test_provider_docker_ssh.sh
# Exit code 0 = all passed; non-zero = at least one failure.
#
# Approach: no real SSH is used. Tests either inspect the script statically
# (grep / bash -n) or run it with a fake `ssh` stub on PATH that records
# calls and controls exit codes.
# ─────────────────────────────────────────────────────────────────────────────
# Do NOT set -e here: the test runner must continue after individual failures.
set -uo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
GITHUB_SCRIPT="$REPO_ROOT/.github/scripts/provider-docker-ssh.sh"
GITLAB_SCRIPT="$REPO_ROOT/.gitlab/scripts/provider-docker-ssh.sh"

PASS=0
FAIL=0

pass() { echo "  PASS  $1"; ((PASS++)); }
fail() { echo "  FAIL  $1"; ((FAIL++)); }

run_test() {
    local name="$1"
    local result="$2"   # "pass" or "fail"
    if [ "$result" = "pass" ]; then
        pass "$name"
    else
        fail "$name"
    fi
}

assert_grep() {
    # assert_grep <description> <pattern> <file>
    local name="$1" pattern="$2" file="$3"
    if grep -qE "$pattern" "$file" 2>/dev/null; then
        pass "$name"
    else
        fail "$name — pattern not found: $pattern"
    fi
}

assert_no_grep() {
    local name="$1" pattern="$2" file="$3"
    if grep -qE "$pattern" "$file" 2>/dev/null; then
        fail "$name — forbidden pattern found: $pattern"
    else
        pass "$name"
    fi
}

# ── helpers for functional tests ─────────────────────────────────────────────

# Creates a temporary directory with a fake `ssh` binary that:
#   - Records all calls to SSH_LOG
#   - Returns the exit code in SSH_EXIT_CODE (default 0)
#   - Outputs SSH_STDOUT to stdout
make_ssh_stub() {
    # make_ssh_stub <stub_dir> <exit_code> <hc_http_code>
    # The stub intercepts all SSH calls. When the remote script contains `curl`
    # (the healthcheck call), it emits <hc_http_code> so the provider parses it
    # correctly. All other calls behave as a no-op success.
    local stub_dir="$1"
    local exit_code="${2:-0}"
    local hc_code="${3:-200}"
    mkdir -p "$stub_dir"
    cat > "$stub_dir/ssh" <<STUB
#!/usr/bin/env bash
# Record the call for later inspection.
echo "\$@" >> "\${SSH_LOG:-/dev/null}"
# Read the heredoc script from stdin.
script="\$(cat)"
echo "\$script" >> "\${SSH_LOG:-/dev/null}.scripts"
# If the script contains a curl healthcheck, emit the expected http_code.
if echo "\$script" | grep -q 'curl'; then
    echo "${hc_code}"
fi
exit ${exit_code}
STUB
    chmod +x "$stub_dir/ssh"
}

run_provider() {
    # run_provider <script> <stub_dir> [extra env vars...]
    # Runs the provider script with required env vars set to safe defaults.
    local script="$1"; shift
    local stub_dir="$1"; shift
    local log_file="$stub_dir/ssh.log"

    SSH_LOG="$log_file" \
    PATH="$stub_dir:$PATH" \
    SSH_HOST="test.host" \
    SSH_USER="deploy" \
    SSH_PORT="22" \
    REGISTRY="ghcr.io" \
    REGISTRY_USER="ci" \
    REGISTRY_PASS="secret" \
    IMAGE="ghcr.io/owner/app:sha-abc123" \
    COMPOSE_FILE="docker-compose.yml" \
    COMPOSE_DIR="/opt/app" \
    SERVICE="app" \
    HC_URL="http://localhost:8080/health" \
    HC_RETRIES="1" \
    HC_INTERVAL="0" \
    "$@" \
    bash "$script" 2>&1
}

# ═════════════════════════════════════════════════════════════════════════════
# GROUP 1: Static analysis — syntax
# ═════════════════════════════════════════════════════════════════════════════
echo
echo "── Static: syntax ──────────────────────────────────────────────────────"

# Prevents: shipping a broken shell script that fails on first execution.
name="GitHub script passes bash -n"
if bash -n "$GITHUB_SCRIPT" 2>/dev/null; then pass "$name"; else fail "$name"; fi

name="GitLab script passes bash -n"
if bash -n "$GITLAB_SCRIPT" 2>/dev/null; then pass "$name"; else fail "$name"; fi

# Prevents: losing set -euo pipefail and silently ignoring errors.
# A provider that swallows errors will report success on broken deploys.
assert_grep "GitHub script has set -euo pipefail" "set -euo pipefail" "$GITHUB_SCRIPT"
assert_grep "GitLab script has set -euo pipefail" "set -euo pipefail" "$GITLAB_SCRIPT"

# ═════════════════════════════════════════════════════════════════════════════
# GROUP 2: Static analysis — required env var contract
# Each var is part of the documented provider contract; if any disappears
# the deploy action must be updated too.
# ═════════════════════════════════════════════════════════════════════════════
echo
echo "── Static: required env vars ───────────────────────────────────────────"

for var in SSH_HOST SSH_USER SSH_PORT \
           REGISTRY REGISTRY_USER REGISTRY_PASS \
           IMAGE COMPOSE_FILE COMPOSE_DIR SERVICE \
           HC_URL HC_RETRIES HC_INTERVAL; do
    # Prevents: removing a var from the script while the deploy action still sets it.
    assert_grep "GitHub script references \$${var}" "\\\$\{?${var}" "$GITHUB_SCRIPT"
    assert_grep "GitLab script references \$${var}" "\\\$\{?${var}" "$GITLAB_SCRIPT"
done

# ═════════════════════════════════════════════════════════════════════════════
# GROUP 3: Static analysis — 5-stage contract
# The action.yml documents stages 1-5; renaming a stage breaks operator runbooks.
# ═════════════════════════════════════════════════════════════════════════════
echo
echo "── Static: 5-stage contract ─────────────────────────────────────────────"

for script in "$GITHUB_SCRIPT" "$GITLAB_SCRIPT"; do
    name_prefix="$(basename "$(dirname "$script")")/$(basename "$script")"
    # Prevents: reordering or renaming stages without updating DEPLOY.md / runbooks.
    assert_grep "$name_prefix stage 1 Authenticate"  "(1/5|==> 1)" "$script"
    assert_grep "$name_prefix stage 2 Prepare"       "(2/5|==> 2)" "$script"
    assert_grep "$name_prefix stage 3 Deploy"        "(3/5|==> 3)" "$script"
    assert_grep "$name_prefix stage 4 Healthcheck"   "(4/5|==> 4)" "$script"
    assert_grep "$name_prefix stage 5 Rollback"      "(5/5|==> 5|[Rr]ollback)" "$script"
done

# ═════════════════════════════════════════════════════════════════════════════
# GROUP 4: Static analysis — security constraints
# ═════════════════════════════════════════════════════════════════════════════
echo
echo "── Static: security constraints ─────────────────────────────────────────"

# Prevents: passing registry password on the command line (leaks in ps/logs).
# Password must go via stdin (--password-stdin).
assert_grep "GitHub script uses --password-stdin" "\-\-password-stdin" "$GITHUB_SCRIPT"
assert_grep "GitLab script uses --password-stdin" "\-\-password-stdin" "$GITLAB_SCRIPT"

# Prevents: removing StrictHostKeyChecking=no which would cause interactive
# prompts in CI and hang the pipeline forever.
assert_grep "GitHub script sets StrictHostKeyChecking=no" "StrictHostKeyChecking=no" "$GITHUB_SCRIPT"
assert_grep "GitLab script sets StrictHostKeyChecking=no" "StrictHostKeyChecking=no" "$GITLAB_SCRIPT"

# Prevents: adding -o StrictHostKeyChecking=yes which breaks fresh hosts.
assert_no_grep "GitHub script does not set StrictHostKeyChecking=yes" "StrictHostKeyChecking=yes" "$GITHUB_SCRIPT"
assert_no_grep "GitLab script does not set StrictHostKeyChecking=yes" "StrictHostKeyChecking=yes" "$GITLAB_SCRIPT"

# ═════════════════════════════════════════════════════════════════════════════
# GROUP 5: Static analysis — rolling update strategy
# Changing flags alters downtime characteristics without any other indicator.
# ═════════════════════════════════════════════════════════════════════════════
echo
echo "── Static: rolling update flags ─────────────────────────────────────────"

# Prevents: removing --no-deps (would restart dependent services).
assert_grep "GitHub script uses --no-deps"      "\-\-no-deps"      "$GITHUB_SCRIPT"
assert_grep "GitLab script uses --no-deps"      "\-\-no-deps"      "$GITLAB_SCRIPT"

# Prevents: removing --remove-orphans (would leave stale containers running).
assert_grep "GitHub script uses --remove-orphans" "\-\-remove-orphans" "$GITHUB_SCRIPT"
assert_grep "GitLab script uses --remove-orphans" "\-\-remove-orphans" "$GITLAB_SCRIPT"

# ═════════════════════════════════════════════════════════════════════════════
# GROUP 6: Static analysis — rollback guard
# ═════════════════════════════════════════════════════════════════════════════
echo
echo "── Static: rollback guard ───────────────────────────────────────────────"

# Prevents: removing the PREVIOUS_IMAGE capture that enables rollback.
assert_grep "GitHub script captures PREVIOUS_IMAGE" "PREVIOUS_IMAGE" "$GITHUB_SCRIPT"
assert_grep "GitLab script captures PREVIOUS_IMAGE" "PREVIOUS_IMAGE" "$GITLAB_SCRIPT"

# Prevents: removing the rollback block entirely (healthcheck failure would
# leave the service on a broken image with no recovery path).
assert_grep "GitHub script has rollback on healthcheck failure" "HEALTHY.*true|rollback" "$GITHUB_SCRIPT"
assert_grep "GitLab script has rollback on healthcheck failure" "HEALTHY.*true|rollback" "$GITLAB_SCRIPT"

# Prevents: unconditionally rolling back even when PREVIOUS_IMAGE is empty.
# Rolling back to an empty image tag causes docker compose to fail.
assert_grep "GitHub script guards rollback with PREVIOUS_IMAGE check" "\\\$\{?PREVIOUS_IMAGE" "$GITHUB_SCRIPT"
assert_grep "GitLab script guards rollback with PREVIOUS_IMAGE check" "\\\$\{?PREVIOUS_IMAGE" "$GITLAB_SCRIPT"

# ═════════════════════════════════════════════════════════════════════════════
# GROUP 7: Functional — healthy deploy exits 0
# ═════════════════════════════════════════════════════════════════════════════
echo
echo "── Functional: successful deploy ────────────────────────────────────────"

for script in "$GITHUB_SCRIPT" "$GITLAB_SCRIPT"; do
    name_prefix="$(basename "$(dirname "$script")")/$(basename "$script")"
    stub_dir="$(mktemp -d)"
    make_ssh_stub "$stub_dir" 0 200   # ssh succeeds; healthcheck emits 200

    # Prevents: provider exiting non-zero on a clean deploy (false failure).
    name="$name_prefix: healthy deploy exits 0"
    if run_provider "$script" "$stub_dir" > /dev/null 2>&1; then
        pass "$name"
    else
        fail "$name (exit code $?)"
    fi
    rm -rf "$stub_dir"
done

# ═════════════════════════════════════════════════════════════════════════════
# GROUP 8: Functional — unhealthy deploy exits 1
# ═════════════════════════════════════════════════════════════════════════════
echo
echo "── Functional: failed healthcheck → exit 1 ──────────────────────────────"

for script in "$GITHUB_SCRIPT" "$GITLAB_SCRIPT"; do
    name_prefix="$(basename "$(dirname "$script")")/$(basename "$script")"
    stub_dir="$(mktemp -d)"
    make_ssh_stub "$stub_dir" 0 503   # healthcheck emits 503

    # Prevents: provider returning 0 when the service fails health checks.
    # A zero exit on broken deploy causes CI to mark the pipeline green.
    name="$name_prefix: unhealthy deploy exits 1"
    if run_provider "$script" "$stub_dir" > /dev/null 2>&1; then
        fail "$name (expected exit 1, got 0)"
    else
        pass "$name"
    fi
    rm -rf "$stub_dir"
done

# ═════════════════════════════════════════════════════════════════════════════
# GROUP 9: Functional — SSH is called for each stage
# ═════════════════════════════════════════════════════════════════════════════
echo
echo "── Functional: SSH call count ───────────────────────────────────────────"

for script in "$GITHUB_SCRIPT" "$GITLAB_SCRIPT"; do
    name_prefix="$(basename "$(dirname "$script")")/$(basename "$script")"
    stub_dir="$(mktemp -d)"
    log="$stub_dir/ssh.log"
    make_ssh_stub "$stub_dir" 0 200

    run_provider "$script" "$stub_dir" > /dev/null 2>&1 || true
    ssh_calls=0
    [ -f "$log" ] && ssh_calls=$(wc -l < "$log" || echo 0)

    # Prevents: collapsing multiple SSH calls into one (would break stage isolation
    # and make it impossible to identify which stage failed).
    # Stages 1 (auth) + 2 (rollback-tag) + 3 (deploy) + 4 (healthcheck) = min 4 calls.
    name="$name_prefix: at least 4 SSH calls on healthy deploy"
    if [ "${ssh_calls:-0}" -ge 4 ]; then
        pass "$name"
    else
        fail "$name — got ${ssh_calls:-0} SSH call(s), want ≥ 4"
    fi
    rm -rf "$stub_dir"
done

# ═════════════════════════════════════════════════════════════════════════════
# GROUP 10: Functional — rollback is attempted on healthcheck failure
# ═════════════════════════════════════════════════════════════════════════════
echo
echo "── Functional: rollback on unhealthy deploy ─────────────────────────────"

for script in "$GITHUB_SCRIPT" "$GITLAB_SCRIPT"; do
    name_prefix="$(basename "$(dirname "$script")")/$(basename "$script")"
    stub_dir="$(mktemp -d)"
    log="$stub_dir/ssh.log"

    # ssh stub: first call (rollback-tag capture) returns a known image,
    # subsequent calls succeed. This lets the rollback path execute.
    call_count=0
    cat > "$stub_dir/ssh" <<'STUB'
#!/usr/bin/env bash
echo "$@" >> "${SSH_LOG:-/dev/null}"
# First call = rollback-tag capture; emit a known previous image
count_file="${SSH_LOG%.log}.count"
n=$(cat "$count_file" 2>/dev/null || echo 0)
n=$((n + 1))
echo "$n" > "$count_file"
if [ "$n" -eq 1 ]; then
    echo "ghcr.io/owner/app:sha-prev999"
fi
exit 0
STUB
    chmod +x "$stub_dir/ssh"

    output="$(run_provider "$script" "$stub_dir" 2>&1 || true)"

    # Prevents: skipping the rollback command when a previous image exists.
    # Without rollback a broken deploy leaves the service down indefinitely.
    name="$name_prefix: rollback image appears in SSH calls after failed healthcheck"
    if echo "$output" | grep -qiE "rollback|previous|prev999|sha-prev" || \
       ([ -f "$log" ] && grep -qiE "prev999|sha-prev|rollback" "$log"); then
        pass "$name"
    else
        fail "$name — rollback not triggered; check PREVIOUS_IMAGE path"
    fi
    rm -rf "$stub_dir"
done

# ═════════════════════════════════════════════════════════════════════════════
# Summary
# ═════════════════════════════════════════════════════════════════════════════
echo
echo "────────────────────────────────────────────────────────────────────────"
echo "Results: ${PASS} passed, ${FAIL} failed"
echo "────────────────────────────────────────────────────────────────────────"
[ "$FAIL" -eq 0 ]
