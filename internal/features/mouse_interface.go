package features

// Mouse マウス入力を扱うインターフェース
type Mouse interface {
	HandleSignals()
	// マウスの移動量を取得する
	GetMouseDelta() (dx int32, dy int32)
	// マウス操作を専有する
	Grab() error
	// マウス操作の専有を解除する
	Release() error
	Close() error
}