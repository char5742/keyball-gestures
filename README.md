# Keyball Gestures

Keyball GesturesはLinux環境でMacOSライクなマウスジェスチャーを実現するツールです。Keyballのトラックボールとキーの組み合わせで、直感的なワークスペース操作を可能にします。

## ドキュメント

詳細なドキュメントはdocsディレクトリにあります：

- [アーキテクチャ設計](docs/architecture.md) - システム全体のアーキテクチャと構成
- [API仕様](docs/api_spec.md) - REST APIの詳細な仕様

## 特徴

- 2本指/4本指スワイプジェスチャーのエミュレーション
- 滑らかな動作を実現するモーションフィルター搭載
- デバイスの自動再接続機能（切断・再接続時に自動で復帰）
- デバイスの健全性チェック機能（定期的にデバイスの状態を確認）
- APIインターフェースによる外部アプリケーション（例: Flutterフロントエンド）との連携
- キーボードとマウスの自動検出および優先デバイス設定
- 設定ファイルによるカスタマイズ

## 動作環境

- **OS**: Pop!_OS COSMIC alpha 6（他のLinuxディストリビューションでも動作する可能性あり）
- **必須デバイス**: Keyball
- **Go**: version 1.24.0 以上

## インストール

### 方法1: スクリプトを使用 (推奨)

インストールスクリプトを実行すると、ビルドと必要な設定（udevルール、systemdサービス）が自動で行われます。

```sh
curl -sSL https://raw.githubusercontent.com/char5742/keyball-gestures/main/scripts/install.sh | bash
```
インストール後、`keyball-gestures` サービスが自動で起動します。

### 方法2: 手動ビルド

```sh
git clone https://github.com/char5742/keyball-gestures.git
cd keyball-gestures
go build cmd/main.go
```
この場合、udevルールの設定やサービスの登録は手動で行う必要があります。

## 使用方法 (手動ビルドの場合)

1.  **root権限で実行**:
    仮想デバイス(`/dev/uinput`)を作成・アクセスするためにroot権限が必要です。

    ```sh
    sudo ./main [options]
    ```

    **オプション**:
    ```
    Options:
      --api         APIサーバーモードで起動します（デフォルトはCLIモード）
      --port int    APIサーバーのポート番号（デフォルト: 8080）
      --config string 設定ファイルのパス（指定しない場合はデフォルトパスを使用）
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

- **設定**:
    - `GET /api/config`: 現在の設定を取得
    - `PUT /api/config`: 設定を更新 (メモリ上)
    - `POST /api/config/save`: 現在の設定を指定パスまたはデフォルトパスに保存
- **デバイス**:
    - `GET /api/devices`: 利用可能なデバイス一覧を取得
    - `PUT /api/devices/preferred`: 優先するキーボード/マウスデバイスを設定
- **サービス**:
    - `POST /api/service/start`: ジェスチャー認識サービスを開始
    - `POST /api/service/stop`: ジェスチャー認識サービスを停止
    - `GET /api/service/status`: サービスの状態を確認 (running/stopped)
- **その他**:
    - `GET /api/health`: サーバーのヘルスチェック

## 設定ファイル

設定はTOML形式で行います。設定ファイルのデフォルトパスは `~/.config/keyball-gestures/config.toml` です。
初回起動時や設定ファイルが存在しない場合は、デフォルト設定で動作します。
設定例については `example-config.toml` を参照してください。

```toml
# example-config.toml の内容例

# 仮想タッチパッドの解像度
[touchpad]
min_x = 0
max_x = 32767
min_y = 0
max_y = 32767

# ジェスチャーを発動するためのキーコード (evtestなどで確認可能)
[input]
two_finger_key = 184  # 例: F14キー
four_finger_key = 183 # 例: F13キー

# マウス移動のスムージングと感度設定
[motion]
filter_smoothing_factor = 0.85 # スムージング係数 (0.0 - 1.0)
filter_warm_up_count = 10      # スムージング開始までのウォームアップ回数
mouse_delta_factor = 15        # マウス移動量の倍率

# ジェスチャーのリセット閾値 (この時間操作がないと指の位置をリセット)
[gesture]
reset_threshold = "50ms"

# 優先デバイス設定 (デバイス名の一部を指定)
# 空欄の場合は最初に見つかったデバイスを使用
[device_prefs]
preferred_keyboard_device = "" # 例: "Keychron"
preferred_mouse_device = ""    # 例: "Logicool"
```

## 注意事項

- **root権限**: 仮想デバイス(`/dev/uinput`)の作成・アクセスにroot権限が必要です。インストールスクリプトを使用すると、udevルールにより一般ユーザーでの実行が可能になる場合があります。
- **対応OS**: 主にPop!_OS COSMICでテストされていますが、他のLinuxディストリビューションでも動作する可能性があります。
- **デバイス検出**: キーボードとマウスは自動検出されます。複数接続されている場合、デフォルトでは最初に見つかったデバイスが使用されますが、設定ファイルで優先デバイスを指定できます。
- **自動再接続**: デバイスが切断された場合、自動的に再接続を試みます。この機能はサービス内で有効/無効を切り替え可能です（API経由での制御は未実装）。
- **健全性チェック**: 定期的にデバイスの応答を確認し、問題があれば再接続を試みます。
- **APIモード**: APIモードで起動すると、HTTP経由で外部アプリケーションからサービスを制御できます。

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
