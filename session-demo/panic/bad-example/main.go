//go:build js && wasm

// 悪い例: panic を recover せず、WASM 全体が止まるケース
package main

import (
	"syscall/js"
)

// triggerPanic はボタンクリックで呼ばれる。recover なしで panic を発生させる。
func triggerPanic(_ js.Value, _ []js.Value) interface{} {
	// 意図的に nil ポインタ参照で panic を発生
	var p *int
	*p = 1 // panic: nil pointer dereference
	return nil
}

// ping は WASM が生きているか確認する用。panic 後は呼んでも反応しない。
func ping(_ js.Value, _ []js.Value) interface{} {
	doc := js.Global().Get("document")
	elem := doc.Call("getElementById", "output")
	elem.Set("textContent", "WASM is alive!")
	return nil
}

func main() {
	js.Global().Set("triggerPanic", js.FuncOf(triggerPanic))
	js.Global().Set("ping", js.FuncOf(ping))
	select {} // メインゴルーチンを維持
}
