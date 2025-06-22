#!/bin/bash

echo "=== イベント検証スクリプト ==="
echo "1. trackpad_dumpを起動"
echo "2. テストプログラムを実行"
echo "3. 出力を比較"

# trackpad_dumpをビルド
cd test/debugging
if [ ! -f trackpad_dump ]; then
    echo "trackpad_dumpをビルド中..."
    cc trackpad_dump.c -o trackpad_dump -framework ApplicationServices -framework CoreFoundation
fi

# trackpad_dumpを起動してログファイルに出力
echo "trackpad_dumpを起動..."
./trackpad_dump > ../../test/manual/actual_output.log 2>&1 &
DUMP_PID=$!

sleep 2

# テストプログラムを実行
echo "テストプログラムを実行中..."
cd ../..
go run test/manual/test_swipe_fixed.go > /dev/null 2>&1

sleep 2

# trackpad_dumpを終了
kill $DUMP_PID 2>/dev/null

echo ""
echo "=== 実際の出力 (actual_output.log) ==="
cat test/manual/actual_output.log | grep -E "Gesture|Swipe" | head -20

echo ""
echo "=== ネイティブログとの比較 ==="
echo "ネイティブ:"
cat test/debugging/nativelog | grep -A1 "4本指操作" | tail -n +2 | grep -E "Gesture.*Swipe" | head -10

echo ""
echo "詳細な差異を確認してください"