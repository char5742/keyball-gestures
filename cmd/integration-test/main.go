package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/char5742/keyball-gestures/internal/features"
)

// イベント記録の構造体
type EventRecord struct {
	StartTime    int64         `json:"startTime"`
	Duration     int           `json:"duration"`
	Events       []EventInfo   `json:"events"`
	EndTime      int64         `json:"endTime"`
	TotalEvents  int           `json:"totalEvents"`
}

type EventInfo struct {
	Timestamp   float64 `json:"timestamp"`
	Type        int     `json:"type"`
	TypeName    string  `json:"typeName"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	DeltaX      float64 `json:"deltaX"`
	DeltaY      float64 `json:"deltaY"`
	IsGesture   bool    `json:"isGesture"`
	FingerCount int     `json:"fingerCount"`
}

// テスト結果
type TestResult struct {
	TestName    string
	Passed      bool
	Message     string
	EventCount  int
	Events      []EventInfo
}

// 統合テストランナー
type IntegrationTestRunner struct {
	touchpad     features.TouchPad
	monitorPath  string
	outputDir    string
}

func NewIntegrationTestRunner() (*IntegrationTestRunner, error) {
	// 出力ディレクトリを作成
	outputDir := "test-results"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("出力ディレクトリの作成に失敗: %v", err)
	}

	// タッチパッドを初期化
	cfg := features.TouchPadConfig{
		Name:                  "integration-test-touchpad",
		MinX:                  0,
		MaxX:                  32767,
		MinY:                  0,
		MaxY:                  32767,
		MouseDeltaFactor:      1.0,
		MotionSmoothingFactor: 0.5,
		MotionWarmUpCount:     3,
	}

	touchpad, err := features.CreateTouchPad(cfg)
	if err != nil {
		return nil, fmt.Errorf("タッチパッドの初期化に失敗: %v", err)
	}

	// 実行ファイルのディレクトリを取得
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("実行ファイルパスの取得に失敗: %v", err)
	}
	execDir := filepath.Dir(execPath)
	
	return &IntegrationTestRunner{
		touchpad:    touchpad,
		monitorPath: filepath.Join(execDir, "event_monitor"),
		outputDir:   outputDir,
	}, nil
}

// イベント監視プログラムをビルド
func (r *IntegrationTestRunner) buildMonitor() error {
	// 既にビルド済みかチェック
	if _, err := os.Stat(r.monitorPath); err == nil {
		log.Println("イベント監視プログラムは既にビルド済みです")
		return nil
	}
	
	cmd := exec.Command("clang",
		"c/event_monitor.c",
		"-framework", "ApplicationServices",
		"-framework", "CoreFoundation",
		"-o", "event_monitor")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("監視プログラムのビルドに失敗: %v\n%s", err, output)
	}
	
	log.Println("イベント監視プログラムをビルドしました")
	return nil
}

// イベント監視を開始
func (r *IntegrationTestRunner) startMonitor(outputFile string, duration int) (*exec.Cmd, error) {
	cmd := exec.Command(r.monitorPath, outputFile, fmt.Sprintf("%d", duration))
	
	// 標準出力と標準エラーをパイプ
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("監視プログラムの起動に失敗: %v", err)
	}
	
	// 監視プログラムが初期化されるまで少し待つ
	time.Sleep(500 * time.Millisecond)
	
	return cmd, nil
}

// イベント記録を読み込み
func (r *IntegrationTestRunner) loadEventRecord(filename string) (*EventRecord, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("イベント記録の読み込みに失敗: %v", err)
	}
	
	var record EventRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("JSONのパースに失敗: %v", err)
	}
	
	return &record, nil
}

// 2本指スクロールテスト
func (r *IntegrationTestRunner) test2FingerScroll() TestResult {
	testName := "2FingerScroll"
	outputFile := filepath.Join(r.outputDir, fmt.Sprintf("%s_%d.json", testName, time.Now().Unix()))
	
	log.Printf("\n=== %s テスト開始 ===\n", testName)
	
	// イベント監視を開始（10秒間）
	monitor, err := r.startMonitor(outputFile, 10)
	if err != nil {
		return TestResult{
			TestName: testName,
			Passed:   false,
			Message:  err.Error(),
		}
	}
	defer monitor.Process.Kill()
	
	// 2本指スクロールジェスチャーをシミュレート
	log.Println("2本指スクロールジェスチャーを実行します...")
	
	// タッチ開始（2本指）
	r.touchpad.MultiTouchDown(0, 1001, 500, 500)
	r.touchpad.MultiTouchDown(1, 1002, 600, 500)
	time.Sleep(100 * time.Millisecond)
	
	// 上方向にスクロール（Y座標を減少）
	for i := 0; i < 10; i++ {
		r.touchpad.MultiTouchMove(0, 500, int32(500-i*20))
		r.touchpad.MultiTouchMove(1, 600, int32(500-i*20))
		time.Sleep(50 * time.Millisecond)
	}
	
	// タッチ終了
	r.touchpad.MultiTouchUp(0)
	r.touchpad.MultiTouchUp(1)
	
	// 監視プログラムが終了するまで待つ
	time.Sleep(2 * time.Second)
	
	// SIGTERMを送って正常終了を促す
	if err := monitor.Process.Signal(os.Interrupt); err != nil {
		log.Printf("監視プログラムへのシグナル送信エラー: %v", err)
	}
	
	// 少し待ってから終了を確認
	time.Sleep(500 * time.Millisecond)
	
	// まだ終了していない場合は強制終了
	if monitor.Process != nil {
		monitor.Process.Kill()
	}
	
	// 終了を待つ
	err = monitor.Wait()
	if err != nil && err.Error() != "signal: interrupt" && err.Error() != "signal: terminated" {
		log.Printf("監視プログラムの終了エラー: %v", err)
	}
	
	// 結果を検証
	record, err := r.loadEventRecord(outputFile)
	if err != nil {
		return TestResult{
			TestName: testName,
			Passed:   false,
			Message:  err.Error(),
		}
	}
	
	// スクロールイベントが記録されているか確認
	scrollEventCount := 0
	var scrollEvents []EventInfo
	hasNonZeroDelta := false
	
	for _, event := range record.Events {
		if event.TypeName == "ScrollWheel" && event.IsGesture && event.FingerCount == 2 {
			scrollEventCount++
			scrollEvents = append(scrollEvents, event)
			
			// delta値が0でないイベントがあるかチェック
			if event.DeltaX != 0 || event.DeltaY != 0 {
				hasNonZeroDelta = true
			}
		}
	}
	
	// イベントが検出され、かつdelta値が正しく記録されている場合のみ成功
	if scrollEventCount > 0 && hasNonZeroDelta {
		return TestResult{
			TestName:   testName,
			Passed:     true,
			Message:    fmt.Sprintf("2本指スクロールイベントを%d個検出しました（delta値あり）", scrollEventCount),
			EventCount: scrollEventCount,
			Events:     scrollEvents,
		}
	}
	
	// delta値が全て0の場合は失敗
	if scrollEventCount > 0 && !hasNonZeroDelta {
		return TestResult{
			TestName:   testName,
			Passed:     false,
			Message:    fmt.Sprintf("2本指スクロールイベントは検出されましたが、delta値が全て0です"),
			EventCount: scrollEventCount,
			Events:     scrollEvents,
		}
	}
	
	return TestResult{
		TestName: testName,
		Passed:   false,
		Message:  "2本指スクロールイベントが検出されませんでした",
		Events:   record.Events,
	}
}

// 4本指スワイプテスト
func (r *IntegrationTestRunner) test4FingerSwipe() TestResult {
	testName := "4FingerSwipe"
	outputFile := filepath.Join(r.outputDir, fmt.Sprintf("%s_%d.json", testName, time.Now().Unix()))
	
	log.Printf("\n=== %s テスト開始 ===\n", testName)
	
	// イベント監視を開始（10秒間）
	monitor, err := r.startMonitor(outputFile, 10)
	if err != nil {
		return TestResult{
			TestName: testName,
			Passed:   false,
			Message:  err.Error(),
		}
	}
	defer monitor.Process.Kill()
	
	// 4本指スワイプジェスチャーをシミュレート
	log.Println("4本指スワイプジェスチャーを実行します...")
	
	// タッチ開始（4本指） - 素早く4本とも追加（50ms以内）
	r.touchpad.MultiTouchDown(0, 2001, 400, 500)
	time.Sleep(10 * time.Millisecond)
	r.touchpad.MultiTouchDown(1, 2002, 500, 500)
	time.Sleep(10 * time.Millisecond)
	r.touchpad.MultiTouchDown(2, 2003, 600, 500)
	time.Sleep(10 * time.Millisecond)
	r.touchpad.MultiTouchDown(3, 2004, 700, 500)
	time.Sleep(100 * time.Millisecond)  // ジェスチャーが認識されるまで少し待つ
	
	// 左方向にスワイプ（X座標を減少）
	for i := 0; i < 10; i++ {
		r.touchpad.MultiTouchMove(0, int32(400-i*30), 500)
		r.touchpad.MultiTouchMove(1, int32(500-i*30), 500)
		r.touchpad.MultiTouchMove(2, int32(600-i*30), 500)
		r.touchpad.MultiTouchMove(3, int32(700-i*30), 500)
		time.Sleep(50 * time.Millisecond)
	}
	
	// タッチ終了
	r.touchpad.MultiTouchUp(0)
	r.touchpad.MultiTouchUp(1)
	r.touchpad.MultiTouchUp(2)
	r.touchpad.MultiTouchUp(3)
	
	// 監視プログラムが終了するまで待つ
	time.Sleep(2 * time.Second)
	
	// SIGTERMを送って正常終了を促す
	if err := monitor.Process.Signal(os.Interrupt); err != nil {
		log.Printf("監視プログラムへのシグナル送信エラー: %v", err)
	}
	
	// 少し待ってから終了を確認
	time.Sleep(500 * time.Millisecond)
	
	// まだ終了していない場合は強制終了
	if monitor.Process != nil {
		monitor.Process.Kill()
	}
	
	// 終了を待つ
	err = monitor.Wait()
	if err != nil && err.Error() != "signal: interrupt" && err.Error() != "signal: terminated" {
		log.Printf("監視プログラムの終了エラー: %v", err)
	}
	
	// 結果を検証
	record, err := r.loadEventRecord(outputFile)
	if err != nil {
		return TestResult{
			TestName: testName,
			Passed:   false,
			Message:  err.Error(),
		}
	}
	
	// 4本指ジェスチャーイベントが記録されているか確認
	gestureEventCount := 0
	var gestureEvents []EventInfo
	hasNonZeroDelta := false
	
	for _, event := range record.Events {
		if (event.TypeName == "GestureBegin" || event.TypeName == "GestureChange" || event.TypeName == "GestureEnd") && 
		   event.IsGesture && event.FingerCount == 4 {
			gestureEventCount++
			gestureEvents = append(gestureEvents, event)
			
			// delta値が0でないイベントがあるかチェック
			if event.DeltaX != 0 || event.DeltaY != 0 {
				hasNonZeroDelta = true
			}
		}
	}
	
	// イベントが検出され、かつdelta値が正しく記録されている場合のみ成功
	if gestureEventCount > 0 && hasNonZeroDelta {
		return TestResult{
			TestName:   testName,
			Passed:     true,
			Message:    fmt.Sprintf("4本指ジェスチャーイベントを%d個検出しました（delta値あり）", gestureEventCount),
			EventCount: gestureEventCount,
			Events:     gestureEvents,
		}
	}
	
	// delta値が全て0の場合は失敗
	if gestureEventCount > 0 && !hasNonZeroDelta {
		return TestResult{
			TestName:   testName,
			Passed:     false,
			Message:    fmt.Sprintf("4本指ジェスチャーイベントは検出されましたが、delta値が全て0です"),
			EventCount: gestureEventCount,
			Events:     gestureEvents,
		}
	}
	
	return TestResult{
		TestName: testName,
		Passed:   false,
		Message:  "4本指ジェスチャーイベントが検出されませんでした",
		Events:   record.Events,
	}
}

// テスト結果を表示
func (r *IntegrationTestRunner) printResult(result TestResult) {
	status := "FAIL"
	if result.Passed {
		status = "PASS"
	}
	
	fmt.Printf("\n[%s] %s: %s\n", status, result.TestName, result.Message)
	
	if len(result.Events) > 0 && !result.Passed {
		fmt.Println("記録されたイベント:")
		for i, event := range result.Events {
			if i >= 5 {
				fmt.Printf("  ... 他 %d イベント\n", len(result.Events)-5)
				break
			}
			fmt.Printf("  - [%.2f] %s (finger:%d, dx:%.1f, dy:%.1f)\n",
				event.Timestamp, event.TypeName, event.FingerCount, event.DeltaX, event.DeltaY)
		}
	}
}

// クリーンアップ
func (r *IntegrationTestRunner) cleanup() {
	if r.touchpad != nil {
		r.touchpad.Close()
	}
}

func main() {
	log.Println("=== Keyball Gestures 統合テスト ===")
	
	// テストランナーを初期化
	runner, err := NewIntegrationTestRunner()
	if err != nil {
		log.Fatalf("テストランナーの初期化に失敗: %v", err)
	}
	defer runner.cleanup()
	
	// イベント監視プログラムをビルド（スキップ - 既にビルド済み）
	log.Println("イベント監視プログラムは既にビルド済みです")
	
	// 各テストを実行
	results := []TestResult{
		runner.test2FingerScroll(),
		runner.test4FingerSwipe(),
	}
	
	// 結果サマリー
	fmt.Println("\n=== テスト結果サマリー ===")
	passedCount := 0
	for _, result := range results {
		runner.printResult(result)
		if result.Passed {
			passedCount++
		}
	}
	
	fmt.Printf("\n合計: %d/%d テストが成功しました\n", passedCount, len(results))
	
	if passedCount < len(results) {
		os.Exit(1)
	}
}