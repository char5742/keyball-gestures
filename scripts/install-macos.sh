#!/bin/bash

# Keyball Gestures macOS インストールスクリプト

set -e

echo "Keyball Gestures macOSインストールスクリプト"
echo "============================================"

# 管理者権限の確認（アクセシビリティ権限のため）
if ! command -v sudo &> /dev/null; then
    echo "エラー: sudoコマンドが見つかりません"
    exit 1
fi

# Go言語のインストール確認
if ! command -v go &> /dev/null; then
    echo "エラー: Go言語がインストールされていません"
    echo "Homebrewでインストール: brew install go"
    exit 1
fi

# ビルド
echo ""
echo "アプリケーションをビルドしています..."
cd "$(dirname "$0")/.."
go build -o keyball-gestures cmd/main.go

# バイナリをインストール
echo ""
echo "バイナリをインストールしています..."
INSTALL_DIR="/usr/local/bin"
sudo mkdir -p "$INSTALL_DIR"
sudo cp keyball-gestures "$INSTALL_DIR/"
sudo chmod +x "$INSTALL_DIR/keyball-gestures"

# 設定ディレクトリの作成
CONFIG_DIR="$HOME/.config/keyball-gestures"
echo ""
echo "設定ディレクトリを作成しています: $CONFIG_DIR"
mkdir -p "$CONFIG_DIR"

# サンプル設定ファイルのコピー
if [ ! -f "$CONFIG_DIR/config.toml" ]; then
    echo "サンプル設定ファイルをコピーしています..."
    cp example-config-macos.toml "$CONFIG_DIR/config.toml"
    echo "設定ファイル: $CONFIG_DIR/config.toml"
fi

# LaunchAgentの作成（オプション）
echo ""
read -p "起動時に自動実行するLaunchAgentを作成しますか? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    LAUNCH_AGENT_DIR="$HOME/Library/LaunchAgents"
    LAUNCH_AGENT_FILE="$LAUNCH_AGENT_DIR/com.keyball.gestures.plist"
    
    mkdir -p "$LAUNCH_AGENT_DIR"
    
    cat > "$LAUNCH_AGENT_FILE" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.keyball.gestures</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/keyball-gestures</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardErrorPath</key>
    <string>/tmp/keyball-gestures.err</string>
    <key>StandardOutPath</key>
    <string>/tmp/keyball-gestures.out</string>
</dict>
</plist>
EOF
    
    echo "LaunchAgentを作成しました: $LAUNCH_AGENT_FILE"
    echo "LaunchAgentをロードしています..."
    launchctl load "$LAUNCH_AGENT_FILE"
fi

# アクセシビリティ権限の確認
echo ""
echo "重要: アクセシビリティ権限の設定"
echo "================================"
echo "keyball-gesturesがキーボードとマウスイベントを監視するには、"
echo "システム環境設定でアクセシビリティ権限を付与する必要があります。"
echo ""
echo "1. システム環境設定 > セキュリティとプライバシー > プライバシー を開く"
echo "2. 左側のリストから「アクセシビリティ」を選択"
echo "3. 右側のリストに keyball-gestures が表示されたらチェックを入れる"
echo ""
echo "初回実行時に権限ダイアログが表示される場合があります。"
echo ""

echo "インストールが完了しました！"
echo ""
echo "使用方法:"
echo "  keyball-gestures              # CLIモードで起動"
echo "  keyball-gestures --api        # APIモード（Web UI付き）で起動"
echo ""
echo "設定ファイル: $CONFIG_DIR/config.toml"
echo ""

# ビルド成果物のクリーンアップ
rm -f keyball-gestures