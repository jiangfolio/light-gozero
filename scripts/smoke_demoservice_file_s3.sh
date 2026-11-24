#!/usr/bin/env bash
# 兼容入口：与 smoke_demoservice_file.sh 同一套无感流程（依赖 yaml Provider=s3）
exec "$(dirname "$0")/smoke_demoservice_file.sh" "$@"
