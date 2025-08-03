# 障害者サービス管理システム - Makefile
# 
# 主要な開発タスクを自動化するためのMakefile
# 使用方法: make <target>

.PHONY: help build test test-verbose test-coverage clean lint fmt security-check install-tools run dev

# デフォルトターゲット
.DEFAULT_GOAL := help

# 変数定義
APP_NAME := shien-system
BUILD_DIR := ./bin
MAIN_PATH := ./cmd/desktop
COVERAGE_FILE := coverage.out
GOLANGCI_LINT_VERSION := v1.55.2

# Go関連の設定
GOPATH := $(shell go env GOPATH)
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# カラー出力用
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
RESET := \033[0m

## help: このヘルプメッセージを表示
help:
	@echo "$(BLUE)障害者サービス管理システム - 開発コマンド$(RESET)"
	@echo ""
	@echo "$(GREEN)利用可能なコマンド:$(RESET)"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(YELLOW)%-15s$(RESET) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""

## build: アプリケーションをビルド
build: clean
	@echo "$(BLUE)アプリケーションをビルド中...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "$(GREEN)ビルド完了: $(BUILD_DIR)/$(APP_NAME)$(RESET)"

## build-all: 全プラットフォーム向けにビルド
build-all: clean
	@echo "$(BLUE)全プラットフォーム向けビルド中...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	
	# Windows 64bit
	@echo "$(YELLOW)Windows 64bit版をビルド中...$(RESET)"
	@GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(MAIN_PATH)
	
	# macOS 64bit (Intel)
	@echo "$(YELLOW)macOS 64bit (Intel)版をビルド中...$(RESET)"
	@GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(MAIN_PATH)
	
	# macOS ARM64 (Apple Silicon)
	@echo "$(YELLOW)macOS ARM64 (Apple Silicon)版をビルド中...$(RESET)"
	@GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(MAIN_PATH)
	
	# Linux 64bit
	@echo "$(YELLOW)Linux 64bit版をビルド中...$(RESET)"
	@GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(MAIN_PATH)
	
	@echo "$(GREEN)全プラットフォーム向けビルド完了$(RESET)"
	@ls -la $(BUILD_DIR)/

## test: テストを実行
test:
	@echo "$(BLUE)テストを実行中...$(RESET)"
	@go test -race ./...
	@echo "$(GREEN)テスト完了$(RESET)"

## test-verbose: 詳細出力でテストを実行
test-verbose:
	@echo "$(BLUE)詳細テストを実行中...$(RESET)"
	@go test -race -v ./...

## test-coverage: カバレッジ付きでテストを実行
test-coverage:
	@echo "$(BLUE)カバレッジ測定中...$(RESET)"
	@go test -race -coverprofile=$(COVERAGE_FILE) ./...
	@go tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "$(GREEN)カバレッジレポート: coverage.html$(RESET)"
	@go tool cover -func=$(COVERAGE_FILE) | grep total

## test-security: セキュリティ関連のテストを実行
test-security:
	@echo "$(BLUE)セキュリティテストを実行中...$(RESET)"
	@go test -v ./internal/validation/...
	@go test -v ./internal/adapter/crypto/...
	@echo "$(GREEN)セキュリティテスト完了$(RESET)"

## benchmark: ベンチマークテストを実行
benchmark:
	@echo "$(BLUE)ベンチマークテストを実行中...$(RESET)"
	@go test -bench=. -benchmem ./...

## lint: コードの静的解析を実行
lint: install-tools
	@echo "$(BLUE)静的解析を実行中...$(RESET)"
	@golangci-lint run ./...
	@echo "$(GREEN)静的解析完了$(RESET)"

## fmt: コードフォーマットを実行
fmt:
	@echo "$(BLUE)コードフォーマット中...$(RESET)"
	@go fmt ./...
	@gofumpt -w .
	@echo "$(GREEN)フォーマット完了$(RESET)"

## security-check: セキュリティチェックを実行
security-check:
	@echo "$(BLUE)セキュリティチェックを実行中...$(RESET)"
	@command -v gosec >/dev/null 2>&1 || go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@gosec -fmt=json -out=security-report.json ./...
	@gosec ./...
	@echo "$(GREEN)セキュリティチェック完了$(RESET)"

## vulnerability-check: 脆弱性チェックを実行
vulnerability-check:
	@echo "$(BLUE)脆弱性チェックを実行中...$(RESET)"
	@command -v govulncheck >/dev/null 2>&1 || go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...
	@echo "$(GREEN)脆弱性チェック完了$(RESET)"

## install-tools: 開発ツールをインストール
install-tools:
	@echo "$(BLUE)開発ツールをインストール中...$(RESET)"
	@command -v golangci-lint >/dev/null 2>&1 || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin $(GOLANGCI_LINT_VERSION)
	@command -v gofumpt >/dev/null 2>&1 || go install mvdan.cc/gofumpt@latest
	@command -v gotests >/dev/null 2>&1 || go install github.com/cweill/gotests/gotests@latest
	@command -v air >/dev/null 2>&1 || go install github.com/cosmtrek/air@latest
	@echo "$(GREEN)開発ツールのインストール完了$(RESET)"

## run: アプリケーションを実行
run:
	@echo "$(BLUE)アプリケーションを起動中...$(RESET)"
	@go run $(MAIN_PATH)

## dev: 開発モード（ホットリロード）で起動
dev: install-tools
	@echo "$(BLUE)開発モードで起動中...$(RESET)"
	@air

## clean: ビルド成果物をクリーンアップ
clean:
	@echo "$(BLUE)クリーンアップ中...$(RESET)"
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE)
	@rm -f coverage.html
	@rm -f security-report.json
	@rm -f *.sqlite *.db *.sqlite3
	@go clean -cache
	@echo "$(GREEN)クリーンアップ完了$(RESET)"

