//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"
	"time"
)

func heavyComputation() int {
	// 重い計算をシミュレート（3秒待機）
	time.Sleep(3 * time.Second)
	return 42 // 計算結果
}

func startHeavyTask(_ js.Value, _ []js.Value) any {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				js.Global().Get("console").Call("error", "Panic:", r)
			}
		}()
		result := heavyComputation()
		callback := js.Global().Get("onTaskComplete")
		callback.Invoke(js.ValueOf(result))
		fmt.Println("タスクが完了しました")
	}()

	fmt.Println("タスクが開始されました")

	return nil
}

func main() {
	js.Global().Set("startHeavyTask", js.FuncOf(startHeavyTask))
	select {}
}
