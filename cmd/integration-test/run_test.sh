#!/bin/bash

# 統合テスト実行スクリプト

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=== Keyball Gestures 統合テスト実行スクリプト ==="
echo

# カラー出力の定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# アクセシビリティ権限の確認
echo "アクセシビリティ権限を確認しています..."
if ! osascript -e 'tell application "System Events" to return (name of every process whose visible is true) contains "Finder"' &>/dev/null; then
    echo -e "${YELLOW}警告: アクセシビリティ権限が必要です${NC}"
    echo "システム設定 > プライバシーとセキュリティ > アクセシビリティ で権限を付与してください"
    echo
fi

# 作業ディレクトリに移動
cd "$SCRIPT_DIR"

# イベント監視プログラムをビルド
echo "イベント監視プログラムをビルドしています..."
if clang c/event_monitor.c -framework ApplicationServices -framework CoreFoundation -o event_monitor; then
    echo -e "${GREEN}✓ イベント監視プログラムのビルドが完了しました${NC}"
else
    echo -e "${RED}✗ イベント監視プログラムのビルドに失敗しました${NC}"
    exit 1
fi

# Go依存関係の確認
echo
echo "Go依存関係を確認しています..."
cd "$PROJECT_ROOT"
if go mod download; then
    echo -e "${GREEN}✓ 依存関係の確認が完了しました${NC}"
else
    echo -e "${RED}✗ 依存関係の確認に失敗しました${NC}"
    exit 1
fi

# 統合テストプログラムをビルド
echo
echo "統合テストプログラムをビルドしています..."
cd "$SCRIPT_DIR"
if go build -o integration_test .; then
    echo -e "${GREEN}✓ 統合テストプログラムのビルドが完了しました${NC}"
else
    echo -e "${RED}✗ 統合テストプログラムのビルドに失敗しました${NC}"
    exit 1
fi

# テスト結果ディレクトリを作成
mkdir -p test-results

# 統合テストを実行
echo
echo "統合テストを実行しています..."
echo "注意: このテストは実際のトラックパッドジェスチャーをシミュレートします"
echo

if sudo ./integration_test; then
    echo
    echo -e "${GREEN}✓ すべてのテストが成功しました！${NC}"
    echo
    echo "テスト結果は test-results/ ディレクトリに保存されています"
    exit 0
else
    echo
    echo -e "${RED}✗ いくつかのテストが失敗しました${NC}"
    echo
    echo "詳細なログは test-results/ ディレクトリを確認してください"
    exit 1
fi