## deps: 依存関係を更新
deps:
	@echo "$(BLUE)依存関係を更新中...$(RESET)"
	@go mod download
	@go mod tidy
	@go mod verify
	@echo "$(GREEN)依存関係の更新完了$(RESET)"

## generate: コード生成を実行
generate:
	@echo "$(BLUE)コード生成中...$(RESET)"
	@go generate ./...
	@echo "$(GREEN)コード生成完了$(RESET)"

## docker-build: Dockerイメージをビルド
docker-build:
	@echo "$(BLUE)Dockerイメージをビルド中...$(RESET)"
	@docker build -t $(APP_NAME):latest .
	@echo "$(GREEN)Dockerイメージビルド完了$(RESET)"

## package: リリース用パッケージを作成
package: build-all
	@echo "$(BLUE)リリースパッケージを作成中...$(RESET)"
	@mkdir -p releases
	
	# Windows版のパッケージ
	@zip -j releases/$(APP_NAME)-windows-amd64.zip $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe README.md LICENSE
	
	# macOS版のパッケージ
	@tar -czf releases/$(APP_NAME)-darwin-amd64.tar.gz -C $(BUILD_DIR) $(APP_NAME)-darwin-amd64 -C .. README.md LICENSE
	@tar -czf releases/$(APP_NAME)-darwin-arm64.tar.gz -C $(BUILD_DIR) $(APP_NAME)-darwin-arm64 -C .. README.md LICENSE
	
	# Linux版のパッケージ
	@tar -czf releases/$(APP_NAME)-linux-amd64.tar.gz -C $(BUILD_DIR) $(APP_NAME)-linux-amd64 -C .. README.md LICENSE
	
	@echo "$(GREEN)リリースパッケージ作成完了:$(RESET)"
	@ls -la releases/

## verify: リリース前の検証を実行
verify: clean fmt lint test security-check vulnerability-check
	@echo "$(GREEN)全ての検証が完了しました$(RESET)"

## db-reset: 開発用データベースをリセット
db-reset:
	@echo "$(YELLOW)開発用データベースをリセット中...$(RESET)"
	@rm -f dev.sqlite test.sqlite *.db
	@echo "$(GREEN)データベースリセット完了$(RESET)"

## migrate: データベースマイグレーションを実行
migrate:
	@echo "$(BLUE)データベースマイグレーションを実行中...$(RESET)"
	@go run $(MAIN_PATH) --migrate
	@echo "$(GREEN)マイグレーション完了$(RESET)"

## backup-dev: 開発データベースのバックアップを作成
backup-dev:
	@echo "$(BLUE)開発データベースのバックアップを作成中...$(RESET)"
	@mkdir -p backups
	@cp dev.sqlite backups/dev-backup-$(shell date +%Y%m%d-%H%M%S).sqlite
	@echo "$(GREEN)バックアップ完了$(RESET)"

# リリース関連のターゲット
## release-patch: パッチバージョンをリリース
release-patch:
	@echo "$(BLUE)パッチリリースを準備中...$(RESET)"
	@./scripts/release.sh patch

## release-minor: マイナーバージョンをリリース
release-minor:
	@echo "$(BLUE)マイナーリリースを準備中...$(RESET)"
	@./scripts/release.sh minor

## release-major: メジャーバージョンをリリース
release-major:
	@echo "$(RED)メジャーリリースを準備中...$(RESET)"
	@./scripts/release.sh major

# 環境変数の例
## env-example: 環境変数の例を表示
env-example:
	@echo "$(BLUE)環境変数の設定例:$(RESET)"
	@echo "export SHIEN_DB_PATH=\"./dev.sqlite\""
	@echo "export SHIEN_LOG_LEVEL=\"debug\""
	@echo "export SHIEN_ENCRYPTION_KEY=\"your-32-byte-encryption-key\""
	@echo "export SHIEN_BACKUP_DIR=\"./backups\""
	@echo ""
	@echo "$(YELLOW)本番環境では適切なセキュリティ設定を行ってください$(RESET)"

# プロジェクト情報
## info: プロジェクト情報を表示
info:
	@echo "$(BLUE)プロジェクト情報:$(RESET)"
	@echo "  名前: 障害者サービス管理システム"
	@echo "  言語: Go $(shell go version | cut -d' ' -f3)"
	@echo "  OS/Arch: $(GOOS)/$(GOARCH)"
	@echo "  モジュール: $(shell head -1 go.mod | cut -d' ' -f2)"
	@echo ""
	@echo "$(GREEN)詳細情報は README.md を参照してください$(RESET)"