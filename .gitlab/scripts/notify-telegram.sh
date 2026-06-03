#!/usr/bin/env bash
# Usage: notify-telegram.sh <status> <environment> <provider> <image>
# Env vars: TELEGRAM_BOT_TOKEN, TELEGRAM_CHAT_ID
#           CI_PROJECT_PATH, CI_COMMIT_REF_NAME, CI_COMMIT_SHORT_SHA, GITLAB_USER_LOGIN
set -euo pipefail

STATUS="${1:-unknown}"
ENV="${2:-unknown}"
PROVIDER="${3:-unknown}"
IMAGE="${4:-unknown}"

case "$STATUS" in
  success)  EMOJI="✅" ;;
  failure)  EMOJI="❌" ;;
  *)        EMOJI="⚠️" ;;
esac

RUN_URL="${CI_PROJECT_URL}/-/pipelines/${CI_PIPELINE_ID}"

MESSAGE="${EMOJI} *Deploy ${STATUS^^}*

*Repo:*        \`${CI_PROJECT_PATH}\`
*Branch/Tag:*  \`${CI_COMMIT_REF_NAME}\`
*Commit:*      \`${CI_COMMIT_SHORT_SHA}\`
*Author:*      ${GITLAB_USER_LOGIN}
*Environment:* \`${ENV}\`
*Provider:*    \`${PROVIDER}\`
*Image:*       \`${IMAGE}\`

[View pipeline](${RUN_URL})"

curl -fsSL \
  -X POST "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage" \
  -H "Content-Type: application/json" \
  --data-binary "$(jq -n \
    --arg chat_id "$TELEGRAM_CHAT_ID" \
    --arg text    "$MESSAGE" \
    '{chat_id: $chat_id, text: $text, parse_mode: "Markdown", disable_web_page_preview: true}'
  )" > /dev/null
