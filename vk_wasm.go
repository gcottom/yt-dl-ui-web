//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"
)

func showKeyboard() {
	js.Global().Get("document").Get("activeElement").Set("contentEditable", true)
	js.Global().Get("navigator").Get("virtualKeyboard").Set("overlaysContent", true)
	js.Global().Get("navigator").Get("virtualKeyboard").Call("show")
}
