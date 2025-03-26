package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/char5742/keyball-gestures/internal/config"
	"github.com/char5742/keyball-gestures/internal/features"
)

// ルートの設定
func (s *Server) setupRoutes(router *http.ServeMux) {
	// 設定関連のエンドポイント
	router.HandleFunc("GET /api/config", s.handleGetConfig)
	router.HandleFunc("PUT /api/config", s.handleUpdateConfig)
	router.HandleFunc("POST /api/config/save", s.handleSaveConfig)

	// デバイス関連のエンドポイント
	router.HandleFunc("GET /api/devices", s.handleGetDevices)
	router.HandleFunc("PUT /api/devices/preferred", s.handleSetPreferredDevices)

	// サービス関連のエンドポイント
	router.HandleFunc("POST /api/service/start", s.handleStartService)
	router.HandleFunc("POST /api/service/stop", s.handleStopService)
	router.HandleFunc("GET /api/service/status", s.handleServiceStatus)

	// ヘルスチェック用エンドポイント
	router.HandleFunc("GET /api/health", s.handleHealthCheck)
}

// 設定取得ハンドラ
func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.GetConfig())
}

// 設定更新ハンドラ
func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var newConfig config.Config

	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		writeError(w, http.StatusBadRequest, "設定の解析に失敗しました")
		return
	}

	s.UpdateConfig(&newConfig)
	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// 設定保存ハンドラ
func (s *Server) handleSaveConfig(w http.ResponseWriter, r *http.Request) {
	var saveRequest struct {
		Path string `json:"path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&saveRequest); err != nil {
		writeError(w, http.StatusBadRequest, "リクエストの解析に失敗しました")
		return
	}

	configPath := saveRequest.Path
	if configPath == "" {
		// デフォルトパスを使用
		userConfigDir, err := config.GetDefaultConfigDir()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "デフォルト設定ディレクトリの取得に失敗しました")
			return
		}
		configPath = filepath.Join(userConfigDir, "config.toml")
	}

	if err := config.SaveConfig(configPath, s.GetConfig()); err != nil {
		writeError(w, http.StatusInternalServerError, "設定の保存に失敗しました: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "success",
		"path":   configPath,
	})
}

// デバイス一覧取得ハンドラ
func (s *Server) handleGetDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := features.GetDevices()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "デバイス一覧の取得に失敗しました: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, devices)
}

// 優先デバイス設定ハンドラ
func (s *Server) handleSetPreferredDevices(w http.ResponseWriter, r *http.Request) {
	var request struct {
		KeyboardDevice string `json:"keyboard_device"`
		MouseDevice    string `json:"mouse_device"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "リクエストの解析に失敗しました")
		return
	}

	cfg := s.GetConfig()
	cfg.DevicePrefs.PreferredKeyboardDevice = request.KeyboardDevice
	cfg.DevicePrefs.PreferredMouseDevice = request.MouseDevice
	s.UpdateConfig(cfg)

	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// ジェスチャー認識サービス
var gestureService *GestureService

// サービス起動ハンドラ
func (s *Server) handleStartService(w http.ResponseWriter, r *http.Request) {
	if gestureService == nil {
		gestureService = NewGestureService(s.GetConfig())
	}

	if gestureService.IsRunning() {
		writeJSON(w, http.StatusOK, map[string]string{"status": "already_running"})
		return
	}

	if err := gestureService.Start(); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("サービスの起動に失敗しました: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

// サービス停止ハンドラ
func (s *Server) handleStopService(w http.ResponseWriter, r *http.Request) {
	if gestureService == nil || !gestureService.IsRunning() {
		writeJSON(w, http.StatusOK, map[string]string{"status": "not_running"})
		return
	}

	if err := gestureService.Stop(); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("サービスの停止に失敗しました: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

// サービス状態取得ハンドラ
func (s *Server) handleServiceStatus(w http.ResponseWriter, r *http.Request) {
	status := "stopped"
	if gestureService != nil && gestureService.IsRunning() {
		status = "running"
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": status})
}

// ヘルスチェックハンドラ
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
