//go:build linux
// +build linux

package platform

import (
	"github.com/char5742/keyball-gestures/internal/config"
	"github.com/char5742/keyball-gestures/internal/platform/common"
	"github.com/char5742/keyball-gestures/internal/platform/linux"
)

// NewPlatformFactory はLinux用のPlatformFactoryを作成する
func NewPlatformFactory(cfg *config.Config) common.PlatformFactory {
	return linux.NewFactory(cfg)
}