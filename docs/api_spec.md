# Keyball Gestures API仕様書

このドキュメントは、Keyball GesturesのREST APIインターフェースについて説明します。

## 概要

Keyball Gestures APIは、ジェスチャー認識サービスを制御し、設定を管理するためのRESTfulなインターフェースを提供します。デフォルトではポート8080で動作し、JSON形式でデータをやり取りします。

## 基本情報

- ベースURL: `http://localhost:8080`
- レスポンス形式: JSON
- 認証: 現在の実装では認証は必要ありません

## エンドポイント一覧

### 設定関連

#### 現在の設定を取得

```
GET /api/config
```

**レスポンス**:

```json
{
  "touchpad": {
    "min_x": 0,
    "max_x": 32767,
    "min_y": 0,
    "max_y": 32767
  },
  "input": {
    "two_finger_key": 184,
    "four_finger_key": 183
  },
  "motion": {
    "filter_smoothing_factor": 0.85,
    "filter_warm_up_count": 10,
    "mouse_delta_factor": 15
  },
  "gesture": {
    "reset_threshold": "50ms"
  },
  "device_prefs": {
    "preferred_keyboard_device": "",
    "preferred_mouse_device": ""
  }
}
```

#### 設定を更新

```
PUT /api/config
```

**リクエスト本文**:

```json
{
  "touchpad": {
    "min_x": 0,
    "max_x": 32767,
    "min_y": 0,
    "max_y": 32767
  },
  "input": {
    "two_finger_key": 184,
    "four_finger_key": 183
  },
  "motion": {
    "filter_smoothing_factor": 0.85,
    "filter_warm_up_count": 10,
    "mouse_delta_factor": 15
  },
  "gesture": {
    "reset_threshold": "50ms"
  },
  "device_prefs": {
    "preferred_keyboard_device": "",
    "preferred_mouse_device": ""
  }
}
```

**レスポンス**:

```json
{
  "status": "success"
}
```

#### 設定をファイルに保存

```
POST /api/config/save
```

**リクエスト本文**:

```json
{
  "path": "/path/to/config.toml"
}
```

パスを空にすると、デフォルトの設定パス（`~/.config/keyball-gestures/config.toml`）に保存されます。

**レスポンス**:

```json
{
  "status": "success",
  "path": "/path/to/config.toml"
}
```

### デバイス関連

#### 利用可能なデバイス一覧を取得

```
GET /api/devices
```

**レスポンス**:

```json
[
  {
    "name": "Keyball Keyboard",
    "path": "/dev/input/event3",
    "type": "keyboard"
  },
  {
    "name": "Keyball Mouse",
    "path": "/dev/input/event4",
    "type": "mouse"
  }
]
```

#### 優先デバイスを設定

```
PUT /api/devices/preferred
```

**リクエスト本文**:

```json
{
  "keyboard_device": "Keyball Keyboard",
  "mouse_device": "Keyball Mouse"
}
```

**レスポンス**:

```json
{
  "status": "success"
}
```

### サービス関連

#### ジェスチャー認識サービスを開始

```
POST /api/service/start
```

**レスポンス**:

```json
{
  "status": "started"
}
```

または、既に実行中の場合:

```json
{
  "status": "already_running"
}
```

#### ジェスチャー認識サービスを停止

```
POST /api/service/stop
```

**レスポンス**:

```json
{
  "status": "stopped"
}
```

または、実行中でない場合:

```json
{
  "status": "not_running"
}
```

#### サービスの状態を確認

```
GET /api/service/status
```

**レスポンス**:

```json
{
  "status": "running"
}
```

または:

```json
{
  "status": "stopped"
}
```

### ヘルスチェック

#### サーバーの状態を確認

```
GET /api/health
```

**レスポンス**:

```json
{
  "status": "ok"
}
```

## ステータスコード

- `200 OK`: リクエストが成功しました
- `400 Bad Request`: リクエストが不正です
- `500 Internal Server Error`: サーバー内部でエラーが発生しました

## エラーレスポンス

エラーが発生した場合、次の形式でレスポンスが返されます:

```json
{
  "error": "エラーメッセージ"
}
