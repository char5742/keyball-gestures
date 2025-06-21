#!/bin/bash

echo "=== Touchpad Debug Test ==="
echo "このスクリプトはデバッグログを有効にして手動テストを実行します。"
echo ""
echo "実行前に："
echo "1. 別のターミナルで trackpad_dump を実行"
echo "   cd test/debugging && ./trackpad_dump"
echo ""

# デバッグログを有効にして実行
export TOUCHPAD_DEBUG=1
./test/manual/manual_touchpad 2>&1 | tee touchpad_debug.log

echo ""
echo "デバッグログは touchpad_debug.log に保存されました。"
echo "activeCount の値を確認してください。"