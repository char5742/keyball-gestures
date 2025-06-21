# トラックパッドイベント解析結果

## ログの読み方 — 何が「本物」で何が「自作」か？

| 行例 | 正体 | 注目フィールド |
| --- | --- | --- |
| `type=29 phase=? subtype=0 mask=0x0` が連続 | **自作イベント**<br>phase (132)・subtype (110)・mask (134) すべて 0 → macOS にとっては「中身ゼロのゴミ」 | |
| `type=29 phase=Began/Changed/Ended subtype=6` ＋前後に **Scroll Δ** | **ネイティブ 2本指スクロール**<br>subtype 6 = *kIOHIDEventTypeScroll*。Point Δ (96/97) が載っている | |
| `type=30 phase=Began/Changed/Cancelled subtype=23 mask=0x1/0x2/0x8` | **ネイティブ 3-/4-finger システムジェスチャー** (Spaces/Exposé)<br>特徴: ① *type=30* (GestureChange) から始まる ② *subtype=23* ③ Δは常に 0、動き量は *mask & flags* に浮動小数で載る | |

## 重要な発見

4本指スワイプをエミュレートするなら **`subtype = 23` + `type 30/31` + 適切な `mask`** を出さないと macOS が反応しません。
`subtype 15` や `flags 0x20` は High Sierra 時代のピンチズーム互換コードで、 Sonoma/Sequoia では無効です。

## 次のアクション（ジェスチャー生成側）

1. **subtype を 23 に**
   ```c
   CGEventSetIntegerValueField(evt, kCGEventFieldIOHIDEventSubtype, 23);
   ```

2. **type と phase の組み合わせを実物通りに**
   - Began/Changed/Ended すべて *type=30*（GestureChange）で来ている
   - 始動トリガー（以前の type 29）は macOS が内部で挿むので自前で送らなくて良い

3. **Δ は 0、代わりに flags に float で距離を入れる**
   ```c
   float distance = -200.0f;
   memcpy(&flags, &distance, 4);
   CGEventSetIntegerValueField(evt, 115, flags);
   ```

4. **mask ビットで「スワイプ方向 & 指本数」を伝える**
   - 左=0x1 右=0x2 上=0x4 下=0x8
   - 4-finger の時は **mask に 0x1/2/4/8 ＋ 0x10** が立つ説も有り

5. **投稿先は `kCGAnnotatedSessionEventTap`** のままで OK

## ツール一覧

- `trackpad_dump` - 基本的なダンプツール（v1）
- `trackpad_dump2` - 改良版ダンプツール（v2）
  - subtype名称表とmask/flagsのデコード付き
  - 29→30/31のフェーズ遷移も色分け
  - Point Δ (96/97)／Fixed Δ (93/94)も両方拾う