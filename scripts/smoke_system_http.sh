#!/usr/bin/env bash
# 全接口功能冒烟：依赖已启动的 system API、MySQL、Redis。
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BASE="${BASE_URL:-http://127.0.0.1:8888/api/v1}"
PRIV="${PWD_PRIV:-c351ebb6ad36d77acffaa929e770d040033d67d454b3aacc543fc957524e5925}"
USER="${SMOKE_USER:-superadmin}"
PASS="${SMOKE_PASS:-Admin@123456}"
ORG_ID="${SMOKE_ORG_ID:-01a20d49-091b-7f18-b8dd-6018e8e0b384}"

enc() { go run "$ROOT/scripts/pwdenc/main.go" "$PRIV" "$1"; }

json_get() { python3 -c "import sys,json; d=json.load(sys.stdin); print(d$1)"; }

assert_ok() {
  local name="$1" body="$2"
  local code
  code="$(printf '%s' "$body" | python3 -c "import sys,json; print(json.load(sys.stdin).get('code',-999))")"
  if [[ "$code" != "0" ]]; then
    echo "FAIL $name code=$code body=$body" >&2
    exit 1
  fi
  echo "OK   $name"
}

echo "== captcha / public_key / config.public =="
CAP="$(curl -sS "$BASE/auth/captcha")"
assert_ok captcha "$CAP"
CID="$(printf '%s' "$CAP" | json_get "['data']['captcha_id']")"
# 验证码明文在 redis，冒烟登录用管理库改/或从 redis 取；此处用 redis-cli
CCODE="$(docker exec redis redis-cli -a 'infra123!' --no-auth-warning GET "captcha:$CID" 2>/dev/null | tr -d '\r')"
if [[ -z "$CCODE" ]]; then
  echo "FAIL captcha code missing in redis for $CID" >&2
  exit 1
fi
PK="$(curl -sS "$BASE/auth/public_key")"
assert_ok public_key "$PK"
PUB="$(curl -sS "$BASE/config/public")"
assert_ok config_public "$PUB"

echo "== login =="
CPWD="$(enc "$PASS")"
LOGIN="$(curl -sS -X POST "$BASE/auth/login" -H 'Content-Type: application/json' \
  -d "{\"username\":\"$USER\",\"password\":\"$CPWD\",\"captcha_id\":\"$CID\",\"captcha_code\":\"$CCODE\"}")"
assert_ok login "$LOGIN"
TOKEN="$(printf '%s' "$LOGIN" | json_get "['data']['token']")"
AUTH=(-H "Authorization: $TOKEN" -H 'Content-Type: application/json')

echo "== dict =="
assert_ok dict_by_code "$(curl -sS "${AUTH[@]}" "$BASE/dict/by_code?dict_code=sys_user_status")"
TPAGE="$(curl -sS "${AUTH[@]}" "$BASE/dict/type/page?page=1&page_size=20")"
assert_ok dict_type_page "$TPAGE"
TYPE_ID="$(printf '%s' "$TPAGE" | python3 -c "import sys,json; items=json.load(sys.stdin)['data']['items'];
print(next(i['id'] for i in items if i.get('dict_code')=='sys_user_status'))")"
assert_ok dict_type_detail "$(curl -sS "${AUTH[@]}" "$BASE/dict/type/detail?id=$TYPE_ID")"
assert_ok dict_item_page "$(curl -sS "${AUTH[@]}" "$BASE/dict/item/page?page=1&page_size=20&parent_id=$TYPE_ID")"
NEW_CODE="smoke_dict_$(date +%s)"
DCREATE="$(curl -sS -X POST "${AUTH[@]}" "$BASE/dict/type/create" -d "{\"dict_code\":\"$NEW_CODE\",\"label\":\"冒烟字典\",\"sort\":\"0\",\"status\":\"1\"}")"
assert_ok dict_type_create "$DCREATE"
DID="$(printf '%s' "$DCREATE" | json_get "['data']['id']")"
ICREATE="$(curl -sS -X POST "${AUTH[@]}" "$BASE/dict/item/create" -d "{\"parent_id\":\"$DID\",\"label\":\"项A\",\"value\":\"a\",\"sort\":\"1\",\"status\":\"1\"}")"
assert_ok dict_item_create "$ICREATE"
IID="$(printf '%s' "$ICREATE" | json_get "['data']['id']")"
assert_ok dict_item_detail "$(curl -sS "${AUTH[@]}" "$BASE/dict/item/detail?id=$IID")"
assert_ok dict_item_update "$(curl -sS -X POST "${AUTH[@]}" "$BASE/dict/item/update" -d "{\"id\":\"$IID\",\"label\":\"项A改\"}")"
assert_ok dict_type_update "$(curl -sS -X POST "${AUTH[@]}" "$BASE/dict/type/update" -d "{\"id\":\"$DID\",\"label\":\"冒烟字典改\"}")"
assert_ok dict_item_delete "$(curl -sS -X POST "${AUTH[@]}" "$BASE/dict/item/delete" -d "{\"ids\":[\"$IID\"]}")"
assert_ok dict_type_delete "$(curl -sS -X POST "${AUTH[@]}" "$BASE/dict/type/delete" -d "{\"ids\":[\"$DID\"]}")"

