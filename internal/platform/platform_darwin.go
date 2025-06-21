//go:build darwin
// +build darwin

package platform

import (
	"github.com/char5742/keyball-gestures/internal/config"
	"github.com/char5742/keyball-gestures/internal/platform/common"
	"github.com/char5742/keyball-gestures/internal/platform/darwin"
)

// NewPlatformFactory はmacOS用のPlatformFactoryを作成する
func NewPlatformFactory(cfg *config.Config) common.PlatformFactory {
	return darwin.NewFactory(cfg)
}