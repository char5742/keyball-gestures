# テストディレクトリの整理方針

## 現在のテスト構成

### 1. 基本的な入力テスト
- **test-simple**: 最小限の機能テスト
- **test-input**: Keyball39からの入力監視
- **test-macos**: featuresパッケージの統合テスト

### 2. HIDレベルのテスト
- **test-hid**: 全HIDデバイスの監視
- **test-keyball-hid**: Keyball39専用HID監視
- **test-keyball-specific**: Keyball39の詳細イベント分析

### 3. タッチパッド/ジェスチャーテスト
- **test-scroll**: スクロール機能の専用テスト
- **test-gestures**: マルチタッチジェスチャーの総合テスト

### 4. システムレベルテスト
- **test-keys**: CGEventTapを使用したシステム全体のキー監視
- **test-iokit**: IOKit APIの動的ロードテスト

### 5. 統合テスト
- **integration-test**: イベント記録と自動検証を含む統合テスト

### 6. その他
- **demo**: デモンストレーション用コード

## 推奨される整理方針

### 1. カテゴリ別にディレクトリを再構成

```
cmd/
├── examples/           # サンプルコード
│   ├── basic/         # 基本的な使用例
│   ├── gestures/      # ジェスチャーのデモ
│   └── hid/           # HIDレベルの例
├── tests/             # テストコード
│   ├── unit/          # ユニットテスト
│   ├── integration/   # 統合テスト
│   └── benchmark/     # パフォーマンステスト
└── tools/             # 開発/デバッグツール
    ├── monitor/       # イベント監視ツール
    └── simulator/     # イベントシミュレーター
```

### 2. 重複の削減
- 同じ機能をテストしているものは統合
- 開発段階で作成された一時的なテストは削除または統合

### 3. ドキュメントの充実
- 各テストディレクトリにREADMEを追加
- テストの目的と使用方法を明確に記載