#!/usr/bin/env bash
set -euo pipefail
state=/tmp/competency-results-verify-state.json
db=$(jq -r .redisDb "$state")
redis_key=$(jq -r .redisKey "$state")
token_id=${redis_key#login_tokens:}
now_ms=$(date +%s%3N)
expire_ms=$((now_ms + 1800000))
payload=$(jq -nc \
  --arg token "$token_id" \
  --argjson now "$now_ms" \
  --argjson expire "$expire_ms" \
  '{userId:1,token:$token,loginTime:$now,expireTime:$expire,permissions:["*:*:*"],roles:["admin"],user:{userId:1,userName:"admin",avatar:""}}')
printf '%s' "$payload" | redis-cli -n "$db" -x SET "$redis_key" >/dev/null
redis-cli -n "$db" EXPIRE "$redis_key" 1800 >/dev/null
[ "$(redis-cli -n "$db" EXISTS "$redis_key")" = "1" ]
echo BROWSER_LOGIN_CONTEXT_READY
