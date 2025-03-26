package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// MacLikeTheme はmacOS風のテーマを提供します
type MacLikeTheme struct {
	fyne.Theme
}

// NewMacLikeTheme は新しいMacOS風テーマを返します
func NewMacLikeTheme() *MacLikeTheme {
	return &MacLikeTheme{
		Theme: theme.DefaultTheme(),
	}
}

// Color は指定されたテーマ名に対応する色を返します
func (m *MacLikeTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0, G: 122, B: 255, A: 255} // macOSの青色
	case theme.ColorNameButton:
		return color.NRGBA{R: 230, G: 230, B: 230, A: 255} // 薄いグレー
	case theme.ColorNameBackground:
		return color.NRGBA{R: 248, G: 248, B: 248, A: 255} // 明るい背景色
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 180, G: 180, B: 180, A: 128} // 薄い灰色
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 140, G: 140, B: 140, A: 255} // 中灰色
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 60} // 半透明の黒
	case theme.ColorNameForeground:
		return color.NRGBA{R: 25, G: 25, B: 25, A: 255} // ほぼ黒
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255} // 白
	default:
		return m.Theme.Color(name, variant)
	}
}

// Size はUI要素のサイズを返します
func (m *MacLikeTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 4 // やや小さめパディング
	case theme.SizeNameText:
		return 14 // 標準的なテキストサイズ
	case theme.SizeNameHeadingText:
		return 20 // 見出しテキストサイズ
	case theme.SizeNameSeparatorThickness:
		return 1 // 薄いセパレータ
	default:
		return m.Theme.Size(name)
	}
}

// Font は指定されたテキスト要素のフォントデータを返します
func (m *MacLikeTheme) Font(style fyne.TextStyle) fyne.Resource {
	return m.Theme.Font(style)
}

// Icon は指定されたアイコン名に対応するリソースを返します
func (m *MacLikeTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return m.Theme.Icon(name)
}
