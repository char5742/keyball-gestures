package features

// Device デバイス情報を表す構造体
type Device struct {
	Name string
	Path string
	Type DeviceType
}

// DeviceType デバイスタイプを表す列挙型
type DeviceType int

const (
	DeviceTypeKeyboard DeviceType = iota
	DeviceTypeMouse
)

// DeviceEventType はデバイスイベントの種類を表す
type DeviceEventType int

const (
	DeviceAdded DeviceEventType = iota
	DeviceRemoved
	DeviceChanged
)

// DeviceEvent はデバイスの変更イベントを表す
type DeviceEvent struct {
	Type   DeviceEventType
	Device *Device
	Path   string
}

// DeviceCallback はデバイスイベント発生時に呼び出されるコールバック関数の型
type DeviceCallback func(event DeviceEvent)