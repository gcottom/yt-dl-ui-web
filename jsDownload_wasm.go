//go:build js && wasm
// +build js,wasm

package main

import (
	"fmt"

	"syscall/js"
)
func jsDownload(outBlob []byte, filename string) {
	downloadFile := js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		data := outBlob
		uia := js.Global().Get("Uint8Array").New(len(data))
		js.CopyBytesToJS(uia, data)
		blob := js.Global().Get("Blob").New([]interface{}{uia})
		url := js.Global().Get("URL").Call("createObjectURL", blob)
		fmt.Println(url)
		// Create a download link element
		downloadLink := js.Global().Get("document").Call("createElement", "a")
		downloadLink.Set("href", url)
		downloadLink.Set("download", filename) // Set the desired file name

		// Trigger a click event on the download link
		clickEvent := js.Global().Get("document").Call("createEvent", "MouseEvents")
		clickEvent.Call("initEvent", "click", true, true)
		downloadLink.Call("dispatchEvent", clickEvent)

		// Revoke the Object URL to release the resource
		js.Global().Get("URL").Call("revokeObjectURL", url)

		return nil
	})

	js.Global().Set("downloadFile", downloadFile) // Expose the Go function to JavaScript
	js.Global().Call("downloadFile")
	trackSavedNotif()
	showMainScreen()
}