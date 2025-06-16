//go:build darwin
// +build darwin

package darwin

// macOSのキーコード定数
const (
	// 修飾キー
	KeyCodeFn         = 0x3F  // Fnキー
	KeyCodeCommand    = 0x37  // Commandキー
	KeyCodeShift      = 0x38  // Shiftキー
	KeyCodeCapsLock   = 0x39  // Caps Lockキー
	KeyCodeOption     = 0x3A  // Optionキー
	KeyCodeControl    = 0x3B  // Controlキー
	KeyCodeRightShift = 0x3C  // 右Shiftキー
	KeyCodeRightOption = 0x3D // 右Optionキー
	KeyCodeRightControl = 0x3E // 右Controlキー
	
	// 数字キー
	KeyCode0 = 0x1D
	KeyCode1 = 0x12
	KeyCode2 = 0x13
	KeyCode3 = 0x14
	KeyCode4 = 0x15
	KeyCode5 = 0x17
	KeyCode6 = 0x16
	KeyCode7 = 0x1A
	KeyCode8 = 0x1C
	KeyCode9 = 0x19
	
	// 文字キー
	KeyCodeA = 0x00
	KeyCodeB = 0x0B
	KeyCodeC = 0x08
	KeyCodeD = 0x02
	KeyCodeE = 0x0E
	KeyCodeF = 0x03
	KeyCodeG = 0x05
	KeyCodeH = 0x04
	KeyCodeI = 0x22
	KeyCodeJ = 0x26
	KeyCodeK = 0x28
	KeyCodeL = 0x25
	KeyCodeM = 0x2E
	KeyCodeN = 0x2D
	KeyCodeO = 0x1F
	KeyCodeP = 0x23
	KeyCodeQ = 0x0C
	KeyCodeR = 0x0F
	KeyCodeS = 0x01
	KeyCodeT = 0x11
	KeyCodeU = 0x20
	KeyCodeV = 0x09
	KeyCodeW = 0x0D
	KeyCodeX = 0x07
	KeyCodeY = 0x10
	KeyCodeZ = 0x06
)

// ConvertConfigKeyCode は設定ファイルのキーコードをmacOSのキーコードに変換する
func ConvertConfigKeyCode(configKey int) int32 {
	// 設定ファイルではLinuxのキーコードを使用しているため変換が必要
	// ここでは簡易的にFnキーのマッピングのみ実装
	switch configKey {
	case 464: // Linux KEY_FN
		return KeyCodeFn
	default:
		// その他のキーは未実装
		return int32(configKey)
	}
}