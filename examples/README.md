# サンプルコード集

このディレクトリには、Keyball Gesturesの使用方法を示すサンプルコードが含まれています。

## ディレクトリ構成

### monitoring/
基本的な入力監視のサンプル
- `basic_monitor.go`: キーボード、マウス、タッチパッドの統合監視

### gestures/
ジェスチャー機能のデモンストレーション
- `gesture_demo.go`: マルチタッチジェスチャーのデモ

### hid-debug/
HIDレベルのデバッグツール
- `hid-monitor/`: 全HIDデバイスの監視
- `keyball-monitor/`: Keyball39専用の詳細モニター

## 実行方法

各サンプルは独立した実行可能プログラムです：

```bash
# 基本的な監視
go run examples/monitoring/basic_monitor.go

# ジェスチャーデモ
go run examples/gestures/gesture_demo.go

# HIDデバッグ - 全デバイス監視
go run examples/hid-debug/hid-monitor/main.go

# HIDデバッグ - Keyball39専用
go run examples/hid-debug/keyball-monitor/main.go
```

## 必要な権限

多くのサンプルはmacOSのアクセシビリティ権限が必要です：
- システム設定 > プライバシーとセキュリティ > アクセシビリティ

一部のサンプル（タッチパッド関連）はroot権限が必要です。

## 注意事項

これらのサンプルは教育目的のコードです。本番環境での使用時は適切なエラーハンドリングとリソース管理を行ってください。