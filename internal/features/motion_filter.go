package features

import "time"

// MotionFilter はマウスの移動値（dx, dy）を滑らかにします
type MotionFilter struct {
	smoothingFactor float64 // 0.0-1.0の範囲。1.0に近いほど滑らかになりますが、遅延が大きくなります
	lastDX          float64
	lastDY          float64
	lastTime        time.Time
	warmUpCount     int
	currentCount    int
	initialized     bool
}

// 新しいモーションフィルターを作成します
func NewMotionFilter(smoothingFactor float64, warmUpCount int) *MotionFilter {
	return &MotionFilter{
		smoothingFactor: smoothingFactor,
		warmUpCount:     warmUpCount,
		lastTime:        time.Now(),
	}
}

// raw dx, dy値にsmoothingを適用します
func (mf *MotionFilter) Filter(dxRaw, dyRaw int32) (int32, int32) {
	now := time.Now()
	mf.lastTime = now

	// 初回または未初期化の場合
	if !mf.initialized || mf.currentCount < mf.warmUpCount {
		mf.currentCount++
		mf.lastDX = float64(dxRaw)
		mf.lastDY = float64(dyRaw)
		mf.initialized = true
		return dxRaw, dyRaw
	}

	// smoothingの適用
	f := mf.smoothingFactor
	newDX := float64(dxRaw)*(1.0-f) + mf.lastDX*f
	newDY := float64(dyRaw)*(1.0-f) + mf.lastDY*f

	// 新しい値を保存
	mf.lastDX = newDX
	mf.lastDY = newDY

	return int32(newDX + 0.5), int32(newDY + 0.5)
}

// フィルターの状態をリセットします
func (mf *MotionFilter) Reset() {
	mf.lastDX = 0
	mf.lastDY = 0
	mf.currentCount = 0
	mf.initialized = false
	mf.lastTime = time.Now()
}
