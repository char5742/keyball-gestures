# Keyball Gestures

Keyball GesturesはLinux環境でMacOSライクなマウスジェスチャーを実現するツールです。Keyballのトラックボールとキーの組み合わせで、直感的なワークスペース操作を可能にします。

## ドキュメント

詳細なドキュメントはdocsディレクトリにあります：

- [アーキテクチャ設計](docs/architecture.md) - システム全体のアーキテクチャと構成
- [API仕様](docs/api_spec.md) - REST APIの詳細な仕様

## 特徴

- 2本指/4本指スワイプジェスチャーのエミュレーション
- 滑らかな動作を実現するモーションフィルター搭載
- APIインターフェースによるFlutterフロントエンドとの連携
- キーボードとマウスの自動検出
- 設定ファイルによるカスタマイズ

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
Usage:
  xhost +SI:localuser:root
  sudo DISPLAY=:1 ./main [options]

Options:
  --api         APIサーバーモードで起動します（デフォルトはCLIモード）
  --port        APIサーバーのポート番号（デフォルト: 8080）
  --config      設定ファイルのパス（指定しない場合はデフォルトパスを使用）
```

2. ジェスチャー操作:
   - **2本指スワイプ**: F14キーを押しながらトラックボール操作
   - **4本指スワイプ**: F13キーを押しながらトラックボール操作

## 動作モード

### CLIモード (デフォルト)
コマンドラインからジェスチャー認識サービスを直接実行します。

```sh
sudo ./main
```

### APIモード
HTTPサーバーを起動し、RESTful APIを通じてサービスの制御や設定の管理が可能になります。
Flutterフロントエンドとの連携を想定しています。

```sh
sudo ./main --api --port 8080
```

#### 主なAPIエンドポイント

- `GET /api/config` - 現在の設定を取得
- `PUT /api/config` - 設定を更新
- `GET /api/devices` - 利用可能なデバイス一覧を取得
- `POST /api/service/start` - ジェスチャー認識サービスを開始
- `POST /api/service/stop` - ジェスチャー認識サービスを停止
- `GET /api/service/status` - サービスの状態を確認

## 設定ファイル

デフォルトでは `~/.config/keyball-gestures/config.toml` に設定ファイルが作成されます。
以下のパラメータが設定可能です：

```toml
[touchpad]
min_x = 0
max_x = 32767
min_y = 0
max_y = 32767

[input]
two_finger_key = 184  # F14
four_finger_key = 183  # F13

[motion]
filter_smoothing_factor = 0.85
filter_warm_up_count = 10
mouse_delta_factor = 15

[gesture]
reset_threshold = "50ms"

[device_prefs]
preferred_keyboard_device = ""
preferred_mouse_device = ""
```

## 注意事項

- rootユーザー権限が必要です（/dev/uinputへのアクセスのため）
- 現在のところ、Pop!_OS COSMICでの動作を主に確認しています
- キーボードとマウスは自動で検出されますが、複数接続されている場合は最初に見つかったデバイスが使用されます
- APIモードでは、任意のFlutterアプリからHTTP経由でサービスを制御できます

## ライセンス

MIT

## 開発者向け情報

プロジェクトの構造:
```
.
├── cmd/
│   └── main.go          # メインエントリーポイント
└── internal/
    ├── api/             # APIサーバー実装
    │   ├── server.go    # サーバー本体
    │   ├── routes.go    # APIルート定義
    │   └── service.go   # ジェスチャーサービス
    ├── config/          # 設定関連
    ├── consts/          # 定数定義
    ├── features/        # 主要機能実装
    ├── types/           # 型定義
    └── utils/           # ユーティリティ関数
