# デバッグツール

このディレクトリには、トラックパッドイベントのデバッグとテストに使用するツールが含まれています。

## ツール一覧

### trackpad_cgevent_simple.c
CGEventTapを使用してマウス/トラックパッドイベントを監視するシンプルなツール。

コンパイル:
```bash
clang trackpad_cgevent_simple.c -framework ApplicationServices -framework CoreFoundation -o trackpad_cgevent_simple
```

### test_direct_scroll.c
基本的なスクロールイベント生成のテスト。

コンパイル:
```bash
clang test_direct_scroll.c -framework ApplicationServices -framework CoreFoundation -o test_direct_scroll
```

### test_fixed_scroll.c
修正版のスクロールイベント生成テスト。フェーズを含む完全なトラックパッドシミュレーション。

コンパイル:
```bash
clang test_fixed_scroll.c -framework ApplicationServices -framework CoreFoundation -o test_fixed_scroll
```

### compare_events.c
実際のトラックパッドイベントとシミュレーションイベントを比較するツール。

コンパイル:
```bash
clang compare_events.c -framework ApplicationServices -framework CoreFoundation -o compare_events
```

### test_scroll_simulation.go
featuresパッケージを使用したスクロールシミュレーションテスト。

実行:
```bash
go run test_scroll_simulation.go
```

### test_actual_scroll.go
実際のスクロール動作を確認するための統合テスト。

実行:
```bash
go run test_actual_scroll.go
```

## 使用方法

1. **イベント監視**: `trackpad_cgevent_simple`を実行して、実際のトラックパッドイベントを観察
2. **スクロールテスト**: `test_fixed_scroll`を実行して、スクロールが正しく動作するか確認
3. **比較テスト**: `compare_events`で実際のイベントとシミュレーションを比較

## 注意事項

- macOSのアクセシビリティ権限が必要です
- システム設定 > プライバシーとセキュリティ > アクセシビリティ でアプリケーションを許可してください