#!/usr/bin/env bash
# demoservice 文件上传统一冒烟（local / s3 同一套前端流程）
# 流程：init → part_url → PUT upload_url（读 ETag）→ complete(parts) → download
set -euo pipefail

BASE="${BASE:-http://127.0.0.1:8889}"
KEY="${KEY:-demo/smoke-$(date +%s).bin}"
WORKDIR="$(mktemp -d)"
trap 'rm -rf "$WORKDIR"' EXIT

need_cmd() { command -v "$1" >/dev/null || { echo "missing $1"; exit 1; }; }
need_cmd curl
need_cmd python3
need_cmd sha256sum
need_cmd jq
need_cmd split

urlenc() { python3 -c 'import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1], safe=""))' "$1"; }
sha256_file() { sha256sum "$1" | awk '{print $1}'; }
abs_url() {
  local u="$1"
  if [[ "$u" == http://* || "$u" == https://* ]]; then
    echo "$u"
  else
    echo "${BASE}${u}"
  fi
}

echo "==> workdir=$WORKDIR key=$KEY base=$BASE"

# 1. init（先拿 chunk_size，再按 provider 生成合适大小的测试文件）
INIT=$(curl -fsS -X POST "$BASE/api/v1/file/upload/init" \
  -H 'Content-Type: application/json' \
  -d "{\"key\":\"$KEY\",\"content_type\":\"application/octet-stream\"}")
echo "$INIT" | jq .
[[ "$(echo "$INIT" | jq -r '.code')" == "0" ]] || exit 1
UPLOAD_ID=$(echo "$INIT" | jq -r '.data.upload_id')
CHUNK=$(echo "$INIT" | jq -r '.data.chunk_size')
PROVIDER=$(echo "$INIT" | jq -r '.data.provider')
UPLOAD_VIA=$(echo "$INIT" | jq -r '.data.upload_via')
echo "==> provider=$PROVIDER upload_via=$UPLOAD_VIA upload_id=$UPLOAD_ID chunk=$CHUNK"
[[ "$UPLOAD_VIA" == "part_url" ]] || { echo "expect upload_via=part_url"; exit 1; }

# 2. 构造略大于 1 个 chunk 的文件（保证至少 2 片）
python3 - "$CHUNK" >"$WORKDIR/payload.bin" <<'PY'
import os, sys
chunk = int(sys.argv[1])
os.write(1, b"A" * chunk + b"B" * max(1, chunk // 4))
PY
SIZE=$(wc -c <"$WORKDIR/payload.bin" | tr -d ' ')
echo "==> payload size=$SIZE"
KEY_ENC=$(urlenc "$KEY")

# 3. 统一：part_url → PUT → 收集 etag/sha256
split -b "$CHUNK" -d -a 4 "$WORKDIR/payload.bin" "$WORKDIR/part_"
PARTS_JSON='[]'
PART_NO=1
for f in "$WORKDIR"/part_*; do
  [[ -f "$f" ]] || continue
  SZ=$(wc -c <"$f" | tr -d ' ')
  HASH=$(sha256_file "$f")
  echo "==> part_url part=$PART_NO size=$SZ"
  URL_RESP=$(curl -fsS -X POST "$BASE/api/v1/file/upload/part_url" \
    -H 'Content-Type: application/json' \
    -d "{\"key\":\"$KEY\",\"upload_id\":\"$UPLOAD_ID\",\"part_number\":\"$PART_NO\",\"size\":\"$SZ\",\"chunk_sha256\":\"$HASH\"}")
  echo "$URL_RESP" | jq 'del(.data.upload_url)'
  [[ "$(echo "$URL_RESP" | jq -r '.code')" == "0" ]] || exit 1
  UPLOAD_URL=$(abs_url "$(echo "$URL_RESP" | jq -r '.data.upload_url')")

  # 组装 part_url 要求的额外头
  HDR_ARGS=()
  while IFS= read -r line; do
    [[ -n "$line" ]] || continue
    HDR_ARGS+=(-H "$line")
  done < <(echo "$URL_RESP" | jq -r '.data.headers // {} | to_entries[] | "\(.key): \(.value)"')

  echo "==> PUT part=$PART_NO url=${UPLOAD_URL%%\?*}"
  HTTP=$(curl -sS -D "$WORKDIR/h$PART_NO.txt" -o "$WORKDIR/b$PART_NO.json" -w '%{http_code}' -X PUT "$UPLOAD_URL" \
    "${HDR_ARGS[@]}" \
    -H "Content-Length: $SZ" \
    --data-binary @"$f")
  echo "http=$HTTP"
  [[ "$HTTP" == "200" ]] || { cat "$WORKDIR/h$PART_NO.txt"; cat "$WORKDIR/b$PART_NO.json" 2>/dev/null || true; exit 1; }

  ETAG=$(awk -F': ' 'tolower($1)=="etag"{gsub(/\r/,"",$2); gsub(/"/,"",$2); print $2; exit}' "$WORKDIR/h$PART_NO.txt")
  # local JSON 体也可能带 sha256；优先头，其次 body
  if [[ -z "$ETAG" && -f "$WORKDIR/b$PART_NO.json" ]]; then
    ETAG=$(jq -r '.data.sha256 // empty' "$WORKDIR/b$PART_NO.json" 2>/dev/null || true)
  fi
  [[ -n "$ETAG" ]] || { echo "missing etag"; exit 1; }
  echo "==> etag=$ETAG"

  PARTS_JSON=$(jq -c --arg n "$PART_NO" --arg s "$SZ" --arg h "$HASH" --arg e "$ETAG" \
    '. + [{part_number:$n,size:$s,sha256:$h,etag:$e}]' <<<"$PARTS_JSON")
  PART_NO=$((PART_NO + 1))
done

# 4. list + complete（两边都传 parts）
PARTS=$(curl -fsS "$BASE/api/v1/file/upload/parts?key=${KEY_ENC}&upload_id=${UPLOAD_ID}")
echo "$PARTS" | jq .
[[ "$(echo "$PARTS" | jq -r '.code')" == "0" ]] || exit 1

DONE=$(curl -fsS -X POST "$BASE/api/v1/file/upload/complete" \
  -H 'Content-Type: application/json' \
  -d "$(jq -n --arg k "$KEY" --arg u "$UPLOAD_ID" --argjson p "$PARTS_JSON" \
    '{key:$k,upload_id:$u,parts:$p}')")
echo "$DONE" | jq .
[[ "$(echo "$DONE" | jq -r '.code')" == "0" ]] || exit 1

# 5. head + 统一经应用下载
HEAD=$(curl -fsS "$BASE/api/v1/file/head?key=${KEY_ENC}")
echo "$HEAD" | jq .
[[ "$(echo "$HEAD" | jq -r '.data.size')" == "$SIZE" ]] || { echo "size mismatch"; exit 1; }

curl -fsS "$BASE/api/v1/file/download?key=${KEY_ENC}" -o "$WORKDIR/downloaded.bin"
ORIG_HASH=$(sha256_file "$WORKDIR/payload.bin")
DL_HASH=$(sha256_file "$WORKDIR/downloaded.bin")
echo "==> origin=$ORIG_HASH downloaded=$DL_HASH"
[[ "$ORIG_HASH" == "$DL_HASH" ]] || exit 1

# 6. abort
INIT2=$(curl -fsS -X POST "$BASE/api/v1/file/upload/init" \
  -H 'Content-Type: application/json' \
  -d "{\"key\":\"${KEY}-abort\"}")
UID2=$(echo "$INIT2" | jq -r '.data.upload_id')
ABORT=$(curl -fsS -X POST "$BASE/api/v1/file/upload/abort" \
  -H 'Content-Type: application/json' \
  -d "{\"key\":\"${KEY}-abort\",\"upload_id\":\"$UID2\"}")
echo "$ABORT" | jq .
[[ "$(echo "$ABORT" | jq -r '.code')" == "0" ]] || exit 1

echo "OK: unified file smoke passed (provider=$PROVIDER)"
