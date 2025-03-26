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

	// 初期選択を設定
	if len(keyboards) > 0 {
		dm.selectedKeyboard = &keyboards[0]
	}
	if len(mice) > 0 {
		dm.selectedMouse = &mice[0]
	}

	dm.statusLabel.SetText(
		"検出されたデバイス: " +
			"キーボード " + strconv.Itoa(len(keyboards)) + "台, " +
			"マウス " + strconv.Itoa(len(mice)) + "台")

	// 表示を更新
	if dm.deviceListWidget != nil {
		dm.deviceListWidget.Refresh()
	}

	if dm.keyboardSelection != nil && dm.mouseSelection != nil {
		// キーボードとマウスの選択肢を更新
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
	})

	dm.mouseSelection = widget.NewSelect(mouseNames, func(selected string) {
		for _, m := range mice {
			if m.Name == selected {
				dm.selectedMouse = &m
				break
			}
		}
	})

	// 初期選択を設定
	if len(keyboards) > 0 {
		dm.keyboardSelection.SetSelected(keyboards[0].Name)
		dm.selectedKeyboard = &keyboards[0]
	}
	if len(mice) > 0 {
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
	if len(keyboards) > 0 {
		dm.keyboardSelection.SetSelected(keyboards[0].Name)
		dm.selectedKeyboard = &keyboards[0]
	} else {
		dm.keyboardSelection.SetSelected("")
		dm.selectedKeyboard = nil
	}

	if len(mice) > 0 {
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
