//go:build darwin
// +build darwin

package features

import (
	"fmt"
	"log"
	"sync"
	"time"
)

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation -framework Foundation

#import <IOKit/IOKitLib.h>
#import <IOKit/hid/IOHIDManager.h>
#import <IOKit/hid/IOHIDDevice.h>
#import <CoreFoundation/CoreFoundation.h>
#import <Foundation/Foundation.h>

// デバイス情報構造体
typedef struct {
    char name[256];
    char path[256];
    int deviceType;  // 0: keyboard, 1: mouse
} DeviceInfo;

// コールバック用のコンテキスト
typedef struct {
    DeviceInfo* devices;
    int count;
    int capacity;
} DeviceContext;

// HIDデバイスから情報を取得
void getDeviceInfo(IOHIDDeviceRef device, DeviceInfo* info) {
    // デバイス名を取得
    CFStringRef product = IOHIDDeviceGetProperty(device, CFSTR(kIOHIDProductKey));
    if (product) {
        CFStringGetCString(product, info->name, sizeof(info->name), kCFStringEncodingUTF8);
    } else {
        strcpy(info->name, "Unknown Device");
    }
    
    // デバイスパスを取得（シンプルにデバイス名を使用）
    // 実際のパスはIORegistryから取得可能だが、macOSではあまり使用されない
    strcpy(info->path, info->name);
    
    // デバイスタイプを判定
    CFNumberRef usagePage = IOHIDDeviceGetProperty(device, CFSTR(kIOHIDPrimaryUsagePageKey));
    CFNumberRef usage = IOHIDDeviceGetProperty(device, CFSTR(kIOHIDPrimaryUsageKey));
    
    if (usagePage && usage) {
        int pageValue, usageValue;
        CFNumberGetValue(usagePage, kCFNumberIntType, &pageValue);
        CFNumberGetValue(usage, kCFNumberIntType, &usageValue);
        
        if (pageValue == kHIDPage_GenericDesktop) {
            if (usageValue == kHIDUsage_GD_Keyboard) {
                info->deviceType = 0; // keyboard
            } else if (usageValue == kHIDUsage_GD_Mouse) {
                info->deviceType = 1; // mouse
            }
        }
    }
}

// HIDデバイスを列挙するコールバック
void deviceEnumerationCallback(void* context, IOReturn result, void* sender, IOHIDDeviceRef device) {
    DeviceContext* ctx = (DeviceContext*)context;
    if (ctx->count < ctx->capacity) {
        getDeviceInfo(device, &ctx->devices[ctx->count]);
        ctx->count++;
    }
}

// デバイスをスキャン
int scanHIDDevices(DeviceInfo* devices, int maxDevices) {
    // HIDマネージャーを作成
    IOHIDManagerRef manager = IOHIDManagerCreate(kCFAllocatorDefault, kIOHIDOptionsTypeNone);
    if (!manager) {
        return 0;
    }
    
    // デバイスマッチング辞書を設定（すべてのHIDデバイス）
    IOHIDManagerSetDeviceMatching(manager, NULL);
    
    // デバイスを列挙
    CFSetRef deviceSet = IOHIDManagerCopyDevices(manager);
    if (!deviceSet) {
        CFRelease(manager);
        return 0;
    }
    
    int count = 0;
    CFIndex deviceCount = CFSetGetCount(deviceSet);
    if (deviceCount > 0 && deviceCount <= maxDevices) {
        IOHIDDeviceRef* deviceArray = malloc(sizeof(IOHIDDeviceRef) * deviceCount);
        CFSetGetValues(deviceSet, (const void**)deviceArray);
        
        for (CFIndex i = 0; i < deviceCount; i++) {
            getDeviceInfo(deviceArray[i], &devices[count]);
            count++;
        }
        
        free(deviceArray);
    }
    
    CFRelease(deviceSet);
    CFRelease(manager);
    
    return count;
}
*/
import "C"

// DeviceMonitor はデバイスの接続状態を監視する構造体
type DeviceMonitor struct {
	callbacks []DeviceCallback
	devices   map[string]*Device
	mutex     sync.RWMutex
	stopChan  chan struct{}
	isRunning bool
}

// グローバルなDeviceMonitorインスタンス
var (
	globalDeviceMonitor *DeviceMonitor
	deviceMonitorOnce   sync.Once
	deviceMonitorMutex  sync.Mutex
)

// ScanDevices は現在接続されているHIDデバイスをスキャンします
func ScanDevices() ([]Device, error) {
	// 最大100デバイスまでサポート
	maxDevices := 100
	cDevices := make([]C.DeviceInfo, maxDevices)
	
	count := int(C.scanHIDDevices(&cDevices[0], C.int(maxDevices)))
	
	devices := make([]Device, 0, count)
	for i := 0; i < count; i++ {
		deviceInfo := cDevices[i]
		deviceType := DeviceTypeKeyboard
		if deviceInfo.deviceType == 1 {
			deviceType = DeviceTypeMouse
		}
		
		// 関連するデバイスのみを追加（キーボードまたはマウス）
		if deviceInfo.deviceType == 0 || deviceInfo.deviceType == 1 {
			devices = append(devices, Device{
				Name: C.GoString(&deviceInfo.name[0]),
				Path: C.GoString(&deviceInfo.path[0]),
				Type: deviceType,
			})
		}
	}
	
	return devices, nil
}

