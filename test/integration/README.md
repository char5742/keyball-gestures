# 統合テスト

このディレクトリには、Keyball Gesturesの統合テストが含まれています。

## テストの実行

### 全ての統合テストを実行
```bash
go test -tags=integration ./test/integration/...
```

### 短いテストのみ実行（時間のかかるテストをスキップ）
```bash
go test -tags=integration -short ./test/integration/...
```

### 特定のテストを実行
```bash
go test -tags=integration -run TestKeyboardCreation ./test/integration/
```

### ベンチマークを実行
```bash
go test -tags=integration -bench=. ./test/integration/
```

### カバレッジレポートを生成
```bash
go test -tags=integration -coverprofile=coverage.out ./test/integration/...
go tool cover -html=coverage.out
```

## テストファイル

- `features_test.go`: featuresパッケージの基本的な統合テスト
  - デバイスの作成テスト
  - 入力検出テスト
  - タッチパッドジェスチャーテスト
  - 統合シナリオテスト

- `event_test.go`: イベント記録とベンチマーク
  - イベント記録機能のテスト
  - パフォーマンスベンチマーク

## 必要な権限

一部のテストは以下の権限が必要です：
- macOSのアクセシビリティ権限
- 仮想HIDデバイス作成のためのroot権限（一部のテスト）

## CI/CDでの使用

CIパイプラインでは以下のコマンドを使用することを推奨します：
```bash
# 短いテストのみ（権限不要）
go test -tags=integration -short ./test/integration/...

# フルテスト（権限必要）
sudo go test -tags=integration ./test/integration/...
```