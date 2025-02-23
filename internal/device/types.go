package device

// UIInput デバイスの定数（uinput.hから）
const (
	MaxNameSize = 80         // デバイス名の最大サイズ
	DevCreate   = 0x5501     // デバイス作成用のIOCTL
	DevDestroy  = 0x5502     // デバイス破棄用のIOCTL
	SetEvBit    = 0x40045564 // イベントビット設定用のIOCTL
	SetKeyBit   = 0x40045565 // キービット設定用のIOCTL
	SetAbsBit   = 0x40045567 // 絶対座標ビット設定用のIOCTL
	BusUsb      = 0x03       // USBバスタイプ
)

// その他のデバイス制御用定数
const (
	AbsSize       = 64         // 絶対座標の配列サイズ
	EVIOCGRAB     = 0x40044590 // デバイスの排他制御用のIOCTL
	PropPointer   = 0x00       // ポインターデバイスプロパティ
	PropButtonpad = 0x02       // ボタンパッドプロパティ
	SetPropBit    = 0x4004556a // プロパティビット設定用のIOCTL
)

// InputID はデバイス識別子を表す構造体
type InputID struct {
	Bustype uint16 // バスタイプ
	Vendor  uint16 // ベンダーID
	Product uint16 // 製品ID
	Version uint16 // バージョン
}

// UserDev はuinputユーザーデバイスの設定を表す構造体
type UserDev struct {
	Name       [MaxNameSize]byte // デバイス名
	ID         InputID           // デバイス識別子
	EffectsMax uint32            // 最大エフェクト数
	Absmax     [AbsSize]int32    // 絶対座標の最大値
	Absmin     [AbsSize]int32    // 絶対座標の最小値
	Absfuzz    [AbsSize]int32    // 絶対座標のファジー値
	Absflat    [AbsSize]int32    // 絶対座標のフラット値
}
