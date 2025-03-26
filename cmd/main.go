package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/char5742/keyball-gestures/internal/api"
	"github.com/char5742/keyball-gestures/internal/config"
)

func main() {
	// コマンドライン引数の解析
	useApi := flag.Bool("api", false, "APIサーバーモードで起動します")
	configPath := flag.String("config", "", "設定ファイルのパス (指定しない場合はデフォルトパスを使用)")
	port := flag.Int("port", 8080, "APIサーバーのポート番号")
	flag.Parse()

	// デフォルト設定ファイルパスの設定
	defaultConfigPath := ""
	configDir, err := config.GetDefaultConfigDir()
	if err == nil {
		defaultConfigPath = filepath.Join(configDir, "config.toml")
	}

	// 設定ファイルパスの決定
	cfgPath := defaultConfigPath
	if *configPath != "" {
		cfgPath = *configPath
	}

	// 設定ファイルの読み込み
	var cfg *config.Config
	if cfgPath != "" {
		cfg, err = config.LoadConfig(cfgPath)
		if err != nil {
			fmt.Printf("設定ファイルの読み込みに失敗しました: %v\nデフォルト設定を使用します\n", err)
			cfg = config.DefaultConfig()
		} else {
			fmt.Printf("設定ファイルを読み込みました: %s\n", cfgPath)
		}
	} else {
		cfg = config.DefaultConfig()
	}

	// シグナルハンドラの設定
	handleSignals()

	// APIモードかCLIモードかを判断
	if *useApi {
		// APIモードで実行
		fmt.Printf("APIサーバーモードで起動します (ポート: %d)...\n", *port)
		runApiServer(cfg, *port)
	} else {
		// CLIモードで実行
		fmt.Println("CLIモードで起動します...")
		runCLI(cfg)
	}
}

// APIサーバーモードでの実行
func runApiServer(cfg *config.Config, port int) {
	// APIサーバーを作成
	server := api.NewServer(cfg, port)

	// サーバー起動
	if err := server.Start(); err != nil {
		log.Fatalf("APIサーバーの起動に失敗しました: %v", err)
	}
}

// CLIモードでの実行
func runCLI(cfg *config.Config) {
	// ジェスチャー認識サービスを作成
	service := api.NewGestureService(cfg)

	// サービス開始
	if err := service.Start(); err != nil {
		fmt.Printf("ジェスチャー認識サービスの起動に失敗しました: %v\n", err)
		os.Exit(1)
	}

	// シグナルが来るまで待機（終了処理はhandleSignals内で行われる）
	select {}
}

func handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("シャットダウンします...")
		os.Exit(0)
	}()
}