// GetDevices は現在接続されているデバイスを取得します
func GetDevices() ([]Device, error) {
	// デバイスモニターが利用可能かチェック
	deviceMonitorMutex.Lock()
	monitor := globalDeviceMonitor
	deviceMonitorMutex.Unlock()
	
	if monitor != nil {
		devices := monitor.GetConnectedDevices()
		if len(devices) > 0 {
			return devices, nil
		}
	}
	
	// モニターがない場合は直接スキャン
	return ScanDevices()
}

// RescanDevices は強制的にデバイスを再スキャンします
func RescanDevices() ([]Device, error) {
	devices, err := ScanDevices()
	if err != nil {
		return nil, err
	}
	
	// モニターが初期化済みならデバイスリストを更新
	deviceMonitorMutex.Lock()
	monitor := globalDeviceMonitor
	deviceMonitorMutex.Unlock()
	
	if monitor != nil {
		monitor.updateDeviceList(devices)
	}
	
	return devices, nil
}

// NewDeviceMonitor は新しいDeviceMonitorを作成します
func NewDeviceMonitor() (*DeviceMonitor, error) {
	return &DeviceMonitor{
		callbacks: make([]DeviceCallback, 0),
		devices:   make(map[string]*Device),
		stopChan:  make(chan struct{}),
	}, nil
}

// Start はデバイスの監視を開始します
func (dm *DeviceMonitor) Start() error {
	if dm.isRunning {
		return nil
	}
	
	log.Println("デバイスモニターを開始します")
	dm.isRunning = true
	
	// 初期デバイス一覧を取得
	devices, err := ScanDevices()
	if err != nil {
		log.Printf("初期デバイス一覧の取得に失敗しました: %v", err)
	} else {
		log.Printf("初期デバイス検出: %d 個のデバイスを検出", len(devices))
		dm.updateDeviceList(devices)
	}
	
	// macOSではIOKitのコールバックを使用してデバイスの接続/切断を監視できるが、
	// 簡単のため定期的なポーリングを使用
	go dm.runPolling()
	
	return nil
}

// Stop はデバイスの監視を停止します
func (dm *DeviceMonitor) Stop() {
	if !dm.isRunning {
		return
	}
	
	log.Println("デバイスモニターを停止します")
	close(dm.stopChan)
	dm.isRunning = false
}

// RegisterCallback はデバイスイベントのコールバック関数を登録します
func (dm *DeviceMonitor) RegisterCallback(callback DeviceCallback) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	
	dm.callbacks = append(dm.callbacks, callback)
}

// runPolling は定期的にデバイスをスキャンします
func (dm *DeviceMonitor) runPolling() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-dm.stopChan:
			return
		case <-ticker.C:
			devices, err := ScanDevices()
			if err != nil {
				log.Printf("デバイススキャンに失敗しました: %v", err)
				continue
			}
			dm.updateDeviceList(devices)
		}
	}
}

// updateDeviceList は現在のデバイス一覧を更新し、変更があれば通知します
func (dm *DeviceMonitor) updateDeviceList(newDevices []Device) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	
	// 現在のデバイス名のセット
	currentNames := make(map[string]bool)
	for name := range dm.devices {
		currentNames[name] = true
	}
	
	// 新しいデバイス名のセット
	newNames := make(map[string]bool)
	for _, device := range newDevices {
		newNames[device.Name] = true
	}
	
	// 新規追加されたデバイスを検出
	for i := range newDevices {
		device := &newDevices[i]
		if _, exists := dm.devices[device.Name]; !exists {
			dm.devices[device.Name] = device
			log.Printf("新しいデバイスを追加: %s", device.Name)
			dm.notifyCallbacks(DeviceEvent{
				Type:   DeviceAdded,
				Device: device,
				Path:   device.Path,
			})
		}
	}
	
	// 削除されたデバイスを検出
	for name, device := range dm.devices {
		if !newNames[name] {
			log.Printf("デバイスを削除: %s", name)
			dm.notifyCallbacks(DeviceEvent{
				Type:   DeviceRemoved,
				Device: device,
				Path:   device.Path,
			})
			delete(dm.devices, name)
		}
	}
}

// notifyCallbacks は登録されているコールバックに通知します
func (dm *DeviceMonitor) notifyCallbacks(event DeviceEvent) {
	// コピーしてロックを解放した状態でコールバックを呼び出す
	var callbacks []DeviceCallback
	dm.mutex.RLock()
	callbacks = append(callbacks, dm.callbacks...)
	dm.mutex.RUnlock()
	
	for _, callback := range callbacks {
		go callback(event)
	}
}

// GetConnectedDevices は現在接続されているデバイスのリストを返します
func (dm *DeviceMonitor) GetConnectedDevices() []Device {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()
	
	devices := make([]Device, 0, len(dm.devices))
	for _, device := range dm.devices {
		devices = append(devices, *device)
	}
	
	return devices
}

// GetDeviceMonitor はグローバルDeviceMonitorインスタンスを返します
func GetDeviceMonitor() (*DeviceMonitor, error) {
	log.Printf("GetDeviceMonitor: デバイスモニターを取得します")
	
	// 既に初期化済みならそれを返す
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
		monitor, err := NewDeviceMonitor()
		if err != nil {
			log.Printf("デバイスモニターの初期化に失敗しました: %v", err)
			initErr = err
			return
		}
		
		// デバイスモニターを起動
		log.Printf("GetDeviceMonitor: デバイスモニターを開始します")
		if err := monitor.Start(); err != nil {
			log.Printf("デバイスモニターの起動に失敗しました: %v", err)
			initErr = err
			return
		}
		
		deviceMonitorMutex.Lock()
		globalDeviceMonitor = monitor
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
		return nil, fmt.Errorf("デバイスモニターの初期化に失敗しました")
	}
	
	return monitor, nil
}