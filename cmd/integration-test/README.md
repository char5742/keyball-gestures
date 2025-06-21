# Keyball Gestures 統合テスト

このディレクトリには、`trackpad_cgevent_simple.c`を活用した統合テストが含まれています。実際のトラックパッドジェスチャーをシミュレートし、生成されたイベントが正しく認識されることを検証します。

## 概要

統合テストは以下のコンポーネントで構成されています：

1. **event_monitor.c** - CGEventTapを使用してシステムイベントを監視し、JSON形式で記録
2. **main.go** - ジェスチャーをシミュレートし、記録されたイベントを検証
3. **run_test.sh** - ビルドと実行を自動化するスクリプト

## テストケース

### 1. 2本指スクロールテスト
- 2本指のタッチダウン
- 上方向へのスクロール動作
- ScrollWheelイベントの検証

### 2. 4本指スワイプテスト
- 4本指のタッチダウン
- 左方向へのスワイプ動作
- GestureBegin/Change/Endイベントの検証

## 実行方法

```bash
# 統合テストを実行
./run_test.sh
```

## 前提条件

1. **アクセシビリティ権限**
   - システム設定 > プライバシーとセキュリティ > アクセシビリティでターミナルに権限を付与

2. **開発環境**
   - macOS（CGEventTapを使用）
   - Xcode Command Line Tools（clangコンパイラ）
   - Go 1.16以上

## 出力

- テスト結果は`test-results/`ディレクトリに保存されます
- 各テストケースごとにJSONファイルが生成されます
- ファイル名形式: `{TestName}_{timestamp}.json`

## トラブルシューティング

### アクセシビリティ権限エラー
```
エラー: イベントタップの作成に失敗しました
```
→ システム設定でターミナルにアクセシビリティ権限を付与してください

### ビルドエラー
```
エラー: イベント監視プログラムのビルドに失敗しました
```
→ Xcode Command Line Toolsがインストールされているか確認してください：
```bash
xcode-select --install
```

## 開発者向け情報

### イベント監視プログラムの単体実行
```bash
# ビルド
clang event_monitor.c -framework ApplicationServices -framework CoreFoundation -o event_monitor

# 実行（10秒間監視）
./event_monitor output.json 10
```

### 統合テストプログラムの単体実行
```bash
# ビルド
go build -o integration_test .

# 実行（要sudo）
sudo ./integration_test
```

## ログファイルの解析

生成されたJSONファイルには以下の情報が含まれます：

```json
{
  "startTime": 1234567890,
  "duration": 10,
  "events": [
    {
      "timestamp": 123.456,
      "type": 22,
      "typeName": "ScrollWheel",
      "x": 500.0,
      "y": 400.0,
      "deltaX": 0,
      "deltaY": -10,
      "isGesture": true,
      "fingerCount": 2
    }
  ],
  "endTime": 1234567900,
  "totalEvents": 42
}
```