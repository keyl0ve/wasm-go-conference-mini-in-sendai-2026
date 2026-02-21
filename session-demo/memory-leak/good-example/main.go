//go:build js && wasm

// 良い例: ① removeEventListener してから Release() する ② または Go 関数をグローバルに公開して onclick で呼ぶ（リスナー管理不要）
package main

import (
	"strconv"
	"syscall/js"
)

var (
	currentListener js.Func
	hasListener     bool
	replaceCount    int
)

func replaceListener(_ js.Value, _ []js.Value) interface{} {
	doc := js.Global().Get("document")
	target := doc.Call("getElementById", "target")
	countElem := doc.Call("getElementById", "replaceCount")

	if hasListener {
		target.Call("removeEventListener", "click", currentListener)
		currentListener.Release()
	}

	id := replaceCount + 1
	cb := js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		js.Global().Get("console").Call("log", "listener fired (replaced, Release used) #"+strconv.Itoa(id))
		return nil
	})
	target.Call("addEventListener", "click", cb)
	currentListener = cb
	hasListener = true
	replaceCount++
	countElem.Set("textContent", strconv.Itoa(replaceCount))
	return nil
}

// handleByGo は onclick から呼ぶ用。js.FuncOf は main で1つだけ登録し、リスナーを増やさない。
func handleByGo(_ js.Value, _ []js.Value) interface{} {
	js.Global().Get("console").Call("log", "click (onclick + Go global, no addEventListener)")
	doc := js.Global().Get("document")
	doc.Call("getElementById", "output2").Set("textContent", "clicked!")
	return nil
}

func main() {
	js.Global().Set("replaceListener", js.FuncOf(replaceListener))
	js.Global().Set("handleByGo", js.FuncOf(handleByGo))
	select {}
}
