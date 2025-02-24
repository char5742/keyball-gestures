# Keyball Gestures

Keyball GesturesはLinux環境でMacOSライクなマウスジェスチャーを実現するツールです。Keyballのトラックボールとキーの組み合わせで、直感的なワークスペース操作を可能にします。

## 特徴

- 2本指/4本指スワイプジェスチャーのエミュレーション
- 滑らかな動作を実現するモーションフィルター搭載
- Pop!_OS COSMICとの完璧な統合
- キーボードとマウスの自動検出

## 動作環境

- **OS**: Pop!_OS COSMIC alpha 6（他のLinuxディストリビューションでも動作する可能性あり）
- **必須デバイス**: Keyball
- **Go**: version 1.24.0 以上

## インストール

```sh
git clone https://github.com/char5742/keyball-gestures.git
cd keyball-gestures
go build cmd/main.go
```

## 使用方法

1. rootユーザー権限で実行:
```sh
sudo ./main
```

2. ジェスチャー操作:
   - **2本指スワイプ**: F14キーを押しながらトラックボール操作
   - **4本指スワイプ**: F13キーを押しながらトラックボール操作

## カスタマイズ

`cmd/main.go`で以下の設定を調整できます:

```go
const (
    // 仮想タッチパッドの範囲設定
    minX = 0
    maxX = 32767
    minY = 0
    maxY = 32767

    // マウスの移動量係数設定
    mouseDeltaFactor = 15

    // トリガーキーの設定
    twoFingerKey  = 184 // F14
    fourFingerKey = 183 // F13

    // モーションフィルターの設定
    motionFilterSmoothingFactor = 0.85
    motionFilterWarmUpCount = 10

    // スクロールをリセットするまでの時間設定
	resetThreshold = 50 * time.Millisecond
)
```

## 注意事項

- rootユーザー権限が必要です（/dev/uinputへのアクセスのため）
- 現在のところ、Pop!_OS COSMICでの動作を主に確認しています
- キーボードとマウスは自動で検出されますが、複数接続されている場合は最初に見つかったデバイスが使用されます

## ライセンス

MIT

## 開発者向け情報

プロジェクトの構造:
```
.
├── cmd/
│   └── main.go          # メインエントリーポイント
└── internal/
    ├── consts/          # 定数定義
    ├── features/        # 主要機能実装
    ├── types/           # 型定義
    └── utils/           # ユーティリティ関数
