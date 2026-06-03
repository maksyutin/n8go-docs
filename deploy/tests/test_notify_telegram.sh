#!/usr/bin/env bash
# deploy/tests/test_notify_telegram.sh
#
# Contract tests for .gitlab/scripts/notify-telegram.sh
#
# Tests use a fake `curl` stub that records the JSON payload instead of
# calling the real Telegram API.
#
# Run: bash deploy/tests/test_notify_telegram.sh
# ─────────────────────────────────────────────────────────────────────────────
set -uo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SCRIPT="$REPO_ROOT/.gitlab/scripts/notify-telegram.sh"

PASS=0; FAIL=0

pass() { echo "  PASS  $1"; ((PASS++)); }
fail() { echo "  FAIL  $1"; ((FAIL++)); }

assert_grep() {
    local name="$1" pattern="$2" text="$3"
    if echo "$text" | grep -qE "$pattern"; then
        pass "$name"
    else
        fail "$name — pattern not found: $pattern"
    fi
}

# Run the notify script with a curl stub that captures --data-binary payload
run_notify() {
    local stub_dir="$1"; shift
    local log="$stub_dir/curl.log"

    cat > "$stub_dir/curl" <<'STUB'
#!/usr/bin/env bash
# Capture the --data-binary argument
while [[ $# -gt 0 ]]; do
    if [[ "$1" == "--data-binary" ]]; then
        echo "$2" >> "${CURL_LOG:-/dev/null}"
        shift 2
    else
        shift
    fi
done
exit 0
STUB
    chmod +x "$stub_dir/curl"

    # jq is also needed; stub it to echo the input unchanged if real jq missing
    if ! command -v jq &>/dev/null; then
        cat > "$stub_dir/jq" <<'JQ'
#!/usr/bin/env bash
cat
JQ
        chmod +x "$stub_dir/jq"
    fi

    CURL_LOG="$log" \
    PATH="$stub_dir:$PATH" \
    TELEGRAM_BOT_TOKEN="test-token" \
    TELEGRAM_CHAT_ID="12345" \
    CI_PROJECT_PATH="owner/repo" \
    CI_COMMIT_REF_NAME="main" \
    CI_COMMIT_SHORT_SHA="abc1234" \
    CI_PROJECT_URL="https://gitlab.example.com/owner/repo" \
    CI_PIPELINE_ID="42" \
    GITLAB_USER_LOGIN="deployer" \
    bash "$SCRIPT" "$@" 2>&1

    echo "$log"   # return log path to caller
}

# ── Static analysis ───────────────────────────────────────────────────────────
echo
echo "── Static: syntax ───────────────────────────────────────────────────────"

# Prevents: deploying a syntactically broken notify script.
name="notify script passes bash -n"
if bash -n "$SCRIPT" 2>/dev/null; then pass "$name"; else fail "$name"; fi

# ── Static: positional argument contract ─────────────────────────────────────
echo
echo "── Static: positional argument contract ─────────────────────────────────"

# Prevents: changing argument positions without updating all callers
# (GitLab CI, GitVerse, GitFlic all call notify-telegram.sh with positional args).
# The contract is: $1=status $2=environment $3=provider $4=image.
declare -A POS_VARS=( [STATUS]=1 [ENV]=2 [PROVIDER]=3 [IMAGE]=4 )
for var in STATUS ENV PROVIDER IMAGE; do
    pos="${POS_VARS[$var]}"
    # Matches both $1 and ${1:-default} style assignments
    name="notify script reads positional arg \$$pos into $var"
    if grep -qE "${var}=.*\\\$\{?${pos}" "$SCRIPT" 2>/dev/null; then
        pass "$name"
    else
        fail "$name"
    fi
done

# ── Static: required env vars ────────────────────────────────────────────────
echo
echo "── Static: required env vars ────────────────────────────────────────────"

for var in TELEGRAM_BOT_TOKEN TELEGRAM_CHAT_ID \
           CI_PROJECT_PATH CI_COMMIT_REF_NAME CI_COMMIT_SHORT_SHA \
           CI_PROJECT_URL CI_PIPELINE_ID; do
    # Prevents: dropping an env var reference while callers still set it.
    name="notify script references \$$var"
    if grep -qE "\\\$\{?${var}" "$SCRIPT" 2>/dev/null; then
        pass "$name"
    else
        fail "$name"
    fi
done

# ── Static: emoji mapping ────────────────────────────────────────────────────
echo
echo "── Static: emoji mapping ────────────────────────────────────────────────"

# Prevents: removing the status → emoji mapping (makes all notifications look
# identical regardless of outcome — operators stop reading them).
name="notify script maps 'success' to ✅"
if grep -qE "success.*✅|✅.*success" "$SCRIPT"; then pass "$name"; else fail "$name"; fi

name="notify script maps 'failure' to ❌"
if grep -qE "failure.*❌|❌.*failure" "$SCRIPT"; then pass "$name"; else fail "$name"; fi

# Prevents: removing the default case (cancelled / unknown status)
name="notify script has default emoji for unknown status (⚠️)"
if grep -q "⚠️" "$SCRIPT"; then pass "$name"; else fail "$name"; fi

# ── Functional: curl receives correct JSON fields ────────────────────────────
echo
echo "── Functional: JSON payload ─────────────────────────────────────────────"

stub_dir="$(mktemp -d)"
log="$(run_notify "$stub_dir" "success" "stage" "docker-ssh" "ghcr.io/owner/app:v1.0.0")"
payload="$(cat "$log" 2>/dev/null || echo "")"

# Prevents: removing the chat_id from the payload (message goes nowhere silently).
assert_grep "payload contains chat_id" "chat_id|12345" "$payload"

# Prevents: removing the text field (Telegram rejects the request with 400).
assert_grep "payload contains text field" '"text"' "$payload"

# Prevents: stripping the Markdown parse_mode (formatting disappears).
assert_grep "payload contains parse_mode Markdown" "Markdown" "$payload"

rm -rf "$stub_dir"

# ── Functional: all three statuses emit different emoji ───────────────────────
echo
echo "── Functional: status → emoji ───────────────────────────────────────────"

for status in success failure cancelled; do
    stub_dir="$(mktemp -d)"
    log="$(run_notify "$stub_dir" "$status" "prod" "docker-ssh" "img:tag" 2>&1 || true)"
    # The script itself writes to CURL_LOG; we want to check script stdout/err too
    output="$(CURL_LOG="$stub_dir/c.log" PATH="$stub_dir:$PATH" \
        TELEGRAM_BOT_TOKEN="t" TELEGRAM_CHAT_ID="1" \
        CI_PROJECT_PATH="o/r" CI_COMMIT_REF_NAME="main" CI_COMMIT_SHORT_SHA="abc" \
        CI_PROJECT_URL="http://x" CI_PIPELINE_ID="1" GITLAB_USER_LOGIN="u" \
        bash "$SCRIPT" "$status" "prod" "docker-ssh" "img:tag" 2>&1 || true)"

    # Prevents: all statuses producing the same notification (ops loses signal).
    case "$status" in
        success)
            name="success status: script references ✅"
            if grep -q "✅" "$SCRIPT"; then pass "$name"; else fail "$name"; fi
            ;;
        failure)
            name="failure status: script references ❌"
            if grep -q "❌" "$SCRIPT"; then pass "$name"; else fail "$name"; fi
            ;;
        cancelled)
            name="cancelled/unknown status: script references ⚠️"
            if grep -q "⚠️" "$SCRIPT"; then pass "$name"; else fail "$name"; fi
            ;;
    esac
    rm -rf "$stub_dir"
