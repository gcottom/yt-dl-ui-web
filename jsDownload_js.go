//go:build js && !wasm
// +build js,!wasm

package main

import (
	"fmt"

	"github.com/gopherjs/gopherjs/js"
)
func jsDownload(outBlob []byte, filename string) {
	downloadFile := js.MakeFunc(func(this *js.Object, p []*js.Object) interface{} {
		data := outBlob // Replace with your file content
		blob := js.Global.Get("Blob").New([]interface{}{data})
		url := js.Global.Get("URL").Call("createObjectURL", blob)
		fmt.Println(url)
		// Create a download link element
		downloadLink := js.Global.Get("document").Call("createElement", "a")
		downloadLink.Set("href", url)
		downloadLink.Set("download", filename) // Set the desired file name

		// Trigger a click event on the download link
		clickEvent := js.Global.Get("document").Call("createEvent", "MouseEvents")
		clickEvent.Call("initEvent", "click", true, true)
		downloadLink.Call("dispatchEvent", clickEvent)

		// Revoke the Object URL to release the resource
		js.Global.Get("URL").Call("revokeObjectURL", url)

		return nil
	})

	js.Global.Set("downloadFile", downloadFile) // Expose the Go function to JavaScript
	js.Global.Call("downloadFile")
	trackSavedNotif()
	showMainScreen()
}