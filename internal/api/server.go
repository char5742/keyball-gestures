package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/char5742/keyball-gestures/internal/config"
)

// Server はAPIサーバーを表す構造体
type Server struct {
	server *http.Server
	cfg    *config.Config
	mutex  sync.RWMutex
	port   int
}

// NewServer は新しいAPIサーバーを作成する
func NewServer(cfg *config.Config, port int) *Server {
	return &Server{
		cfg:  cfg,
		port: port,
	}
}

// Start はAPIサーバーを開始する
func (s *Server) Start() error {
	// ルーターの設定
	router := http.NewServeMux()

	// APIエンドポイントの設定
	s.setupRoutes(router)

	// HTTPサーバーの設定
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: router,
	}

	// サーバーの起動
	log.Printf("APIサーバーを開始します: http://localhost:%d", s.port)
	return s.server.ListenAndServe()
}

// Stop はAPIサーバーを停止する
func (s *Server) Stop() error {
	if s.server != nil {
		log.Println("APIサーバーを停止します...")
		return s.server.Shutdown(context.Background())
	}
	return nil
}

// GetConfig は現在の設定を返す
func (s *Server) GetConfig() *config.Config {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.cfg
}

// UpdateConfig は設定を更新する
func (s *Server) UpdateConfig(cfg *config.Config) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.cfg = cfg
}

// writeJSON はJSONレスポンスを書き込む
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("JSONエンコードエラー: %v", err)
		}
	}
}

// writeError はエラーレスポンスを書き込む
func writeError(w http.ResponseWriter, status int, message string) {
	response := map[string]string{"error": message}
	writeJSON(w, status, response)
}
