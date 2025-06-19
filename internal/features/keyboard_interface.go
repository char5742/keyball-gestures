package features

// Keyboard キーボードからの入力を処理するインターフェース
type Keyboard interface {
	GetKey() (key int32)
	Close() error
}