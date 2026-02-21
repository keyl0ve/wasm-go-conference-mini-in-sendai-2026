//go:build js && wasm

package main

import (
	"runtime/debug"
	"syscall/js"
)

// 共通: スタックトレースをDOMに表示
func showStack(stack []byte) {
	js.Global().Get("document").
		Call("getElementById", "output").
		Set("textContent", string(stack))
}

// ① 直接 panic（1段）
func triggerSimplePanic(_ js.Value, _ []js.Value) any {
	defer func() {
		if r := recover(); r != nil {
			showStack(debug.Stack())
		}
	}()
	panic("something went wrong")
}

// ② ネストした関数呼び出し（A → B → C → panic）
func triggerNestedPanic(_ js.Value, _ []js.Value) any {
	defer func() {
		if r := recover(); r != nil {
			showStack(debug.Stack())
		}
	}()
	funcA()
	return nil
}

func funcA() { funcB() }
func funcB() { funcC() }
func funcC() { panic("deep panic in funcC") }

// nilDeref は nil ポインタを間接参照してランタイム panic を起こす。
// 別関数に切り出すことでスタックトレースに関数名が現れ、読み方の練習になる。
func nilDeref(p *int) int { return *p }

// ③ nil ポインタ参照（ランタイム panic）
func triggerNilPanic(_ js.Value, _ []js.Value) any {
	defer func() {
		if r := recover(); r != nil {
			showStack(debug.Stack())
		}
	}()
	var p *int
	_ = nilDeref(p)
	return nil
}

func main() {
	js.Global().Set("triggerSimplePanic", js.FuncOf(triggerSimplePanic))
	js.Global().Set("triggerNestedPanic", js.FuncOf(triggerNestedPanic))
	js.Global().Set("triggerNilPanic", js.FuncOf(triggerNilPanic))
	select {}
}
