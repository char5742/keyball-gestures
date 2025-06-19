# macOS TouchSimulator実装仕様書

## 概要

本文書は、Mac Mouse Fixプロジェクトで実装されているタッチジェスチャーシミュレーション機能の詳細な技術仕様をまとめたものです。この実装は、マウス入力をトラックパッドのようなジェスチャー入力に変換することで、macOSのネイティブなジェスチャー機能を利用可能にします。

## 1. 基本アーキテクチャ

### 1.1 使用API
- **Core Graphics Event API** (公開API)
  - `CGEventCreate()`
  - `CGEventSetIntegerValueField()` / `CGEventSetDoubleValueField()`
  - `CGEventPost()`
  - `CGEventSetTimestamp()`

### 1.2 重要な非公開フィールド
TouchSimulatorは、CGEventの未文書化フィールドを使用してイベントを構築します。これらのフィールドは、将来のmacOSバージョンで変更される可能性があることに注意してください。

## 2. CGEventフィールドマッピング

### 2.1 イベントタイプフィールド

| フィールド番号 | 説明 | 設定値 |
|--------------|------|--------|
| 55 | NSEventType | 22 (ScrollWheel), 29 (Gesture), 30 (Magnify) |
| 110 | IOHIDEventType サブタイプ | 5-23の範囲（詳細は後述） |

### 2.2 IOHIDEventTypeサブタイプ

| 値 | 定数名 | 用途 |
|----|--------|------|
| 5 | kIOHIDEventTypeRotation | 回転ジェスチャー |
| 6 | kIOHIDEventTypeScroll | スクロールジェスチャー |
| 8 | kIOHIDEventTypeZoom | ズームジェスチャー |
| 16 | kIOHIDEventTypeNavigationSwipe | ナビゲーションスワイプ |
| 22 | kIOHIDEventTypeZoomToggle | スマートズーム |
| 23 | kIOHIDEventTypeDockSwipe | Dockスワイプ |

### 2.3 フェーズ関連フィールド

| フィールド番号 | 説明 | 値の範囲 |
|--------------|------|----------|
| 99 | スクロールイベントフェーズ | IOHIDEventPhaseBits |
| 123 | モメンタムスクロールフェーズ | CGMomentumScrollPhase |
| 132 | ジェスチャーイベントフェーズ | IOHIDEventPhaseBits |

### 2.4 スクロール関連フィールド

| フィールド番号 | 説明 | 単位 |
|--------------|------|------|
| 11 | Y軸デルタ（行単位） | 行 |
| 12 | X軸デルタ（行単位） | 行 |
| 93 | Y軸固定小数点デルタ | 固定小数点 |
| 94 | X軸固定小数点デルタ | 固定小数点 |
| 96 | Y軸ピクセルデルタ | ピクセル |
| 97 | X軸ピクセルデルタ | ピクセル |
| 88 | 連続スクロールフラグ | 1 = 連続 |
| 137 | 方向反転フラグ | 0/1 |

## 3. 各ジェスチャータイプの実装

### 3.1 ナビゲーションスワイプ（ブラウザ前後）

最もシンプルな実装。Safariなどでページの前後移動に使用。

```objc
// 実装例
CGEventRef e = CGEventCreate(NULL);
CGEventSetIntegerValueField(e, 55, NSEventTypeGesture);        // イベントタイプ
CGEventSetIntegerValueField(e, 110, kIOHIDEventTypeNavigationSwipe);  // サブタイプ
CGEventSetIntegerValueField(e, 132, kIOHIDEventPhaseBegan);    // フェーズ
CGEventSetIntegerValueField(e, 115, dir);                      // 方向（左右）

CGEventPost(kCGHIDEventTap, e);

// 終了イベント
CGEventSetIntegerValueField(e, 115, kIOHIDSwipeNone);
CGEventSetIntegerValueField(e, 132, kIOHIDEventPhaseEnded);
CGEventPost(kCGHIDEventTap, e);
CFRelease(e);
```

**重要な点:**
- 開始フェーズと終了フェーズの2つのイベントを送信
- 方向は`kIOHIDSwipeLeft` (4) または `kIOHIDSwipeRight` (8)
- `kCGHIDEventTap`を使用（`kCGSessionEventTap`ではない）

### 3.2 Dockスワイプ（Mission Control、Spaces切り替え）

より複雑な実装。2つの異なるイベントタイプを組み合わせて送信。

