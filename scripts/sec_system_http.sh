#!/usr/bin/env bash
# 安全用例：未登录/坏 token/批量上限/深分页/限流等。
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BASE="${BASE_URL:-http://127.0.0.1:8888/api/v1}"
PRIV="${PWD_PRIV:-c351ebb6ad36d77acffaa929e770d040033d67d454b3aacc543fc957524e5925}"
USER="${SMOKE_USER:-superadmin}"
PASS="${SMOKE_PASS:-Admin@123456}"

enc() { go run "$ROOT/scripts/pwdenc/main.go" "$PRIV" "$1"; }
code_of() { python3 -c "import sys,json; print(json.load(sys.stdin).get('code',-999))"; }

expect_code() {
  local name="$1" want="$2" body="$3"
  local got
  got="$(printf '%s' "$body" | code_of)"
  if [[ "$got" != "$want" ]]; then
    echo "FAIL $name want=$want got=$got body=$body" >&2
    exit 1
  fi
  echo "OK   $name ($got)"
}

echo "== unauth / bad token =="
expect_code no_token 401 "$(curl -sS "$BASE/org/page?page=1&page_size=10")"
expect_code bad_token 401 "$(curl -sS -H 'Authorization: deadbeef' "$BASE/org/page?page=1&page_size=10")"

echo "== login for authed negatives =="
CAP="$(curl -sS "$BASE/auth/captcha")"
CID="$(printf '%s' "$CAP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['captcha_id'])")"
CCODE="$(docker exec redis redis-cli -a 'infra123!' --no-auth-warning GET "captcha:$CID" 2>/dev/null | tr -d '\r')"
CPWD="$(enc "$PASS")"
LOGIN="$(curl -sS -X POST "$BASE/auth/login" -H 'Content-Type: application/json' \
  -d "{\"username\":\"$USER\",\"password\":\"$CPWD\",\"captcha_id\":\"$CID\",\"captcha_code\":\"$CCODE\"}")"
TOKEN="$(printf '%s' "$LOGIN" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")"
AUTH=(-H "Authorization: $TOKEN" -H 'Content-Type: application/json')

echo "== page / batch limits =="
expect_code page_size 400 "$(curl -sS "${AUTH[@]}" "$BASE/org/page?page=1&page_size=9999")"
expect_code page_deep 400 "$(curl -sS "${AUTH[@]}" "$BASE/org/page?page=10001&page_size=10")"
IDS="$(python3 -c "print(','.join(['\"x%d\"'%i for i in range(101)]))")"
expect_code too_many_ids 400 "$(curl -sS -X POST "${AUTH[@]}" "$BASE/org/delete" -d "{\"ids\":[$IDS]}")"

echo "== dict by_code requires auth =="
expect_code dict_by_code_unauth 401 "$(curl -sS "$BASE/dict/by_code?dict_code=sys_user_status")"

echo "== rate limit login (best-effort) =="
# 使用错误验证码快速打满配额，避免真实锁定账号
HIT=0
for i in $(seq 1 25); do
  BODY="$(curl -sS -X POST "$BASE/auth/login" -H 'Content-Type: application/json' \
    -d "{\"username\":\"nobody\",\"password\":\"aa\",\"captcha_id\":\"x\",\"captcha_code\":\"y\"}")"
  C="$(printf '%s' "$BODY" | code_of)"
  if [[ "$C" == "429" ]]; then HIT=1; break; fi
done
if [[ "$HIT" != "1" ]]; then
  echo "FAIL rate_limit_login never got 429" >&2
  exit 1
fi
echo "OK   rate_limit_login (429)"

echo "ALL SEC PASSED"
