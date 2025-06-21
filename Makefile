# Keyball Gestures Makefile

# 変数定義
BINARY_NAME = keyball-gestures
GO = go
GOFLAGS = -v
MAIN_PATH = cmd/main.go
INSTALL_DIR = /usr/local/bin

# OS検出
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    TARGET_OS = linux
endif
ifeq ($(UNAME_S),Darwin)
    TARGET_OS = darwin
endif

# デフォルトターゲット
.PHONY: all
all: build

# ビルド
.PHONY: build
build:
	@echo "Building for $(TARGET_OS)..."
	GOOS=$(TARGET_OS) $(GO) build $(GOFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

# Linux用ビルド
.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	GOOS=linux $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-linux $(MAIN_PATH)

# macOS用ビルド
.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	GOOS=darwin $(GO) build $(GOFLAGS) -o $(BINARY_NAME)-darwin $(MAIN_PATH)

# クロスプラットフォームビルド
.PHONY: build-all
build-all: build-linux build-darwin

# テスト
.PHONY: test
test:
	$(GO) test -v ./...

# 統合テスト
.PHONY: integration-test
integration-test:
	$(GO) test -tags=integration -v ./test/integration/...

# 短い統合テスト（CI用）
.PHONY: integration-test-short
integration-test-short:
	$(GO) test -tags=integration -short -v ./test/integration/...

# イベント検証付き統合テスト（macOS専用）
.PHONY: integration-test-validation
integration-test-validation:
	$(GO) test -tags=integration -v -run "Validation" ./test/integration/...

# カバレッジ付きテスト
.PHONY: test-coverage
test-coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# ベンチマーク
.PHONY: benchmark
benchmark:
	$(GO) test -tags=integration -bench=. ./test/integration/

# フォーマット
.PHONY: fmt
fmt:
	$(GO) fmt ./...

# 静的解析
.PHONY: vet
vet:
	$(GO) vet ./...

# lint (golangci-lintが必要)
.PHONY: lint
lint:
	golangci-lint run

# インストール
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
ifeq ($(TARGET_OS),darwin)
	@sudo cp $(BINARY_NAME) $(INSTALL_DIR)/
	@sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
else
	@sudo cp $(BINARY_NAME) $(INSTALL_DIR)/
	@sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
endif

# アンインストール
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
ifeq ($(TARGET_OS),darwin)
	@launchctl unload ~/Library/LaunchAgents/com.keyball.gestures.plist 2>/dev/null || true
	@rm -f ~/Library/LaunchAgents/com.keyball.gestures.plist
endif

# クリーン
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME) $(BINARY_NAME)-linux $(BINARY_NAME)-darwin
	@rm -f coverage.out coverage.html
	@rm -rf cmd/integration-test/test-results/*
	@rm -f cmd/integration-test/event_monitor
	@rm -f cmd/integration-test/integration_test

# 実行（開発用）
.PHONY: run
run: build
	./$(BINARY_NAME)

# API モードで実行（開発用）
.PHONY: run-api
run-api: build
	./$(BINARY_NAME) --api

# 依存関係の更新
.PHONY: deps
deps:
	$(GO) get -u ./...
	$(GO) mod tidy

# 依存関係のダウンロード
.PHONY: mod-download
mod-download:
	$(GO) mod download

# サンプル実行
.PHONY: run-demo
run-demo:
	$(GO) run ./cmd/demo/main.go

.PHONY: run-monitor
run-monitor:
	$(GO) run ./examples/monitoring/basic_monitor.go

.PHONY: run-gesture-demo
run-gesture-demo:
	$(GO) run ./examples/gestures/gesture_demo.go

# integration-testツールの実行
.PHONY: run-integration-tool
run-integration-tool:
	cd cmd/integration-test && ./run_test.sh

# ヘルプ
.PHONY: help
help:
	@echo "利用可能なターゲット:"
	@echo "  make build              - 現在のOS用にビルド"
	@echo "  make build-linux        - Linux用にビルド"
	@echo "  make build-darwin       - macOS用にビルド"
	@echo "  make build-all          - 全プラットフォーム用にビルド"
	@echo "  make test               - ユニットテストを実行"
	@echo "  make integration-test   - 統合テストを実行"
	@echo "  make test-coverage      - カバレッジレポートを生成"
	@echo "  make benchmark          - ベンチマークを実行"
	@echo "  make fmt                - コードをフォーマット"
	@echo "  make vet                - 静的解析を実行"
	@echo "  make lint               - lintを実行"
	@echo "  make install            - バイナリをインストール"
	@echo "  make uninstall          - バイナリをアンインストール"
	@echo "  make clean              - ビルド成果物を削除"
	@echo "  make run                - ビルドして実行"
	@echo "  make run-api            - APIモードでビルドして実行"
	@echo "  make run-demo           - デモを実行"
	@echo "  make run-monitor        - 入力モニターを実行"
	@echo "  make run-gesture-demo   - ジェスチャーデモを実行"
	@echo "  make deps               - 依存関係を更新"
	@echo "  make help               - このヘルプを表示"