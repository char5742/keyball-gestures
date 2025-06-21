# Touchpad Manual Testing

このディレクトリには、touchpad_darwin.goの修正版が実際にtrackpad_dump.cで検出できるかを確認する手動テストが含まれています。

## 修正内容

touchpad_darwin.goで以下の修正を行いました：

1. **スワイプイベント送信レベルの変更**:
   - `kCGAnnotatedSessionEventTap` → `kCGSessionEventTap`
   - これにより、trackpad_dump.cで4本指スワイプイベントが検出可能になりました

2. **スクロールイベント送信レベルの統一**:
   - `kCGHIDEventTap` → `kCGSessionEventTap`
   - 一貫性のため、すべてのイベントを同じレベルで送信

## テスト手順

### 1. 事前準備

アクセシビリティ権限が設定されていることを確認してください：
- システム環境設定 > セキュリティとプライバシー > プライバシー > アクセシビリティ
- ターミナルアプリケーションを追加

### 2. trackpad_dumpを起動

最初のターミナルで以下を実行：

```bash
cd test/debugging
./trackpad_dump
```

以下のような出力が表示されます：
```
=== CGEvent dump v2 ===
Listening…  Ctrl+C to quit
```

### 3. 手動テストを実行

2番目のターミナルで以下を実行：

```bash
./test/manual/manual_touchpad
```

または、直接実行：

```bash
go run test/manual/manual_touchpad.go
```

### 4. 期待される結果

手動テストを実行すると、trackpad_dumpのターミナルで以下のようなログが表示されるはずです：

#### 4本指スワイプ
```
[time.xxx] Gesture type=29(Begin) sub=0(?) phase=? mask=0 flags=0x0
[time.xxx] Gesture type=29(Begin) sub=23(Swipe) phase=Began mask=0x2 [Right ↦ ] flags=0x0
          Δ(0.0,0.0)
[time.xxx] Gesture type=30(Change) sub=23(Swipe) phase=Changed mask=0x2 [Right ↦ ] flags=0x0
          Δ(0.0,0.0)
[time.xxx] Gesture type=31(End) sub=23(Swipe) phase=Ended mask=0x2 [Right ↦ ] flags=0x0
          Δ(0.0,0.0)
```

#### 2本指スクロール
```
[time.xxx] Scroll MayBegin Δ(0.0,0.0)
[time.xxx] Scroll Began Δ(0.0,0.0)
[time.xxx] Scroll Changed Δ(x.x,y.y)
[time.xxx] Scroll Ended Δ(0.0,0.0)
```

### 5. 成功の判定基準

以下が確認できれば修正が成功しています：

✅ **ジェスチャーイベントが検出される**:
- `type=29(Begin) sub=23(Swipe)`
- `type=30(Change) sub=23(Swipe)`
- `type=31(End) sub=23(Swipe)`

✅ **スクロールイベントが検出される**:
- `Scroll Began`, `Scroll Changed`, `Scroll Ended`

✅ **方向マスクが正しく設定される**:
- 右スワイプ: `mask=0x2 [Right ↦ ]`

## トラブルシューティング

### イベントが検出されない場合

1. **アクセシビリティ権限**:
   - ターミナルアプリケーションにアクセシビリティ権限が設定されているか確認

2. **trackpad_dumpの権限**:
   ```bash
   # trackpad_dumpが権限を要求した場合
   Need Accessibility permission
   ```
   この場合、システム環境設定で権限を設定し直してください

3. **ビルドの確認**:
   ```bash
   go build -o keyball-gestures .
   # 修正版のコードがビルドされていることを確認
   ```

### 部分的にしか検出されない場合

- **スクロールのみ検出**: 修正前の状態。touchpad_darwin.goの再ビルドが必要
- **ジェスチャーのみ検出**: 正常。スクロールイベントは実装によって検出されない場合があります

## 修正前後の比較

### 修正前 (問題あり)
```c
// スワイプイベント
CGEventPost(kCGAnnotatedSessionEventTap, gesture);  // trackpad_dumpで検出されない

// スクロールイベント  
CGEventPost(kCGHIDEventTap, e22);  // 検出される可能性はあるが一貫性がない
```

### 修正後 (修正版)
```c
// スワイプイベント
CGEventPost(kCGSessionEventTap, gesture);  // trackpad_dumpで検出される

// スクロールイベント
CGEventPost(kCGSessionEventTap, e22);  // 一貫性を保って同じレベルで送信
```

## 関連ファイル

- `internal/features/touchpad_darwin.go` - 修正されたメイン実装
- `test/debugging/trackpad_dump.c` - 比較用の実際のトラックパッドログ取得
- `test/manual/manual_touchpad.go` - この手動テストプログラム

## 備考

この修正により、touchpad_darwin.goで生成されるイベントが、実際のトラックパッドと同じレベル（kCGSessionEventTap）で送信されるようになり、trackpad_dump.cによる監視が可能になりました。