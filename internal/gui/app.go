package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

// App はKeyball GesturesのGUIアプリケーション構造体
type App struct {
	fyneApp        fyne.App
	window         fyne.Window
	deviceManager  *DeviceManager
	configManager  *ConfigManager
	serviceManager *ServiceManager
}

// NewApp は新しいGUIアプリケーションを作成する
func NewApp() *App {
	a := &App{
		fyneApp: app.New(),
	}
	a.window = a.fyneApp.NewWindow("Keyball Gestures")
	a.window.Resize(fyne.NewSize(800, 600))

	// macOS風のテーマを設定
	fyne.CurrentApp().Settings().SetTheme(NewMacLikeTheme())

	// 各マネージャーを初期化
	a.deviceManager = NewDeviceManager(a)
	a.configManager = NewConfigManager(a)
	a.serviceManager = NewServiceManager(a, a.deviceManager, a.configManager)

	// DeviceManagerにConfigManagerを設定
	a.deviceManager.SetConfigManager(a.configManager)

	return a
}

// Run はGUIアプリケーションを実行する
func (a *App) Run() {
	// タブコンテナを作成
	tabs := container.NewAppTabs(
		container.NewTabItem("デバイス設定", a.makeDeviceTab()),
		container.NewTabItem("ジェスチャー設定", a.makeGestureTab()),
		container.NewTabItem("操作モニター", a.makeMonitorTab()),
	)

	// メインコンテンツとしてタブを設定
	a.window.SetContent(tabs)

	// ウィンドウを表示
	a.window.ShowAndRun()
}

// makeDeviceTab はデバイス設定タブを作成する
func (a *App) makeDeviceTab() fyne.CanvasObject {
	return a.deviceManager.CreateDeviceConfigPanel()
}

// makeGestureTab はジェスチャー設定タブを作成する
func (a *App) makeGestureTab() fyne.CanvasObject {
	return a.configManager.CreateConfigPanel()
}

// makeMonitorTab は操作モニタータブを作成する
func (a *App) makeMonitorTab() fyne.CanvasObject {
	return a.serviceManager.CreateServicePanel()
}
