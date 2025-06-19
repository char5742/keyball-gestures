//go:build darwin
// +build darwin

package utils

import (
	"fmt"
	"os"
)

// IOCtl macOSではioctlは使用しないため、ダミー実装
func IOCtl(deviceFile *os.File, cmd, ptr uintptr) error {
	return fmt.Errorf("ioctl is not supported on macOS")
}