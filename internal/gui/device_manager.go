package gui

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/char5742/keyball-gestures/internal/features"
)

// DeviceType はデバイスタイプを表す型
type DeviceType int

// DeviceManager はデバイス管理のための構造体
type DeviceManager struct {
	app               *App
	devices           []features.Device
	selectedKeyboard  *features.Device
	selectedMouse     *features.Device
	deviceListWidget  *widget.List
	keyboardSelection *widget.Select
	mouseSelection    *widget.Select
	statusLabel       *widget.Label
	configManager     *ConfigManager
}

// NewDeviceManager は新しいデバイスマネージャを作成する
func NewDeviceManager(app *App) *DeviceManager {
	dm := &DeviceManager{
		app:         app,
		statusLabel: widget.NewLabel("デバイスをスキャン中..."),
	}

	// 初期デバイス一覧を取得
	dm.RefreshDevices()

	return dm
}

// SetConfigManager は設定マネージャを設定する
func (dm *DeviceManager) SetConfigManager(cm *ConfigManager) {
	dm.configManager = cm

	// 保存された設定があればそれをロード
	dm.LoadSavedDevicePreferences()
}

// LoadSavedDevicePreferences は保存されたデバイス設定をロードする
func (dm *DeviceManager) LoadSavedDevicePreferences() {
	if dm.configManager == nil {
		return
	}

	config := dm.configManager.GetConfig()
	preferredKeyboard := config.DevicePrefs.PreferredKeyboardDevice
	preferredMouse := config.DevicePrefs.PreferredMouseDevice

	// 保存されている設定と一致するデバイスがあれば選択する
	if preferredKeyboard != "" {
		for _, device := range dm.devices {
			if device.Type == features.DeviceTypeKeyboard && device.Name == preferredKeyboard {
				dm.selectedKeyboard = &device
				if dm.keyboardSelection != nil {
					dm.keyboardSelection.SetSelected(device.Name)
				}
				break
			}
		}
	}

	if preferredMouse != "" {
		for _, device := range dm.devices {
			if device.Type == features.DeviceTypeMouse && device.Name == preferredMouse {
				dm.selectedMouse = &device
				if dm.mouseSelection != nil {
					dm.mouseSelection.SetSelected(device.Name)
				}
				break
			}
		}
	}
}

// SaveDevicePreferences はデバイス設定を保存する
func (dm *DeviceManager) SaveDevicePreferences() {
	if dm.configManager == nil || dm.selectedKeyboard == nil || dm.selectedMouse == nil {
		return
	}

	config := dm.configManager.GetConfig()
	config.DevicePrefs.PreferredKeyboardDevice = dm.selectedKeyboard.Name
	config.DevicePrefs.PreferredMouseDevice = dm.selectedMouse.Name

	// 設定を保存
	err := dm.configManager.SaveConfig()
	if err != nil {
		// エラー通知
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "エラー",
			Content: "デバイス設定の保存に失敗しました: " + err.Error(),
		})
	} else {
		// 成功通知
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "成功",
			Content: "デバイス設定を保存しました",
		})
	}
}

// RefreshDevices はデバイス一覧を更新する
func (dm *DeviceManager) RefreshDevices() {
	// デバイス一覧を取得
	var err error
	dm.devices, err = features.GetDevices()
	if err != nil {
		dm.statusLabel.SetText("エラー: デバイス一覧の取得に失敗しました")
		return
	}

	// デバイスが見つからない場合
	if len(dm.devices) == 0 {
		dm.statusLabel.SetText("デバイスが見つかりませんでした")
		return
	}

	// 初期選択のために、キーボードとマウスのデバイスをフィルタリング
	var keyboards []features.Device
	var mice []features.Device

	for _, device := range dm.devices {
		if device.Type == features.DeviceTypeKeyboard {
			keyboards = append(keyboards, device)
		} else if device.Type == features.DeviceTypeMouse {
			mice = append(mice, device)
		}
	}

	dm.statusLabel.SetText(
		"検出されたデバイス: " +
			"キーボード " + strconv.Itoa(len(keyboards)) + "台, " +
			"マウス " + strconv.Itoa(len(mice)) + "台")

	// 表示を更新
	if dm.deviceListWidget != nil {
		dm.deviceListWidget.Refresh()
	}

	// 保存された設定があればそれを優先して読み込む
	prefKeyboardFound := false
	prefMouseFound := false

	if dm.configManager != nil {
		config := dm.configManager.GetConfig()
		preferredKeyboard := config.DevicePrefs.PreferredKeyboardDevice
		preferredMouse := config.DevicePrefs.PreferredMouseDevice

		// 保存されている設定と一致するデバイスがあれば選択する
		if preferredKeyboard != "" {
			for i, device := range keyboards {
				if device.Name == preferredKeyboard {
					dm.selectedKeyboard = &keyboards[i]
					prefKeyboardFound = true
					break
				}
			}
		}

		if preferredMouse != "" {
			for i, device := range mice {
				if device.Name == preferredMouse {
					dm.selectedMouse = &mice[i]
					prefMouseFound = true
					break
				}
			}
		}
	}

	// 設定で見つからなかった場合のみ、デフォルトで最初のデバイスを選択
	if !prefKeyboardFound && len(keyboards) > 0 {
		dm.selectedKeyboard = &keyboards[0]
	}

	if !prefMouseFound && len(mice) > 0 {
		dm.selectedMouse = &mice[0]
	}

	if dm.keyboardSelection != nil && dm.mouseSelection != nil {
		// キーボードとマウスの選択肢を更新し、選択状態を設定
		dm.updateSelections(keyboards, mice)
	}
}

