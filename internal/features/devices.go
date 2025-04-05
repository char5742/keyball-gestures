package features

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Device struct {
	Name string
	Path string
	Type DeviceType
}

// デバイスタイプを表す列挙型
type DeviceType int

const (
	DeviceTypeKeyboard DeviceType = iota
	DeviceTypeMouse
)

// DeviceEventType はデバイスイベントの種類を表す
type DeviceEventType int

const (
	DeviceAdded DeviceEventType = iota
	DeviceRemoved
	DeviceChanged
)

// DeviceEvent はデバイスの変更イベントを表す
type DeviceEvent struct {
	Type   DeviceEventType
	Device *Device
	Path   string
}

// DeviceCallback はデバイスイベント発生時に呼び出されるコールバック関数の型
type DeviceCallback func(event DeviceEvent)

// DeviceMonitor はデバイスの接続状態を監視する構造体
type DeviceMonitor struct {
	watcher         *fsnotify.Watcher
	callbacks       []DeviceCallback
	devices         map[string]*Device // パスをキーにしたデバイスマップ
	devicesByName   map[string]*Device // 名前をキーにしたデバイスマップ
	mutex           sync.RWMutex
	stopChan        chan struct{}
	autoRescanTimer *time.Timer
	pollingTicker   *time.Ticker
	isRunning       bool
}

// グローバルなDeviceMonitorインスタンス
var (
	globalDeviceMonitor *DeviceMonitor
	deviceMonitorOnce   sync.Once
	deviceMonitorMutex  sync.Mutex
)

// ScanDevices は基本的なデバイス検出を行い、現在接続されているデバイスリストを返します
// デバイスモニターを使用せず直接検出を行うため、キャッシュの影響を受けません
func ScanDevices() ([]Device, error) {
	entries, err := os.ReadDir("/dev/input/by-id")
	if err != nil {
		return nil, err
	}
	var devices []Device
	for _, entry := range entries {
		// eventが含まれない場合はスキップ
		if !strings.Contains(entry.Name(), "event") {
			continue
		}
		fullPath := "/dev/input/by-id/" + entry.Name()
		realPath, err := os.Readlink(fullPath)
		if err != nil {
			continue
		}

		// 絶対パスを構築
		absPath := ""
		if strings.HasPrefix(realPath, "/") {
			absPath = realPath
		} else {
			absPath = "/dev/input/" + filepath.Base(realPath)
		}

		if strings.Contains(entry.Name(), "kbd") {
			devices = append(devices, Device{Name: entry.Name(), Path: absPath, Type: DeviceTypeKeyboard})
		}
		if strings.Contains(entry.Name(), "mouse") {
			devices = append(devices, Device{Name: entry.Name(), Path: absPath, Type: DeviceTypeMouse})
		}
	}

	return devices, nil
}

// 基本的なデバイス検出用の関数 (デバイスモニターを使用しない)
// 内部向け実装 - 後方互換性のため残しておく
func scanDevices() ([]Device, error) {
	return ScanDevices()
}

// 現在接続されているデバイスを取得する関数
func GetDevices() ([]Device, error) {
	// 再帰を防ぐためのフラグ
	static := false
	if static {
		return scanDevices()
	}

	// デバイスモニターが利用可能かチェック (ただし初期化はしない)
	deviceMonitorMutex.Lock()
	monitor := globalDeviceMonitor
	deviceMonitorMutex.Unlock()

	if monitor != nil {
		// すでに初期化済みの場合はキャッシュを使用
		devices := monitor.GetConnectedDevices()
		if len(devices) > 0 {
			return devices, nil
		}
	}

	// モニターがない場合は直接スキャン
	return scanDevices()
}

// RescanDevices は現在接続されているデバイスを再スキャンします
func RescanDevices() ([]Device, error) {
	// モニターが初期化済みならそれを使用
	deviceMonitorMutex.Lock()
	monitor := globalDeviceMonitor
	deviceMonitorMutex.Unlock()

	if monitor != nil {
		// 強制的に再スキャン
		devices, err := scanDevices()
		if err != nil {
			return nil, err
		}

		// モニターのデバイスリストを更新
		monitor.updateDeviceList(devices)
		return devices, nil
	}

	// モニターがなければ単に検出
	return scanDevices()
}

// NewDeviceMonitor は新しいDeviceMonitorを作成する
func NewDeviceMonitor() (*DeviceMonitor, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &DeviceMonitor{
		watcher:   watcher,
		callbacks: make([]DeviceCallback, 0),
		devices:   make(map[string]*Device),
		stopChan:  make(chan struct{}),
	}, nil
}

