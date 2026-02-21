//go:build js && wasm

// 理想的な例: js.Global().Set() で Go 関数を1つだけ公開し、addEventListener を使わない
package main

import (
	"strconv"
	"syscall/js"
)

var clickCount int

// handleClick は onclick から呼ばれる。js.FuncOf は main で1回だけ登録。
func handleClick(_ js.Value, _ []js.Value) interface{} {
	clickCount++
	doc := js.Global().Get("document")
	doc.Call("getElementById", "output").Set("textContent", "Clicked! (count: "+strconv.Itoa(clickCount)+")")
	js.Global().Get("console").Call("log", "handleClick called, count:", clickCount)
	return nil
}

// incrementCounter も main で1回だけ登録。
func incrementCounter(_ js.Value, _ []js.Value) interface{} {
	doc := js.Global().Get("document")
	counterElem := doc.Call("getElementById", "counter")
	currentText := counterElem.Get("textContent").String()

	count, _ := strconv.Atoi(currentText)
	count++
	counterElem.Set("textContent", strconv.Itoa(count))
	return nil
}

func main() {
	// js.FuncOf を main で1回だけ登録。
	// addEventListener を繰り返し使わないので、リスナーが増え続けることがない。
	// プログラムが終わるまで使うので、Release() しなくても実害は少ない。
	js.Global().Set("handleClick", js.FuncOf(handleClick))
	js.Global().Set("incrementCounter", js.FuncOf(incrementCounter))

	select {} // メインゴルーチンを維持
}