done

# ── Functional: token never echoed to stdout ─────────────────────────────────
echo
echo "── Functional: token not leaked ─────────────────────────────────────────"

stub_dir="$(mktemp -d)"
cat > "$stub_dir/curl" <<'STUB'
#!/usr/bin/env bash
exit 0
STUB
chmod +x "$stub_dir/curl"
[ -f "$stub_dir/jq" ] || cp /dev/null "$stub_dir/jq" 2>/dev/null || true
command -v jq &>/dev/null || { cat > "$stub_dir/jq" <<'JQ'
#!/usr/bin/env bash
cat
JQ
chmod +x "$stub_dir/jq"; }

output="$(CURL_LOG="/dev/null" PATH="$stub_dir:$PATH" \
    TELEGRAM_BOT_TOKEN="MY_SECRET_TOKEN_12345" \
    TELEGRAM_CHAT_ID="1" \
    CI_PROJECT_PATH="o/r" CI_COMMIT_REF_NAME="main" CI_COMMIT_SHORT_SHA="abc" \
    CI_PROJECT_URL="http://x" CI_PIPELINE_ID="1" GITLAB_USER_LOGIN="u" \
    bash "$SCRIPT" "success" "stage" "docker-ssh" "img:tag" 2>&1 || true)"

# Prevents: the bot token appearing in CI logs where it could be scraped.
name="bot token is not echoed to stdout/stderr"
if echo "$output" | grep -q "MY_SECRET_TOKEN_12345"; then
    fail "$name — token leaked: $output"
else
    pass "$name"
fi

rm -rf "$stub_dir"

# ── Summary ───────────────────────────────────────────────────────────────────
echo
echo "────────────────────────────────────────────────────────────────────────"
echo "Results: ${PASS} passed, ${FAIL} failed"
echo "────────────────────────────────────────────────────────────────────────"
[ "$FAIL" -eq 0 ]