// Start はデバイスの監視を開始する
func (dm *DeviceMonitor) Start() error {
	if dm.isRunning {
		return nil // すでに実行中
	}

	log.Println("デバイスモニターを開始します")
	dm.isRunning = true
	dm.devicesByName = make(map[string]*Device)

	// 監視対象のディレクトリを追加
	dirs := []string{
		"/dev/input",
		"/dev/input/by-id",
		"/dev/input/by-path",
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			if err := dm.watcher.Add(dir); err != nil {
				log.Printf("ディレクトリの監視に失敗しました: %s - %v", dir, err)
			} else {
				log.Printf("ディレクトリ監視を開始: %s", dir)
			}
		}
	}

	// 初期デバイス一覧を取得
	devices, err := ScanDevices()
	if err != nil {
		log.Printf("初期デバイス一覧の取得に失敗しました: %v", err)
	} else {
		log.Printf("初期デバイス検出: %d 個のデバイスを検出", len(devices))
		dm.updateDeviceList(devices)
	}

	// イベント監視ゴルーチンを起動
	go dm.watchEvents()

	// 定期的なリスキャンタイマーを設定（5秒ごと）
	dm.autoRescanTimer = time.AfterFunc(5*time.Second, dm.periodicRescan)

	// デバイスのポーリング監視を開始（2秒ごと）
	dm.pollingTicker = time.NewTicker(2 * time.Second)
	go dm.runPolling()

	return nil
}

// Stop はデバイスの監視を停止する
func (dm *DeviceMonitor) Stop() {
	if !dm.isRunning {
		return
	}

	log.Println("デバイスモニターを停止します")

	// 停止シグナルを送信
	close(dm.stopChan)

	// タイマーを停止
	if dm.autoRescanTimer != nil {
		dm.autoRescanTimer.Stop()
	}

	// ポーリングタイマーを停止
	if dm.pollingTicker != nil {
		dm.pollingTicker.Stop()
	}

	// ウォッチャーを閉じる
	dm.watcher.Close()

	dm.isRunning = false
}

// RegisterCallback はデバイスイベントのコールバック関数を登録する
func (dm *DeviceMonitor) RegisterCallback(callback DeviceCallback) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	dm.callbacks = append(dm.callbacks, callback)
}

// periodicRescan は定期的にデバイス一覧を再スキャンする
func (dm *DeviceMonitor) periodicRescan() {
	// デバイス一覧を再スキャン
	dm.RescanDevices()

	// タイマーを再設定（5秒後に再度実行）
	if dm.isRunning {
		dm.autoRescanTimer.Reset(5 * time.Second)
	}
}

// RescanDevices はデバイス一覧を強制的に再スキャンする
func (dm *DeviceMonitor) RescanDevices() {
	devices, err := RescanDevices()
	if err != nil {
		log.Printf("デバイス再スキャンに失敗しました: %v", err)
		return
	}

	dm.updateDeviceList(devices)
}

// runPolling はデバイスの存在を定期的に確認する
func (dm *DeviceMonitor) runPolling() {
	log.Println("デバイスポーリング監視を開始します")

	for {
		select {
		case <-dm.stopChan:
			log.Println("デバイスポーリング監視を停止します")
			return
		case <-dm.pollingTicker.C:
			// デバイスの存在チェックと新規デバイス検出を実行
			dm.checkDeviceExistence()
		}
	}
}

// checkDeviceExistence は現在登録されているデバイスの存在確認と新規デバイスの検出を行う
func (dm *DeviceMonitor) checkDeviceExistence() {
	// デバイス一覧を取得
	newDevices, err := ScanDevices()
	if err != nil {
		log.Printf("デバイススキャンに失敗しました: %v", err)
		return
	}

	// デバイス名をキーにしたマップを作成
	newDeviceMap := make(map[string]Device)
	for _, dev := range newDevices {
		newDeviceMap[dev.Name] = dev
	}

	// 現在のデバイス一覧を取得
	dm.mutex.RLock()
	currentDevices := make(map[string]bool)
	nameToPaths := make(map[string]string)

	for path, device := range dm.devices {
		currentDevices[path] = true
		nameToPaths[device.Name] = path
	}
	dm.mutex.RUnlock()

	// 削除されたデバイスを検出
	var removedPaths []string
	for path := range currentDevices {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			removedPaths = append(removedPaths, path)
		}
	}

	// 新しく見つかったデバイスを検出
	var newDeviceList []Device
	for _, dev := range newDevices {
		if _, exists := nameToPaths[dev.Name]; !exists {
			newDeviceList = append(newDeviceList, dev)
		}
	}

	// 削除されたデバイスがあれば通知
	if len(removedPaths) > 0 {
		log.Printf("ポーリングで削除されたデバイスを検出: %d 個", len(removedPaths))
		for _, path := range removedPaths {
			dm.mutex.Lock()
			if device, exists := dm.devices[path]; exists {
				log.Printf("デバイスが削除されました: %s (%s)", device.Name, path)
				dm.notifyCallbacks(DeviceEvent{
					Type:   DeviceRemoved,
					Device: device,
					Path:   path,
				})
				delete(dm.devices, path)
				delete(dm.devicesByName, device.Name)
			}
			dm.mutex.Unlock()
		}
	}

	// 新しいデバイスがあれば追加
	if len(newDeviceList) > 0 {
		log.Printf("ポーリングで新しいデバイスを検出: %d 個", len(newDeviceList))
		dm.updateDeviceList(newDeviceList)
	}

	// 既存のデバイスでパスが変更されたものがないか確認
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	for name, newDevice := range newDeviceMap {
		if oldDevice, exists := dm.devicesByName[name]; exists {
			// パスが変更された場合
			if oldDevice.Path != newDevice.Path {
				log.Printf("デバイスパスが変更されました: %s: %s → %s",
					name, oldDevice.Path, newDevice.Path)

				// 古いパスエントリを削除
				delete(dm.devices, oldDevice.Path)

				// 新しいパスで更新
				deviceCopy := newDevice
				dm.devices[newDevice.Path] = &deviceCopy
				dm.devicesByName[name] = &deviceCopy

				// パス変更を通知
				dm.notifyCallbacks(DeviceEvent{
					Type:   DeviceChanged,
					Device: &deviceCopy,
					Path:   newDevice.Path,
				})
			}
		}
	}
}

