#!/usr/bin/env bash
# demoservice + 本地 silo（S3 兼容）预签名分片直传冒烟
set -euo pipefail

BASE="${BASE:-http://127.0.0.1:8889}"
KEY="${KEY:-demo/s3-smoke-$(date +%s).bin}"
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

echo "==> workdir=$WORKDIR key=$KEY base=$BASE"

# 6MiB：1 个满 5MiB + 1MiB 尾片
python3 - <<'PY' >"$WORKDIR/payload.bin"
import os
os.write(1, b"S" * (5 * 1024 * 1024) + b"T" * (1 * 1024 * 1024))
PY
SIZE=$(wc -c <"$WORKDIR/payload.bin" | tr -d ' ')
echo "==> payload size=$SIZE"
KEY_ENC=$(urlenc "$KEY")

# 1. init
INIT=$(curl -fsS -X POST "$BASE/api/v1/file/upload/init" \
  -H 'Content-Type: application/json' \
  -d "{\"key\":\"$KEY\",\"content_type\":\"application/octet-stream\"}")
echo "$INIT" | jq .
[[ "$(echo "$INIT" | jq -r '.code')" == "0" ]] || exit 1
UPLOAD_ID=$(echo "$INIT" | jq -r '.data.upload_id')
CHUNK=$(echo "$INIT" | jq -r '.data.chunk_size')
PROVIDER=$(echo "$INIT" | jq -r '.data.provider')
echo "==> provider=$PROVIDER upload_id=$UPLOAD_ID chunk=$CHUNK"
[[ "$PROVIDER" == "s3" ]] || { echo "expect s3 provider"; exit 1; }

# 2. 切分并直传各 part
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
    -d "{\"key\":\"$KEY\",\"upload_id\":\"$UPLOAD_ID\",\"part_number\":\"$PART_NO\",\"size\":\"$SZ\"}")
  echo "$URL_RESP" | jq 'del(.data.upload_url)' 
  [[ "$(echo "$URL_RESP" | jq -r '.code')" == "0" ]] || exit 1
  UPLOAD_URL=$(echo "$URL_RESP" | jq -r '.data.upload_url')

  echo "==> PUT silo part=$PART_NO"
  HDRS=$(curl -sS -D "$WORKDIR/h$PART_NO.txt" -o /dev/null -w '%{http_code}' -X PUT "$UPLOAD_URL" \
    -H "Content-Length: $SZ" \
    --data-binary @"$f")
  echo "http=$HDRS"
  [[ "$HDRS" == "200" ]] || { cat "$WORKDIR/h$PART_NO.txt"; exit 1; }
  ETAG=$(awk -F': ' 'tolower($1)=="etag"{gsub(/\r/,"",$2); gsub(/"/,"",$2); print $2; exit}' "$WORKDIR/h$PART_NO.txt")
  echo "==> etag=$ETAG"
  [[ -n "$ETAG" ]] || { echo "missing etag"; exit 1; }

  PARTS_JSON=$(jq -c --arg n "$PART_NO" --arg s "$SZ" --arg h "$HASH" --arg e "$ETAG" \
    '. + [{part_number:$n,size:$s,sha256:$h,etag:$e}]' <<<"$PARTS_JSON")
  PART_NO=$((PART_NO + 1))
done

# 3. list parts（silo 侧）
PARTS=$(curl -fsS "$BASE/api/v1/file/upload/parts?key=${KEY_ENC}&upload_id=${UPLOAD_ID}")
echo "$PARTS" | jq .
[[ "$(echo "$PARTS" | jq -r '.code')" == "0" ]] || exit 1

# 4. complete
DONE=$(curl -fsS -X POST "$BASE/api/v1/file/upload/complete" \
  -H 'Content-Type: application/json' \
  -d "$(jq -n --arg k "$KEY" --arg u "$UPLOAD_ID" --argjson p "$PARTS_JSON" \
    '{key:$k,upload_id:$u,parts:$p}')")
echo "$DONE" | jq .
[[ "$(echo "$DONE" | jq -r '.code')" == "0" ]] || exit 1

# 5. head + 经应用下载比对
HEAD=$(curl -fsS "$BASE/api/v1/file/head?key=${KEY_ENC}")
echo "$HEAD" | jq .
[[ "$(echo "$HEAD" | jq -r '.data.size')" == "$SIZE" ]] || { echo "size mismatch"; exit 1; }

curl -fsS "$BASE/api/v1/file/download?key=${KEY_ENC}" -o "$WORKDIR/downloaded.bin"
ORIG_HASH=$(sha256_file "$WORKDIR/payload.bin")
DL_HASH=$(sha256_file "$WORKDIR/downloaded.bin")
echo "==> origin=$ORIG_HASH downloaded=$DL_HASH"
[[ "$ORIG_HASH" == "$DL_HASH" ]] || exit 1

# 6. abort 新会话
INIT2=$(curl -fsS -X POST "$BASE/api/v1/file/upload/init" \
  -H 'Content-Type: application/json' \
  -d "{\"key\":\"${KEY}-abort\"}")
UID2=$(echo "$INIT2" | jq -r '.data.upload_id')
ABORT=$(curl -fsS -X POST "$BASE/api/v1/file/upload/abort" \
  -H 'Content-Type: application/json' \
  -d "{\"key\":\"${KEY}-abort\",\"upload_id\":\"$UID2\"}")
echo "$ABORT" | jq .
[[ "$(echo "$ABORT" | jq -r '.code')" == "0" ]] || exit 1

echo "OK: demoservice S3/silo smoke passed"
