package gui

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Config はアプリケーションの設定を表す構造体
type Config struct {
	TwoFingerKey            int     `json:"twoFingerKey"`
	FourFingerKey           int     `json:"fourFingerKey"`
	MotionFilterFactor      float64 `json:"motionFilterFactor"`
	MotionFilterWarmUpCount int     `json:"motionFilterWarmUpCount"`
	MouseDeltaFactor        int     `json:"mouseDeltaFactor"`
	PreferredKeyboardDevice string  `json:"preferredKeyboardDevice"`
	PreferredMouseDevice    string  `json:"preferredMouseDevice"`
	ResetThresholdMillisec  int     `json:"resetThresholdMillisec"`
}

// DefaultConfig はデフォルト設定を返す
func DefaultConfig() Config {
	return Config{
		TwoFingerKey:            184, // F14
		FourFingerKey:           183, // F13
		MotionFilterFactor:      0.85,
		MotionFilterWarmUpCount: 10,
		MouseDeltaFactor:        15,
		ResetThresholdMillisec:  50,
	}
}

// ConfigManager は設定管理のための構造体
type ConfigManager struct {
	app              *App
	config           Config
	configPath       string
	twoFingerKey     *widget.Entry
	fourFingerKey    *widget.Entry
	motionFactor     *widget.Slider
	warmUpCount      *widget.Entry
	mouseDeltaFactor *widget.Entry
	resetThreshold   *widget.Entry
	onSave           func(Config)
}

// NewConfigManager は新しい設定マネージャを作成する
func NewConfigManager(app *App) *ConfigManager {
	// 設定保存先
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}

	configPath := filepath.Join(configDir, "keyball-gestures", "config.json")

	cm := &ConfigManager{
		app:        app,
		configPath: configPath,
		config:     DefaultConfig(),
	}

	// 設定ファイルの読み込み
	if err := cm.LoadConfig(); err != nil {
		// エラーが発生した場合はデフォルト設定を使用
		cm.config = DefaultConfig()
	}

	return cm
}

// LoadConfig は設定ファイルから設定を読み込む
func (cm *ConfigManager) LoadConfig() error {
	// 設定ディレクトリの作成
	configDir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// 設定ファイルが存在しない場合は作成
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		return cm.SaveConfig()
	}

	// 設定ファイルから読み込み
	data, err := ioutil.ReadFile(cm.configPath)
	if err != nil {
		return err
	}

	// JSONデコード
	return json.Unmarshal(data, &cm.config)
}

// SaveConfig は設定をファイルに保存する
func (cm *ConfigManager) SaveConfig() error {
	// 設定ディレクトリの作成
	configDir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// JSONエンコード
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return err
	}

	// ファイルに書き込み
	return ioutil.WriteFile(cm.configPath, data, 0644)
}

// UpdateConfigFromUI はUIから設定を更新する
func (cm *ConfigManager) UpdateConfigFromUI() {
	// 2本指スワイプキー
	if val, err := strconv.Atoi(cm.twoFingerKey.Text); err == nil {
		cm.config.TwoFingerKey = val
	}

	// 4本指スワイプキー
	if val, err := strconv.Atoi(cm.fourFingerKey.Text); err == nil {
		cm.config.FourFingerKey = val
	}

	// モーションフィルター係数
	cm.config.MotionFilterFactor = cm.motionFactor.Value

	// モーションフィルターウォームアップ
	if val, err := strconv.Atoi(cm.warmUpCount.Text); err == nil {
		cm.config.MotionFilterWarmUpCount = val
	}

	// マウス移動係数
	if val, err := strconv.Atoi(cm.mouseDeltaFactor.Text); err == nil {
		cm.config.MouseDeltaFactor = val
	}

	// リセット閾値
	if val, err := strconv.Atoi(cm.resetThreshold.Text); err == nil {
		cm.config.ResetThresholdMillisec = val
	}
}

// CreateConfigPanel は設定パネルを作成する
func (cm *ConfigManager) CreateConfigPanel() fyne.CanvasObject {
	// 設定項目のUIコンポーネント
	cm.twoFingerKey = widget.NewEntry()
	cm.twoFingerKey.SetText(strconv.Itoa(cm.config.TwoFingerKey))

	cm.fourFingerKey = widget.NewEntry()
	cm.fourFingerKey.SetText(strconv.Itoa(cm.config.FourFingerKey))

	cm.motionFactor = widget.NewSlider(0, 1)
	cm.motionFactor.Step = 0.01
	cm.motionFactor.Value = cm.config.MotionFilterFactor

	cm.warmUpCount = widget.NewEntry()
	cm.warmUpCount.SetText(strconv.Itoa(cm.config.MotionFilterWarmUpCount))

	cm.mouseDeltaFactor = widget.NewEntry()
	cm.mouseDeltaFactor.SetText(strconv.Itoa(cm.config.MouseDeltaFactor))

	cm.resetThreshold = widget.NewEntry()
	cm.resetThreshold.SetText(strconv.Itoa(cm.config.ResetThresholdMillisec))

	// 保存ボタン
	saveButton := widget.NewButton("設定を保存", func() {
		cm.UpdateConfigFromUI()
		if err := cm.SaveConfig(); err != nil {
			// エラー表示（実装は簡略化）
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "エラー",
				Content: "設定の保存に失敗しました: " + err.Error(),
			})
			return
		}

		// 保存成功通知
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "成功",
			Content: "設定を保存しました",
		})

		// コールバック呼び出し
		if cm.onSave != nil {
			cm.onSave(cm.config)
		}
	})

	// レイアウト
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "2本指スワイプキー (F14=184)", Widget: cm.twoFingerKey},
			{Text: "4本指スワイプキー (F13=183)", Widget: cm.fourFingerKey},
			{Text: "モーションフィルター係数 (0-1)", Widget: cm.motionFactor},
			{Text: "モーションフィルターウォームアップ", Widget: cm.warmUpCount},
			{Text: "マウス移動量係数", Widget: cm.mouseDeltaFactor},
			{Text: "リセット閾値 (ミリ秒)", Widget: cm.resetThreshold},
		},
	}

	return container.NewVBox(
		widget.NewLabel("ジェスチャー設定"),
		form,
		saveButton,
	)
}

// GetConfig は現在の設定を返す
func (cm *ConfigManager) GetConfig() Config {
	return cm.config
}

// SetOnSave は設定保存時のコールバックを設定する
func (cm *ConfigManager) SetOnSave(callback func(Config)) {
	cm.onSave = callback
}