// updateDeviceList は現在のデバイス一覧を更新し、変更があれば通知する
func (dm *DeviceMonitor) updateDeviceList(newDevices []Device) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	// デバイスマップがまだ初期化されていない場合は初期化
	if dm.devices == nil {
		dm.devices = make(map[string]*Device)
	}

	if dm.devicesByName == nil {
		dm.devicesByName = make(map[string]*Device)
	}

	// 現在のデバイスマップをコピー
	currentDevices := make(map[string]bool)
	for path := range dm.devices {
		currentDevices[path] = true
	}

	// 新しいデバイス情報を整理
	deviceNames := make(map[string]bool)

	// 新しいデバイスを確認
	for i := range newDevices {
		device := &newDevices[i]
		path := device.Path
		deviceNames[device.Name] = true

		// 新規デバイスの場合
		if _, exists := dm.devices[path]; !exists {
			dm.devices[path] = device
			dm.devicesByName[device.Name] = device

			log.Printf("新しいデバイスを追加: %s (%s)", device.Name, path)
			dm.notifyCallbacks(DeviceEvent{
				Type:   DeviceAdded,
				Device: device,
				Path:   path,
			})
		} else {
			// 既存のデバイス（変更があるかチェック）
			if dm.devices[path].Name != device.Name {
				log.Printf("デバイス情報が変更: %s → %s (%s)",
					dm.devices[path].Name, device.Name, path)

				dm.devices[path] = device
				dm.devicesByName[device.Name] = device

				dm.notifyCallbacks(DeviceEvent{
					Type:   DeviceChanged,
					Device: device,
					Path:   path,
				})
			}
			// 既知のデバイスとしてマーク
			delete(currentDevices, path)
		}
	}

	// 削除されたデバイスを確認
	for path := range currentDevices {
		device := dm.devices[path]
		// デバイス名が別のパスで再登録されていないことを確認
		if !deviceNames[device.Name] {
			log.Printf("デバイスを削除: %s (%s)", device.Name, path)
			dm.notifyCallbacks(DeviceEvent{
				Type:   DeviceRemoved,
				Device: device,
				Path:   path,
			})
			delete(dm.devices, path)
			delete(dm.devicesByName, device.Name)
		}
	}
}

// notifyCallbacks は登録されているすべてのコールバックに通知する
func (dm *DeviceMonitor) notifyCallbacks(event DeviceEvent) {
	// イベント情報をログに出力
	if event.Type == DeviceRemoved {
		log.Printf("デバイス切断イベント: タイプ=%v, パス=%s, 名前=%s",
			event.Type, event.Path, event.Device.Name)
	} else if event.Type == DeviceAdded {
		log.Printf("デバイス接続イベント: タイプ=%v, パス=%s, 名前=%s",
			event.Type, event.Path, event.Device.Name)
	}

	// コピーしてロックを解放した状態でコールバックを呼び出す
	var callbacks []DeviceCallback
	dm.mutex.RLock()
	callbacks = append(callbacks, dm.callbacks...)
	dm.mutex.RUnlock()

	for i, callback := range callbacks {
		log.Printf("コールバック #%d を実行します", i)
		go func(cb DeviceCallback, i int, ev DeviceEvent) {
			log.Printf("コールバック #%d を実行中...", i)
			cb(ev)
			log.Printf("コールバック #%d 完了", i)
		}(callback, i, event)
	}
}

