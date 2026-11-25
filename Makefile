# light-gozero 常用命令
ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
GOPROXY ?= https://goproxy.cn,direct
export GOPROXY

SYSTEM_DIR := $(ROOT)/service/system/api
DEMO_DIR   := $(ROOT)/service/demoservice/api
GOCTL      ?= $(shell go env GOPATH)/bin/goctl

WEB_DIR := $(ROOT)/web/admin

.PHONY: help tidy build test test-storage \
	run-system run-demo run-demo-local \
	smoke-system smoke-system-sec smoke-demo smoke-demo-s3 \
	init-sql init-submodule gen-system-api gen-demo-api \
	web-install web-dev web-build

help: ## 显示可用目标
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} \
		/^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

tidy: ## go mod tidy
	cd $(ROOT) && go mod tidy

build: ## 编译 system / demoservice / common
	cd $(ROOT) && go build ./common/... ./service/system/... ./service/demoservice/...

test: ## 跑全仓 go test
	cd $(ROOT) && go test ./...

test-storage: ## 只跑 common/storage 单测
	cd $(ROOT) && go test ./common/storage/ -count=1

run-system: ## 启动 system（:8888）
	cd $(SYSTEM_DIR) && go run . -f etc/system.yaml

run-demo: ## 启动 demoservice（当前 etc/demoservice.yaml，默认 s3）
	cd $(DEMO_DIR) && go run . -f etc/demoservice.yaml

run-demo-local: ## 启动 demoservice（local 存储）
	cd $(DEMO_DIR) && go run . -f etc/demoservice-local.yaml

smoke-system: ## system HTTP 冒烟（需已启动 system + MySQL/Redis）
	bash $(ROOT)/scripts/smoke_system_http.sh

smoke-system-sec: ## system 安全用例
	bash $(ROOT)/scripts/sec_system_http.sh

smoke-demo: ## demoservice 文件上传统一冒烟（随 yaml Provider）
	bash $(ROOT)/scripts/smoke_demoservice_file.sh

smoke-demo-s3: ## 同 smoke-demo（兼容入口）
	bash $(ROOT)/scripts/smoke_demoservice_file_s3.sh

init-sql: ## 导入 UUIDv7 初始化 SQL（默认 docker mysql 容器名 mysql；务必 utf8mb4）
	docker exec -i mysql mysql -uroot -p'infra123!' --default-character-set=utf8mb4 < $(ROOT)/deploy/sql/light_gozero_init_uuidv7.sql

init-submodule: ## 初始化 .cursorrules submodule
	cd $(ROOT) && git submodule update --init --recursive

gen-system-api: ## goctl 重新生成 system API（需本机 goctl）
	cd $(SYSTEM_DIR) && $(GOCTL) api go -api system.api -dir . --style go_zero

gen-demo-api: ## goctl 重新生成 demoservice API
	cd $(DEMO_DIR) && $(GOCTL) api go -api demoservice.api -dir . --style go_zero

web-install: ## 安装 web/admin 依赖（pnpm）
	cd $(WEB_DIR) && pnpm install

web-dev: ## 启动管理端前端（:8081；/api/system→:8888，/api/demo→:8889）
	cd $(WEB_DIR) && pnpm run dev

web-build: ## 构建管理端前端
	cd $(WEB_DIR) && pnpm run build