```objc
// タイプ29イベント（ジェスチャー）
CGEventRef e29 = CGEventCreate(NULL);
CGEventSetDoubleValueField(e29, 55, NSEventTypeGesture);
CGEventSetDoubleValueField(e29, 41, 33231);  // 特殊な定数

// タイプ30イベント（マグニファイ）
CGEventRef e30 = CGEventCreate(NULL);
CGEventSetDoubleValueField(e30, 55, NSEventTypeMagnify);
CGEventSetDoubleValueField(e30, 110, kIOHIDEventTypeDockSwipe);
CGEventSetDoubleValueField(e30, 132, phase);
CGEventSetDoubleValueField(e30, 124, originOffset);  // 累積オフセット

// DockSwipeタイプに応じた特殊値
double weirdTypeOrSum;
if (type == kMFDockSwipeTypeHorizontal) {
    weirdTypeOrSum = 1.401298464324817e-45;  // 水平（Spaces切り替え）
} else if (type == kMFDockSwipeTypeVertical) {
    weirdTypeOrSum = 2.802596928649634e-45;  // 垂直（Mission Control）
} else if (type == kMFDockSwipeTypePinch) {
    weirdTypeOrSum = 4.203895392974451e-45;  // ピンチ（デスクトップ表示）
}
CGEventSetDoubleValueField(e30, 119, weirdTypeOrSum);

// 両イベントを送信（順序重要）
CGEventPost(kCGSessionEventTap, e30);
CGEventPost(kCGSessionEventTap, e29);
```

**重要な点:**
- originOffsetは累積値として管理
- 終了フェーズではexitSpeedを設定（フィールド129, 130）
- `kCGSessionEventTap`を使用
- "stuck bug"対策として、終了イベントを200ms後と500ms後に再送信

### 3.3 ジェスチャースクロール（2本指スクロール）

最も複雑な実装。スクロールイベントとジェスチャーイベントを組み合わせ。

```objc
// タイプ22イベント（スクロール）
CGEventRef e22 = CGEventCreate(NULL);
CGEventSetIntegerValueField(e22, 55, 22);  // NSEventTypeScrollWheel
CGEventSetIntegerValueField(e22, 88, 1);   // 連続スクロール

// 3種類のデルタを設定
CGEventSetIntegerValueField(e22, 11, vecScrollLineInt.y);   // 行単位Y
CGEventSetIntegerValueField(e22, 96, vecScrollPoint.y);      // ピクセル単位Y
CGEventSetIntegerValueField(e22, 93, fixedScrollDelta(vecScrollLine.y)); // 固定小数点Y

CGEventSetIntegerValueField(e22, 12, vecScrollLineInt.x);   // 行単位X
CGEventSetIntegerValueField(e22, 97, vecScrollPoint.x);      // ピクセル単位X
CGEventSetIntegerValueField(e22, 94, fixedScrollDelta(vecScrollLine.x)); // 固定小数点X

// フェーズ設定
CGEventSetIntegerValueField(e22, 99, phase);
CGEventSetIntegerValueField(e22, 123, momentumPhase);

// タイプ29イベント（ジェスチャー）も送信
if (phase != kIOHIDEventPhaseUndefined) {
    CGEventRef e29 = CGEventCreate(NULL);
    CGEventSetIntegerValueField(e29, 55, 29);  // NSEventTypeGesture
    CGEventSetIntegerValueField(e29, 110, 6);  // kIOHIDEventTypeScroll
    // ジェスチャーデルタ設定...
}
```

**デルタベクトルの計算:**
- **scrollPoint**: ピクセル単位の生のデルタ
- **scrollLine**: scrollPoint / 10（CGEventSource.pixelsPerLineのデフォルト値）
- **scrollLineInt**: 特殊な丸め処理（0-1は切り上げ、それ以外は切り捨て）
- **gesture**: scrollPoint * 1.67（ページスワイプの感度調整）

## 4. 高度な機能

### 4.1 TouchAnimator（スムージング）

高負荷時のパフォーマンス低下による不規則なイベントタイミングを平滑化。

```swift
// 基本的な使用方法
touchAnimator.start(params: { valueLeft, isRunning, curve, currentSpeed in
    // パラメータ計算
    return [
        "vector": combinedVector,
        "duration": 3.0/60.0,  // 3フレーム分
        "curve": linearCurve
    ]
}) { deltaVec, phase, momentumHint in
    // スムージングされたデルタでイベント送信
}
```

**設定推奨値:**
- 継続時間: 3.0/60.0秒（3フレーム@60Hz）
- カーブ: リニア
- サブピクセレーター: バイアス付き（最初のフレームで必ず非ゼロデルタを生成）

### 4.2 慣性スクロール（Momentum Scroll）

物理的な減速をシミュレート。