echo "== org =="
assert_ok org_page "$(curl -sS "${AUTH[@]}" "$BASE/org/page?page=1&page_size=20")"
assert_ok org_detail "$(curl -sS "${AUTH[@]}" "$BASE/org/detail?id=$ORG_ID")"
OCREATE="$(curl -sS -X POST "${AUTH[@]}" "$BASE/org/create" -d "{\"parent_id\":\"$ORG_ID\",\"name\":\"冒烟组织\",\"code\":\"smoke_org_$(date +%s)\",\"sort\":\"9\",\"status\":\"1\"}")"
assert_ok org_create "$OCREATE"
OID="$(printf '%s' "$OCREATE" | json_get "['data']['id']")"
assert_ok org_update "$(curl -sS -X POST "${AUTH[@]}" "$BASE/org/update" -d "{\"id\":\"$OID\",\"name\":\"冒烟组织改\"}")"

echo "== user =="
assert_ok user_page "$(curl -sS "${AUTH[@]}" "$BASE/user/page?page=1&page_size=20")"
UPWD="$(enc 'Smoke@123456')"
UCREATE="$(curl -sS -X POST "${AUTH[@]}" "$BASE/user/create" -d "{\"org_id\":\"$ORG_ID\",\"username\":\"smoke_u_$(date +%s)\",\"password\":\"$UPWD\",\"nickname\":\"冒烟\",\"status\":\"1\"}")"
assert_ok user_create "$UCREATE"
USER_ID="$(printf '%s' "$UCREATE" | json_get "['data']['id']")"
assert_ok user_detail "$(curl -sS "${AUTH[@]}" "$BASE/user/detail?id=$USER_ID")"
assert_ok user_update "$(curl -sS -X POST "${AUTH[@]}" "$BASE/user/update" -d "{\"id\":\"$USER_ID\",\"nickname\":\"冒烟改\"}")"
ROLE_ID="01a20d49-091b-7f2c-ad6c-4d9d2222b2e7"
assert_ok user_bind "$(curl -sS -X POST "${AUTH[@]}" "$BASE/user/bind_roles" -d "{\"user_id\":\"$USER_ID\",\"role_ids\":[\"$ROLE_ID\"]}")"
assert_ok user_unbind "$(curl -sS -X POST "${AUTH[@]}" "$BASE/user/unbind_roles" -d "{\"user_id\":\"$USER_ID\",\"role_ids\":[\"$ROLE_ID\"]}")"
ME="$(printf '%s' "$LOGIN" | json_get "['data']['user']['id']")"
assert_ok user_resources "$(curl -sS "${AUTH[@]}" "$BASE/user/resources?user_id=$ME")"

echo "== role =="
assert_ok role_page "$(curl -sS "${AUTH[@]}" "$BASE/role/page?page=1&page_size=20")"
RCREATE="$(curl -sS -X POST "${AUTH[@]}" "$BASE/role/create" -d "{\"name\":\"冒烟角色\",\"code\":\"smoke_role_$(date +%s)\",\"data_scope\":\"1\",\"status\":\"1\"}")"
assert_ok role_create "$RCREATE"
RID="$(printf '%s' "$RCREATE" | json_get "['data']['id']")"
assert_ok role_detail "$(curl -sS "${AUTH[@]}" "$BASE/role/detail?id=$RID")"
assert_ok role_update "$(curl -sS -X POST "${AUTH[@]}" "$BASE/role/update" -d "{\"id\":\"$RID\",\"name\":\"冒烟角色改\",\"data_scope\":\"1\"}")"
# 取一个按钮资源绑定
RES_ID="$(curl -sS "${AUTH[@]}" "$BASE/resource/page?page=1&page_size=5" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['items'][0]['id'])")"
assert_ok role_bind "$(curl -sS -X POST "${AUTH[@]}" "$BASE/role/bind_resources" -d "{\"role_id\":\"$RID\",\"resource_ids\":[\"$RES_ID\"]}")"
assert_ok role_unbind "$(curl -sS -X POST "${AUTH[@]}" "$BASE/role/unbind_resources" -d "{\"role_id\":\"$RID\",\"resource_ids\":[\"$RES_ID\"]}")"

