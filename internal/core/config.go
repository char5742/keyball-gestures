package core

// GestureConfig はジェスチャー認識の設定
type GestureConfig struct {
	// TwoFingerKey は2本指ジェスチャーのトリガーキー
	TwoFingerKey int
	// FourFingerKey は4本指ジェスチャーのトリガーキー
	FourFingerKey int
	// ResetThreshold はジェスチャーリセットまでの時間閾値
	ResetThreshold int64 // ミリ秒
	// MouseDeltaFactor はマウス移動量の倍率
	MouseDeltaFactor float64
	// SmoothingFactor はモーションフィルターの平滑化係数
	SmoothingFactor float64
	// WarmUpCount はモーションフィルターのウォームアップカウント
	WarmUpCount int
}