package gui

import (
	"os"
	"path/filepath"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/char5742/keyball-gestures/internal/config"
)

// ConfigManager は設定管理のための構造体
type ConfigManager struct {
	app              *App
	config           *config.Config
	configPath       string
	twoFingerKey     *widget.Entry
	fourFingerKey    *widget.Entry
	motionFactor     *widget.Slider
	warmUpCount      *widget.Entry
	mouseDeltaFactor *widget.Entry
	resetThreshold   *widget.Entry
	onSave           func(*config.Config)
}

// NewConfigManager は新しい設定マネージャを作成する
func NewConfigManager(app *App) *ConfigManager {
	// 設定保存先
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}

	configPath := filepath.Join(configDir, "keyball-gestures", "config.toml")

	// 既存のJSON設定ファイルパス（互換性のため）
	oldConfigPath := filepath.Join(configDir, "keyball-gestures", "config.json")

	// 設定マネージャの初期化
	cm := &ConfigManager{
		app:        app,
		configPath: configPath,
	}

	// 設定ファイルの読み込み
	var cfg *config.Config
	cfg, err = config.LoadConfig(configPath)
	if err != nil {
		// TOML設定ファイルの読み込みに失敗した場合

		// 古いJSON設定ファイルを確認
		if _, jsonErr := os.Stat(oldConfigPath); jsonErr == nil {
			// JSON設定ファイルが存在する場合は移行メッセージを表示
			// 実際の移行処理は行わず、デフォルト設定を使用
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "設定ファイル形式の変更",
				Content: "設定ファイル形式がTOMLに変更されました。保存することで新形式に移行します。",
			})
		}

		// デフォルト設定を使用
		cfg = config.DefaultConfig()
	}

	cm.config = cfg
	return cm
}

// SaveConfig は設定をファイルに保存する
func (cm *ConfigManager) SaveConfig() error {
	return config.SaveConfig(cm.configPath, cm.config)
}

// UpdateConfigFromUI はUIから設定を更新する
func (cm *ConfigManager) UpdateConfigFromUI() {
	// 2本指スワイプキー
	if val, err := strconv.Atoi(cm.twoFingerKey.Text); err == nil {
		cm.config.Input.TwoFingerKey = val
	}

	// 4本指スワイプキー
	if val, err := strconv.Atoi(cm.fourFingerKey.Text); err == nil {
		cm.config.Input.FourFingerKey = val
	}

	// モーションフィルター係数
	cm.config.Motion.FilterSmoothingFactor = cm.motionFactor.Value

	// モーションフィルターウォームアップ
	if val, err := strconv.Atoi(cm.warmUpCount.Text); err == nil {
		cm.config.Motion.FilterWarmUpCount = val
	}

	// マウス移動係数
	if val, err := strconv.Atoi(cm.mouseDeltaFactor.Text); err == nil {
		cm.config.Motion.MouseDeltaFactor = val
	}

	// リセット閾値
	if val, err := strconv.Atoi(cm.resetThreshold.Text); err == nil {
		cm.config.Gesture.ResetThreshold = time.Duration(val) * time.Millisecond
	}
}

// CreateConfigPanel は設定パネルを作成する
func (cm *ConfigManager) CreateConfigPanel() fyne.CanvasObject {
	// 設定項目のUIコンポーネント
	cm.twoFingerKey = widget.NewEntry()
	cm.twoFingerKey.SetText(strconv.Itoa(cm.config.Input.TwoFingerKey))

	cm.fourFingerKey = widget.NewEntry()
	cm.fourFingerKey.SetText(strconv.Itoa(cm.config.Input.FourFingerKey))

	cm.motionFactor = widget.NewSlider(0, 1)
	cm.motionFactor.Step = 0.01
	cm.motionFactor.Value = cm.config.Motion.FilterSmoothingFactor

	cm.warmUpCount = widget.NewEntry()
	cm.warmUpCount.SetText(strconv.Itoa(cm.config.Motion.FilterWarmUpCount))

	cm.mouseDeltaFactor = widget.NewEntry()
	cm.mouseDeltaFactor.SetText(strconv.Itoa(cm.config.Motion.MouseDeltaFactor))

	cm.resetThreshold = widget.NewEntry()
	cm.resetThreshold.SetText(strconv.Itoa(int(cm.config.Gesture.ResetThreshold / time.Millisecond)))

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
func (cm *ConfigManager) GetConfig() *config.Config {
	return cm.config
}

// SetOnSave は設定保存時のコールバックを設定する
func (cm *ConfigManager) SetOnSave(callback func(*config.Config)) {
	cm.onSave = callback
}
