#!/bin/bash

echo "=== 統合解析スクリプト ==="
echo ""
echo "1. イベントアナライザーを起動"
echo "2. 3秒待機"
echo "3. テストプログラムを実行"
echo "4. 結果を解析"
echo ""

# ディレクトリ移動
cd test/debugging

# アナライザーをバックグラウンドで起動
./event_analyzer > analyzer_output.txt 2>&1 &
ANALYZER_PID=$!

echo "アナライザーを起動しました (PID: $ANALYZER_PID)"
sleep 3

# テストプログラムを実行
echo "テストプログラムを実行中..."
cd ../..
go run test/manual/test_swipe_fixed.go > /dev/null 2>&1

# アナライザーが終了するまで待つ
echo "解析中..."
wait $ANALYZER_PID

# 結果を表示
echo ""
echo "=== 解析結果 ==="
cat test/debugging/analyzer_output.txt

# クリーンアップ
rm -f test/debugging/analyzer_output.txt