// watchEvents はfsnotifyのイベントを監視する
func (dm *DeviceMonitor) watchEvents() {
	log.Println("ファイルシステムイベント監視を開始します")

	// デバイスのデバッグ情報を表示
	debugDevices := func() {
		devicePaths := make([]string, 0, len(dm.devices))
		dm.mutex.RLock()
		defer dm.mutex.RUnlock()

		for path := range dm.devices {
			devicePaths = append(devicePaths, path)
		}

		// パスを並べ替えて表示
		sort.Strings(devicePaths)
		log.Println("現在監視中のデバイス一覧:")
		for _, path := range devicePaths {
			dev := dm.devices[path]
			log.Printf("  - パス: %s, 名前: %s, タイプ: %v", path, dev.Name, dev.Type)
		}
	}

	// 初回デバッグ情報
	debugDevices()

	// 一時的なファイルシステムイベントを収集してバッチ処理するためのしくみ
	eventDebounceTime := 500 * time.Millisecond
	eventTimer := time.NewTimer(eventDebounceTime)
	eventTimer.Stop() // 初期状態では停止
	pendingRescan := false

	for {
		select {
		case <-dm.stopChan:
			log.Println("ファイルシステムイベント監視を停止します")
			return

		case <-eventTimer.C:
			if pendingRescan {
				log.Println("ファイルシステムイベントをバッチ処理します")
				pendingRescan = false
				dm.RescanDevices()
			}

		case event, ok := <-dm.watcher.Events:
			if !ok {
				log.Println("イベントチャネルが閉じられました")
				return
			}

			// 特定のデバイス関連イベントのみログ出力
			isDeviceEvent := strings.Contains(event.Name, "/dev/input")
			if isDeviceEvent {
				log.Printf("ファイルシステムイベント: %s %s", event.Op.String(), event.Name)
			}

			// デバイスに関連するイベントのみ処理
			if isDeviceEvent && (event.Op&fsnotify.Create == fsnotify.Create ||
				event.Op&fsnotify.Remove == fsnotify.Remove ||
				event.Op&fsnotify.Write == fsnotify.Write) {

				// タイマーをリセットして複数のイベントをバッチ処理
				if !pendingRescan {
					pendingRescan = true
					eventTimer.Reset(eventDebounceTime)
				}
			}

		case err, ok := <-dm.watcher.Errors:
			if !ok {
				log.Println("エラーチャネルが閉じられました")
				return
			}
			log.Printf("ファイルシステム監視エラー: %v", err)
		}
	}
}

// GetConnectedDevices は現在接続されているデバイスのスナップショットを返す
func (dm *DeviceMonitor) GetConnectedDevices() []Device {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	devices := make([]Device, 0, len(dm.devices))
	for _, device := range dm.devices {
		devices = append(devices, *device)
	}

	return devices
}

// GetDeviceMonitor はグローバルDeviceMonitorインスタンスを返す（必要に応じて作成）
func GetDeviceMonitor() (*DeviceMonitor, error) {
	log.Printf("GetDeviceMonitor: デバイスモニターを取得します")

	// 既に初期化済みならそれを返す（高速パス）
	deviceMonitorMutex.Lock()
	if globalDeviceMonitor != nil {
		monitor := globalDeviceMonitor
		deviceMonitorMutex.Unlock()
		return monitor, nil
	}
	deviceMonitorMutex.Unlock()

	// 初期化処理
	var initErr error
	deviceMonitorOnce.Do(func() {
		log.Printf("GetDeviceMonitor: デバイスモニターを初期化します")
		deviceMonitor, err := NewDeviceMonitor()
		if err != nil {
			log.Printf("デバイスモニターの初期化に失敗しました: %v", err)
			initErr = err
			return
		}

		// デバイスモニターを起動
		log.Printf("GetDeviceMonitor: デバイスモニターを開始します")
		if err := deviceMonitor.Start(); err != nil {
			log.Printf("デバイスモニターの起動に失敗しました: %v", err)
			initErr = err
			return
		}

		deviceMonitorMutex.Lock()
		globalDeviceMonitor = deviceMonitor
		deviceMonitorMutex.Unlock()
		log.Printf("GetDeviceMonitor: デバイスモニターの初期化が完了しました")
	})

	if initErr != nil {
		return nil, initErr
	}

	deviceMonitorMutex.Lock()
	monitor := globalDeviceMonitor
	deviceMonitorMutex.Unlock()

	if monitor == nil {
		log.Printf("GetDeviceMonitor: デバイスモニターが初期化されていません")
		return nil, fmt.Errorf("デバイスモニターの初期化に失敗しました")
	}

	return monitor, nil
}
