# Keyball Gestures コマンドディレクトリ

このディレクトリには、Keyball Gesturesプロジェクトのメインアプリケーションと関連ツールが含まれています。

## ディレクトリ構成

### メインアプリケーション
- **main.go**: Keyball Gesturesのメインアプリケーション

### デモ
- **demo/**: 基本的なデモンストレーション

### 統合テスト
- **integration-test/**: イベント記録と自動検証を含む統合テストツール
  - Cコンポーネントを使用したイベント監視
  - 自動化されたジェスチャーテスト

### 開発ツール（デバッグ用）
- **test-keyball-hid/**: Keyball39のHIDレベルデバッグ
- **test-keys/**: システム全体のキーイベント監視
- **test-iokit/**: IOKit APIの技術検証

## テストについて

Go標準の統合テストは `/test/integration/` ディレクトリにあります：
```bash
# 統合テストの実行
go test -tags=integration ./test/integration/...
```

サンプルコードとモニタリングツールは `/examples/` ディレクトリにあります。

## 必要な権限

多くのツールはmacOSのアクセシビリティ権限を必要とします：
- システム設定 > プライバシーとセキュリティ > アクセシビリティ

一部のツールはsudo権限が必要です（仮想HIDデバイスの作成のため）。