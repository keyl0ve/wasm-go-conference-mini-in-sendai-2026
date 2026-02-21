//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"
)

// ① fmt.Println を使ったシンプルなログ
func logWithPrintln(_ js.Value, args []js.Value) any {
	value := args[0].String()
	fmt.Println("debug:", value)
	return nil
}

// ② fmt.Sprintf でフォーマット
func logWithSprintf(_ js.Value, args []js.Value) any {
	key := args[0].String()
	val := args[1].String()
	fmt.Printf("key=%v value=%v\n", key, val)
	return nil
}

// ③ console.Call で複数引数（ラベル付き）
func logWithConsoleCall(_ js.Value, args []js.Value) any {
	label := args[0].String()
	data := args[1].String()
	js.Global().Get("console").Call("log", label, js.ValueOf(data))
	return nil
}

func main() {
	js.Global().Set("logWithPrintln", js.FuncOf(logWithPrintln))
	js.Global().Set("logWithSprintf", js.FuncOf(logWithSprintf))
	js.Global().Set("logWithConsoleCall", js.FuncOf(logWithConsoleCall))
	select {}
}
