//go:build js && wasm

// 悪い例: js.FuncOf で作ったコールバックを Release() せずに addEventListener で登録し続ける → メモリリーク
package main

import (
	"strconv"
	"syscall/js"
)

var listenerCount int

func addListener(_ js.Value, _ []js.Value) interface{} {
	doc := js.Global().Get("document")
	target := doc.Call("getElementById", "target")
	countElem := doc.Call("getElementById", "count")

	id := listenerCount + 1
	cb := js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		js.Global().Get("console").Call("log", "listener fired (no Release) #"+strconv.Itoa(id))
		return nil
	})
	target.Call("addEventListener", "click", cb)
	// cb.Release() を呼ばない → リスナーが増えるたびにメモリリーク

	listenerCount++
	countElem.Set("textContent", strconv.Itoa(listenerCount))
	return nil
}

func addListener100(_ js.Value, _ []js.Value) interface{} {
	for i := 0; i < 100; i++ {
		addListener(js.Value{}, nil)
	}
	return nil
}

func main() {
	js.Global().Set("addListener", js.FuncOf(addListener))
	js.Global().Set("addListener100", js.FuncOf(addListener100))
	select {}
}
