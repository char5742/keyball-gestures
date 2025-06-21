# Touchpad Integration Tests

このディレクトリには、`internal/features/touchpad_darwin.go`の統合テストが含まれています。これらのテストは、タッチパッド機能が実際にmacOSのCoreGraphicsイベントシステムと正しく連携することを検証します。

## テストの概要

### `touchpad_darwin_test.go`

このファイルには以下の統合テストが含まれています：

1. **`TestTouchpadScrollIntegration`** - 2本指スクロールの統合テスト
2. **`TestTouchpadSwipeIntegration`** - 4本指スワイプの統合テスト  
3. **`TestTouchpadEventSequenceValidation`** - イベントシーケンスの検証テスト

## テストの実行方法

### 前提条件

1. **macOS環境**: これらのテストはmacOS専用です
2. **アクセシビリティ権限**: テストはCGEventTapを使用するため、アクセシビリティ権限が必要です
3. **CGO有効**: テストはCコードを含むため、cgoが有効である必要があります

### 権限の設定

テスト実行前に、以下の手順でアクセシビリティ権限を設定してください：

1. システム環境設定 > セキュリティとプライバシー > プライバシー
2. 「アクセシビリティ」を選択
3. テストを実行するターミナルアプリケーションを追加

### テストの実行

```bash
# 統合テストのみ実行
go test -tags="integration,darwin,cgo" ./test/integration/

# 特定のテストのみ実行
go test -tags="integration,darwin,cgo" -run TestTouchpadScrollIntegration ./test/integration/

# 詳細なログ出力
go test -tags="integration,darwin,cgo" -v ./test/integration/

# 短縮モードを無効化（全テストを実行）
go test -tags="integration,darwin,cgo" -v -short=false ./test/integration/
```

## テストの動作原理

### DetailedEventMonitor

テストでは`DetailedEventMonitor`という拡張されたイベント監視機能を使用します：

- **CGEventTap**: macOSのCGEventTapを使用してシステムイベントを監視
- **詳細なフィールド取得**: サブタイプ、フェーズ、マスク、フラグなど、ジェスチャー固有の情報を取得
- **リアルタイム監視**: タッチパッドが生成するイベントをリアルタイムで記録

### 検証項目

#### 2本指スクロール
- スクロールイベント（type=22）の生成
- ジェスチャー開始イベント（type=29）の生成
- 適切なサブタイプ（6 = Scroll）の設定
- フェーズの正しい遷移（MayBegin → Began → Changed → Ended）

#### 4本指スワイプ
- ジェスチャーイベント（type=29, 30, 31）の生成
- 適切なサブタイプ（23 = Swipe）の設定
- 方向マスクの正しい設定
- フェーズの正しい遷移

### 実際のトラックパッドとの比較

テストで生成されるイベントは、`test/debugging/trackpad_dump.c`で記録される実際のトラックパッドイベントと同じ形式になるように設計されています：

```
[31665.505] Gesture type=30(Change) sub=23(Swipe) phase=Began mask=0x1 [Left ↤ ] flags=0x0
          Δ(0.0,0.0)
[31665.518] Gesture type=29(Begin) sub=0(?) phase=? mask=0 flags=0x0
```

## トラブルシューティング

### 一般的な問題

1. **権限エラー**: アクセシビリティ権限が設定されていない
   - 解決方法: 上記の権限設定手順を実行

2. **CGO関連エラー**: cgoが有効でない、またはフレームワークがリンクされていない
   - 解決方法: CGO_ENABLED=1を設定し、必要なフレームワークが利用可能であることを確認

3. **イベントが記録されない**: イベントモニターが正しく動作していない
   - 解決方法: システムの負荷を下げ、十分な待機時間を設定

### デバッグ

テストが失敗した場合、以下の情報を確認してください：

1. **詳細ログ**: `monitor.PrintDebugInfo()`の出力を確認
2. **イベント数**: 期待されるイベント数と実際のイベント数を比較
3. **フェーズ遷移**: イベントのフェーズが正しく遷移しているか確認

## 制限事項

1. **実行環境**: macOS 10.14以降を推奨
2. **システム負荷**: 高負荷時はイベントの取得が不安定になる可能性
3. **権限**: 管理者権限またはアクセシビリティ権限が必要
4. **リソース**: イベントモニターはシステムリソースを消費

## 継続的インテグレーション

CI環境での実行時は以下を考慮してください：

1. **権限**: CI環境でのアクセシビリティ権限の設定
2. **ヘッドレス**: GUI環境なしでの実行可能性
3. **タイムアウト**: ネットワーク遅延やシステム負荷を考慮した適切なタイムアウト設定

## 関連ファイル

- `internal/features/touchpad_darwin.go` - テスト対象のメイン実装
- `test/debugging/trackpad_dump.c` - 実際のトラックパッドイベントの参考実装
- `test/integration/event_monitor.go` - 基本的なイベント監視機能
- `test/integration/features_test.go` - 基本的な機能テスト