```swift
// パラメータ
let stopSpeed = 1.0        // 停止速度閾値
let dragCoeff = 30.0       // ドラッグ係数
let dragExp = 0.7          // ドラッグ指数

// DragCurveによる減速計算
let animationCurve = DragCurve(
    coefficient: dragCoeff,
    exponent: dragExp,
    initialSpeed: initialSpeed,
    stopSpeed: stopSpeed
)
```

**実装の流れ:**
1. 最後のイベントから出口速度を計算
2. DragCurveで減速アニメーションを生成
3. DisplayLinkと同期してイベントを送信
4. 速度がstopSpeedを下回ったら終了

### 4.3 サブピクセル処理

整数への丸めによる精度損失を防ぐ。

```objc
// VectorSubPixelatorの使用
static VectorSubPixelator *pixelator = [VectorSubPixelator biasedPixelator];

// リセット（ジェスチャー開始時）
[pixelator reset];

// 整数ベクトルの取得
Vector intVector = [pixelator intVectorWithDoubleVector:doubleVector];
```

## 5. エラー処理とタイミング

### 5.1 "Stuck Bug"対策

システム負荷時にDockSwipeの終了イベントが無視される問題への対処。

```objc
// 終了イベントの再送信
dispatch_async(dispatch_get_main_queue(), ^{
    // 200ms後に再送信
    _doubleSendTimer = [NSTimer scheduledTimerWithTimeInterval:0.2 
                                                         target:self 
                                                       selector:@selector(resendEndEvent:) 
                                                       userInfo:events 
                                                        repeats:NO];
    // 500ms後にも再送信
    _tripleSendTimer = [NSTimer scheduledTimerWithTimeInterval:0.5 
                                                         target:self 
                                                       selector:@selector(resendEndEvent:) 
                                                       userInfo:events 
                                                        repeats:NO];
});
```

### 5.2 スレッド管理

```objc
// 専用キューの作成
dispatch_queue_attr_t attr = dispatch_queue_attr_make_with_qos_class(
    DISPATCH_QUEUE_SERIAL, 
    QOS_CLASS_USER_INTERACTIVE, 
    -1
);
_momentumQueue = dispatch_queue_create("com.example.gesture-scroll", attr);

// 同期グループでの待機
dispatch_group_t waitGroup = dispatch_group_create();
dispatch_group_enter(waitGroup);
// ... 非同期処理 ...
dispatch_group_leave(waitGroup);
dispatch_group_wait(waitGroup, DISPATCH_TIME_FOREVER);
```

## 6. 実装上の注意点

### 6.1 制限事項
- 生のタッチ座標情報は含まれない（トラックパッドとは異なる）
- Chromiumベースのアプリで制限あり（Issue 40322807）
- 未文書化APIのため、将来のmacOSで動作しなくなる可能性

### 6.2 パフォーマンス考慮事項
- イベント送信は高優先度キューで実行
- DisplayLinkと同期して最大60Hz（または画面リフレッシュレート）
- 複数の非同期キューの調整が必要

### 6.3 デバッグ推奨事項
- `CGEventTap`でイベントをモニタリング
- タイムスタンプとフェーズの追跡
- サブピクセレーターの状態確認

## 7. 実装順序の推奨

1. **基本的なナビゲーションスワイプ**から開始
   - 最もシンプルで検証しやすい
   - ブラウザでの動作確認が容易

2. **ジェスチャースクロール**の実装
   - アプリケーションの互換性が高い
   - ユーザー体験への影響が大きい

3. **Dockスワイプ**の実装
   - より複雑だが視覚的にわかりやすい
   - システムレベルの機能統合

4. **高度な機能**の追加
   - TouchAnimatorによるスムージング
   - 慣性スクロール
   - エラー処理の強化

## 8. テスト方法

### 8.1 基本動作確認
- Safari: ナビゲーションスワイプでページ前後移動
- Finder: ジェスチャースクロールでファイルリスト操作
- Mission Control: 3本指上スワイプで起動

### 8.2 エッジケース
- 高CPU負荷時の動作
- 複数ディスプレイ環境
- 異なるリフレッシュレート（30Hz, 60Hz, 120Hz）

### 8.3 互換性テスト
- ネイティブアプリ（Safari, Mail, Finder）
- Electronアプリ（VS Code, Slack）
- Chromiumベースブラウザ

## まとめ

この実装は、macOSの内部動作を深く理解し、多くの試行錯誤を経て作られています。未文書化APIの使用はリスクを伴いますが、ネイティブなトラックパッド体験に近い操作感を提供できます。実装時は、まず基本的な機能から始め、段階的に高度な機能を追加することを推奨します。