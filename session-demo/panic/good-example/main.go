//go:build js && wasm

// 良い例: defer + recover で panic を捕捉し、JS 側に通知して継続動作
package main

import (
	"fmt"
	"syscall/js"
)

// triggerPanic はボタンクリックで呼ばれる。recover で panic を捕捉し、エラーを JS に渡す。
func triggerPanic(_ js.Value, _ []js.Value) interface{} {
	defer func() {
		if e := recover(); e != nil {
			msg := fmt.Sprintf("recovered: %v", e)
			doc := js.Global().Get("document")
			elem := doc.Call("getElementById", "output")
			elem.Set("textContent", msg)
			// コンソールにも出しておく
			js.Global().Get("console").Call("log", "Go recover:", msg)
		}
	}()

	// 意図的に nil ポインタ参照で panic を発生
	var p *int
	*p = 1 // panic → 上記 defer recover で捕捉される
	return nil
}

// ping は WASM が生きているか確認する用。recover 後も正常に反応する。
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
