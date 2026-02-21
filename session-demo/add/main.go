//go:build js && wasm

package main

import "syscall/js"

func add(_ js.Value, args []js.Value) any {
	a := args[0].Int()
	b := args[1].Int()
	return a + b
}

func main() {
	js.Global().Set("add", js.FuncOf(add))
	select {} // プロセスを維持（ブラウザでは必須）
}
