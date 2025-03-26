package gui

import (
	"log"
	"os"
)

// RunGUI はGUIアプリケーションを起動する
func RunGUI() {
	// ルート権限チェック
	if os.Geteuid() != 0 {
		log.Fatal("このアプリケーションはルート権限で実行する必要があります")
	}

	// アプリケーションの初期化
	app := NewApp()

	// アプリケーションの実行
	app.Run()
}
