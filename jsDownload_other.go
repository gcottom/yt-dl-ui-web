//go:build !js && !wasm
// +build !js,!wasm

package main

import (
	"fmt"
)

func jsDownload(outBlob []byte, filename string) {
	fmt.Println("other jsdownload")
}