// CreateDeviceConfigPanel はデバイス設定パネルを作成する
func (dm *DeviceManager) CreateDeviceConfigPanel() fyne.CanvasObject {
	// 初期デバイス一覧を取得
	var keyboards []features.Device
	var mice []features.Device
	var keyboardNames []string
	var mouseNames []string

	for _, device := range dm.devices {
		if device.Type == features.DeviceTypeKeyboard {
			keyboards = append(keyboards, device)
			keyboardNames = append(keyboardNames, device.Name)
		} else if device.Type == features.DeviceTypeMouse {
			mice = append(mice, device)
			mouseNames = append(mouseNames, device.Name)
		}
	}

	// デバイス選択のドロップダウン
	dm.keyboardSelection = widget.NewSelect(keyboardNames, func(selected string) {
		for _, kb := range keyboards {
			if kb.Name == selected {
				dm.selectedKeyboard = &kb
				break
			}
		}
		// 設定変更時に保存
		if dm.configManager != nil {
			dm.SaveDevicePreferences()
		}
	})

	dm.mouseSelection = widget.NewSelect(mouseNames, func(selected string) {
		for _, m := range mice {
			if m.Name == selected {
				dm.selectedMouse = &m
				break
			}
		}
		// 設定変更時に保存
		if dm.configManager != nil {
			dm.SaveDevicePreferences()
		}
	})

	// 初期選択を設定（優先デバイスがあればそれを選択）
	config := dm.configManager.GetConfig()
	preferredKeyboardFound := false
	preferredMouseFound := false

	if dm.configManager != nil {
		preferredKeyboard := config.DevicePrefs.PreferredKeyboardDevice
		preferredMouse := config.DevicePrefs.PreferredMouseDevice

		// キーボードの優先デバイスを設定
		if preferredKeyboard != "" {
			for i, kb := range keyboards {
				if kb.Name == preferredKeyboard {
					dm.keyboardSelection.SetSelected(kb.Name)
					dm.selectedKeyboard = &keyboards[i]
					preferredKeyboardFound = true
					break
				}
			}
		}

		// マウスの優先デバイスを設定
		if preferredMouse != "" {
			for i, m := range mice {
				if m.Name == preferredMouse {
					dm.mouseSelection.SetSelected(m.Name)
					dm.selectedMouse = &mice[i]
					preferredMouseFound = true
					break
				}
			}
		}
	}

	// 優先デバイスが見つからなかった場合は最初のデバイスを選択
	if !preferredKeyboardFound && len(keyboards) > 0 {
		dm.keyboardSelection.SetSelected(keyboards[0].Name)
		dm.selectedKeyboard = &keyboards[0]
	}

	if !preferredMouseFound && len(mice) > 0 {
		dm.mouseSelection.SetSelected(mice[0].Name)
		dm.selectedMouse = &mice[0]
	}

	// 更新ボタン
	refreshButton := widget.NewButton("デバイスリストを更新", func() {
		dm.RefreshDevices()
	})

	// レイアウト
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "キーボード", Widget: dm.keyboardSelection},
			{Text: "マウス", Widget: dm.mouseSelection},
		},
	}

	return container.NewVBox(
		dm.statusLabel,
		form,
		refreshButton,
	)
}

// 選択肢を更新する
func (dm *DeviceManager) updateSelections(keyboards []features.Device, mice []features.Device) {
	// キーボード選択肢を更新
	var keyboardNames []string
	for _, kb := range keyboards {
		keyboardNames = append(keyboardNames, kb.Name)
	}
	dm.keyboardSelection.Options = keyboardNames

	// マウス選択肢を更新
	var mouseNames []string
	for _, m := range mice {
		mouseNames = append(mouseNames, m.Name)
	}
	dm.mouseSelection.Options = mouseNames

	// 選択状態を更新
	if dm.selectedKeyboard != nil {
		dm.keyboardSelection.SetSelected(dm.selectedKeyboard.Name)
	} else if len(keyboards) > 0 {
		dm.keyboardSelection.SetSelected(keyboards[0].Name)
		dm.selectedKeyboard = &keyboards[0]
	} else {
		dm.keyboardSelection.SetSelected("")
		dm.selectedKeyboard = nil
	}

	if dm.selectedMouse != nil {
		dm.mouseSelection.SetSelected(dm.selectedMouse.Name)
	} else if len(mice) > 0 {
		dm.mouseSelection.SetSelected(mice[0].Name)
		dm.selectedMouse = &mice[0]
	} else {
		dm.mouseSelection.SetSelected("")
		dm.selectedMouse = nil
	}
}

// GetSelectedKeyboard は選択されたキーボードを返す
func (dm *DeviceManager) GetSelectedKeyboard() *features.Device {
	return dm.selectedKeyboard
}

// GetSelectedMouse は選択されたマウスを返す
func (dm *DeviceManager) GetSelectedMouse() *features.Device {
	return dm.selectedMouse
}