echo "== resource =="
assert_ok resource_page "$(curl -sS "${AUTH[@]}" "$BASE/resource/page?page=1&page_size=20")"
assert_ok resource_detail "$(curl -sS "${AUTH[@]}" "$BASE/resource/detail?id=$RES_ID")"
RSC="$(curl -sS -X POST "${AUTH[@]}" "$BASE/resource/create" -d "{\"parent_id\":\"\",\"name\":\"冒烟目录\",\"type\":\"directory\",\"path\":\"/smoke_$(date +%s)\",\"sort\":\"99\",\"status\":\"1\"}")"
assert_ok resource_create "$RSC"
RSID="$(printf '%s' "$RSC" | json_get "['data']['id']")"
assert_ok resource_update "$(curl -sS -X POST "${AUTH[@]}" "$BASE/resource/update" -d "{\"id\":\"$RSID\",\"name\":\"冒烟目录改\"}")"

echo "== config / audit =="
assert_ok config_page "$(curl -sS "${AUTH[@]}" "$BASE/config/page?page=1&page_size=20")"
CFG_ROW="$(curl -sS "${AUTH[@]}" "$BASE/config/detail?config_key=captcha.length")"
assert_ok config_detail "$CFG_ROW"
CFG_VAL="$(printf '%s' "$CFG_ROW" | json_get "['data']['config_value']")"
assert_ok config_update "$(curl -sS -X POST "${AUTH[@]}" "$BASE/config/update" -d "{\"items\":[{\"config_key\":\"captcha.length\",\"config_value\":\"$CFG_VAL\"}]}")"
assert_ok audit_login "$(curl -sS "${AUTH[@]}" "$BASE/audit/login_page?page=1&page_size=10")"
assert_ok audit_oper "$(curl -sS "${AUTH[@]}" "$BASE/audit/oper_page?page=1&page_size=10")"

echo "== cleanup =="
assert_ok user_delete "$(curl -sS -X POST "${AUTH[@]}" "$BASE/user/delete" -d "{\"ids\":[\"$USER_ID\"]}")"
assert_ok role_delete "$(curl -sS -X POST "${AUTH[@]}" "$BASE/role/delete" -d "{\"ids\":[\"$RID\"]}")"
assert_ok resource_delete "$(curl -sS -X POST "${AUTH[@]}" "$BASE/resource/delete" -d "{\"ids\":[\"$RSID\"]}")"
assert_ok org_delete "$(curl -sS -X POST "${AUTH[@]}" "$BASE/org/delete" -d "{\"ids\":[\"$OID\"]}")"

echo "== change_password + logout =="
OLD_ENC="$(enc "$PASS")"
NEW_ENC="$(enc 'Admin@123456x')"
assert_ok change_password "$(curl -sS -X POST "${AUTH[@]}" "$BASE/auth/change_password" -d "{\"old_password\":\"$OLD_ENC\",\"new_password\":\"$NEW_ENC\"}")"
# 改密会踢会话；用 SQL 恢复种子密码，避免污染本机环境
docker exec mysql mysql -uroot -p'infra123!' -e \
  "USE light_gozero; UPDATE sys_user SET must_change_password=0, password_hash='\$2a\$10\$h82iSyCZiOwJV9EbIq5awOd3zdB6WGMR8MHfgtzQoBTIpxymvz6zK' WHERE username='${USER}'; DELETE FROM sys_password_history WHERE user_id IN (SELECT id FROM (SELECT id FROM sys_user WHERE username='${USER}') t);" \
  >/dev/null 2>&1 || true
echo "ALL SMOKE PASSED"
