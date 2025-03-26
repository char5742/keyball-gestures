package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

// Config はアプリケーション全体の設定を表す構造体
type Config struct {
	TouchPad    TouchPadConfig    `toml:"touchpad"`
	Input       InputConfig       `toml:"input"`
	Motion      MotionConfig      `toml:"motion"`
	Gesture     GestureConfig     `toml:"gesture"`
	DevicePrefs DevicePrefsConfig `toml:"device_prefs"`
}

// TouchPadConfig は仮想タッチパッドの設定
type TouchPadConfig struct {
	MinX int32 `toml:"min_x"`
	MaxX int32 `toml:"max_x"`
	MinY int32 `toml:"min_y"`
	MaxY int32 `toml:"max_y"`
}

// InputConfig はキー入力の設定
type InputConfig struct {
	TwoFingerKey  int `toml:"two_finger_key"`
	FourFingerKey int `toml:"four_finger_key"`
}

// MotionConfig はモーション制御の設定
type MotionConfig struct {
	FilterSmoothingFactor float64 `toml:"filter_smoothing_factor"`
	FilterWarmUpCount     int     `toml:"filter_warm_up_count"`
	MouseDeltaFactor      int     `toml:"mouse_delta_factor"`
}

// GestureConfig はジェスチャー認識の設定
type GestureConfig struct {
	ResetThreshold time.Duration `toml:"reset_threshold"`
}

// DevicePrefsConfig はデバイス設定の設定
type DevicePrefsConfig struct {
	PreferredKeyboardDevice string `toml:"preferred_keyboard_device"`
	PreferredMouseDevice    string `toml:"preferred_mouse_device"`
}

// DefaultConfig はデフォルト設定を返す
func DefaultConfig() *Config {
	return &Config{
		TouchPad: TouchPadConfig{
			MinX: 0,
			MaxX: 32767,
			MinY: 0,
			MaxY: 32767,
		},
		Input: InputConfig{
			TwoFingerKey:  184, // F14
			FourFingerKey: 183, // F13
		},
		Motion: MotionConfig{
			FilterSmoothingFactor: 0.85,
			FilterWarmUpCount:     10,
			MouseDeltaFactor:      15,
		},
		Gesture: GestureConfig{
			ResetThreshold: 50 * time.Millisecond,
		},
		DevicePrefs: DevicePrefsConfig{
			PreferredKeyboardDevice: "",
			PreferredMouseDevice:    "",
		},
	}
}

// LoadConfig は設定ファイルから設定を読み込む
func LoadConfig(configPath string) (*Config, error) {
	// デフォルト設定を用意
	config := DefaultConfig()

	// ファイルが存在しない場合はデフォルト設定を保存して返す
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 設定ディレクトリの作成
		configDir := filepath.Dir(configPath)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return config, err
		}

		// デフォルト設定の保存
		if err := SaveConfig(configPath, config); err != nil {
			return config, err
		}

		return config, nil
	}

	// 設定ファイルの読み込み
	if _, err := toml.DecodeFile(configPath, config); err != nil {
		return config, err
	}

	return config, nil
}

// SaveConfig は設定をTOMLファイルに保存する
func SaveConfig(configPath string, config *Config) error {
	// 設定ディレクトリの作成
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// ファイルを開く（なければ作成）
	f, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// TOML形式でエンコードして書き込み
	encoder := toml.NewEncoder(f)
	return encoder.Encode(config)
}
