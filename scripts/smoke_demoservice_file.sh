#!/usr/bin/env bash
# demoservice 本地分片上传 + 公开直链完整冒烟
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

urlenc() { python3 -c 'import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1], safe=""))' "$1"; }

echo "==> workdir=$WORKDIR key=$KEY base=$BASE"

# 构造 2500 字节测试文件 → 按 1024 分片（demo yaml ChunkSizeBytes）
python3 - <<'PY' >"$WORKDIR/payload.bin"
import os
os.write(1, b"A" * 1024 + b"B" * 1024 + b"C" * 452)
PY
SIZE=$(wc -c <"$WORKDIR/payload.bin" | tr -d ' ')
echo "==> payload size=$SIZE"

sha256_file() { sha256sum "$1" | awk '{print $1}'; }
KEY_ENC=$(urlenc "$KEY")

# 1. init
INIT=$(curl -fsS -X POST "$BASE/api/v1/file/upload/init" \
  -H 'Content-Type: application/json' \
  -d "{\"key\":\"$KEY\",\"content_type\":\"application/octet-stream\"}")
echo "$INIT" | jq .
CODE=$(echo "$INIT" | jq -r '.code')
[[ "$CODE" == "0" ]] || { echo "init failed"; exit 1; }
UPLOAD_ID=$(echo "$INIT" | jq -r '.data.upload_id')
CHUNK=$(echo "$INIT" | jq -r '.data.chunk_size')
echo "==> upload_id=$UPLOAD_ID chunk_size=$CHUNK"

# 2. 切分并上传各 part
split -b "$CHUNK" -d -a 4 "$WORKDIR/payload.bin" "$WORKDIR/part_"
PART_NO=1
for f in "$WORKDIR"/part_*; do
  [[ -f "$f" ]] || continue
  SZ=$(wc -c <"$f" | tr -d ' ')
  HASH=$(sha256_file "$f")
  echo "==> put part=$PART_NO size=$SZ sha=$HASH"
  RESP=$(curl -fsS -X PUT \
    "$BASE/api/v1/file/upload/part?key=${KEY_ENC}&upload_id=${UPLOAD_ID}&part_number=${PART_NO}" \
    -H "Content-Type: application/octet-stream" \
    -H "X-Chunk-SHA256: $HASH" \
    -H "Content-Length: $SZ" \
    --data-binary @"$f")
  echo "$RESP" | jq .
  [[ "$(echo "$RESP" | jq -r '.code')" == "0" ]] || { echo "put part failed"; exit 1; }
  PART_NO=$((PART_NO + 1))
done

# 3. 幂等重传第 1 片
FIRST=$(ls "$WORKDIR"/part_* | head -n1)
SZ=$(wc -c <"$FIRST" | tr -d ' ')
HASH=$(sha256_file "$FIRST")
echo "==> idempotent put part=1"
RESP=$(curl -fsS -X PUT \
  "$BASE/api/v1/file/upload/part?key=${KEY_ENC}&upload_id=${UPLOAD_ID}&part_number=1" \
  -H "Content-Type: application/octet-stream" \
  -H "X-Chunk-SHA256: $HASH" \
  -H "Content-Length: $SZ" \
  --data-binary @"$FIRST")
[[ "$(echo "$RESP" | jq -r '.code')" == "0" ]] || { echo "idempotent put failed"; exit 1; }

# 4. list parts
PARTS=$(curl -fsS "$BASE/api/v1/file/upload/parts?key=${KEY_ENC}&upload_id=${UPLOAD_ID}")
echo "$PARTS" | jq .
[[ "$(echo "$PARTS" | jq -r '.code')" == "0" ]] || { echo "list parts failed"; exit 1; }
COUNT=$(echo "$PARTS" | jq '.data.items | length')
echo "==> listed parts=$COUNT"

# 5. 坏哈希应失败
echo "==> expect bad checksum fail"
BAD=$(curl -sS -X PUT \
  "$BASE/api/v1/file/upload/part?key=${KEY_ENC}&upload_id=${UPLOAD_ID}&part_number=99" \
  -H "Content-Type: application/octet-stream" \
  -H "X-Chunk-SHA256: $(printf 'a%.0s' {1..64})" \
  -H "Content-Length: 3" \
  --data-binary "xyz" || true)
echo "$BAD" | jq . || echo "$BAD"
[[ "$(echo "$BAD" | jq -r '.code // empty')" != "0" ]] || { echo "bad checksum should fail"; exit 1; }

# 6. complete
DONE=$(curl -fsS -X POST "$BASE/api/v1/file/upload/complete" \
  -H 'Content-Type: application/json' \
  -d "{\"key\":\"$KEY\",\"upload_id\":\"$UPLOAD_ID\"}")
echo "$DONE" | jq .
[[ "$(echo "$DONE" | jq -r '.code')" == "0" ]] || { echo "complete failed"; exit 1; }
PUBLIC_URL=$(echo "$DONE" | jq -r '.data.public_url')
MERKLE=$(echo "$DONE" | jq -r '.data.merkle_sha256')
echo "==> public_url=$PUBLIC_URL merkle=$MERKLE"
[[ -n "$PUBLIC_URL" && "$PUBLIC_URL" != "null" ]] || { echo "missing public_url"; exit 1; }

# 7. head
HEAD=$(curl -fsS "$BASE/api/v1/file/head?key=${KEY_ENC}")
echo "$HEAD" | jq .
[[ "$(echo "$HEAD" | jq -r '.data.size')" == "$SIZE" ]] || { echo "head size mismatch"; exit 1; }

# 8. 公开直链下载并比对
curl -fsS "$PUBLIC_URL" -o "$WORKDIR/downloaded.bin"
ORIG_HASH=$(sha256_file "$WORKDIR/payload.bin")
DL_HASH=$(sha256_file "$WORKDIR/downloaded.bin")
echo "==> origin=$ORIG_HASH downloaded=$DL_HASH"
[[ "$ORIG_HASH" == "$DL_HASH" ]] || { echo "download content mismatch"; exit 1; }

# 9. abort 幂等（对新会话）
INIT2=$(curl -fsS -X POST "$BASE/api/v1/file/upload/init" \
  -H 'Content-Type: application/json' \
  -d "{\"key\":\"${KEY}-abort\"}")
UID2=$(echo "$INIT2" | jq -r '.data.upload_id')
ABORT=$(curl -fsS -X POST "$BASE/api/v1/file/upload/abort" \
  -H 'Content-Type: application/json' \
  -d "{\"key\":\"${KEY}-abort\",\"upload_id\":\"$UID2\"}")
echo "$ABORT" | jq . || true
[[ "$(echo "$ABORT" | jq -r '.code // 0')" == "0" ]] || { echo "abort failed"; exit 1; }

echo "OK: demoservice file smoke passed"
