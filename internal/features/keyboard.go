package features

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// キーボードからの入力を処理するインターフェース
type Keyboard interface {
	GetKey() (key int32)
	Close() error
}

type virtualKeyboard struct {
	*os.File
}

// 監視するデバイスのパスを指定してキーボードを作成する
func CreateKeyboard(path string) (Keyboard, error) {
	// デバイスを読み取り、非ブロッキングモードで開く
	f, err := os.OpenFile(path, syscall.O_RDONLY|syscall.O_NONBLOCK, 0660)
	if err != nil {
		return nil, fmt.Errorf("デバイスファイルを開くのに失敗しました: %w", err)
	}
	return &virtualKeyboard{f}, nil
}

func (v virtualKeyboard) GetKey() (key int32) {
	keys, err := getPressedKeys(v.File)
	if err != nil {
		return -1
	}
	if len(keys) > 0 {
		return int32(keys[0])
	}
	return -1
}

func getPressedKeys(file *os.File) ([]int, error) {
	const (
		keyMax    = 0x2ff
		eviocgkey = 0x80484518
	)

	keyBitsSize := (keyMax / 8) + 1
	keyBits := make([]byte, keyBitsSize)

	fd := file.Fd()

	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		fd,
		uintptr(eviocgkey),
		uintptr(unsafe.Pointer(&keyBits[0])),
	)
	if errno != 0 {
		return nil, errno
	}

	var pressed []int
	for keyCode := 0; keyCode < keyMax; keyCode++ {
		byteIndex := keyCode / 8
		bitIndex := keyCode % 8
		if (keyBits[byteIndex] & (1 << bitIndex)) != 0 {
			pressed = append(pressed, keyCode)
		}
	}
	return pressed, nil
}
