//go:build darwin

package main

import "fmt"

func register() error {
	fmt.Println("macOS registration is not yet automated.")
	fmt.Println()
	fmt.Println("To register browser-router on macOS:")
	fmt.Println("  1. Wrap the binary in a .app bundle with an Info.plist that declares")
	fmt.Println("     CFBundleURLSchemes for http and https.")
	fmt.Println("  2. Open System Settings > General > Default web browser and select it.")
	fmt.Println()
	fmt.Println("Alternatively, use the 'defaultbrowser' CLI tool:")
	fmt.Println("  brew install defaultbrowser && defaultbrowser browser-router")
	return nil
}

func unregister() error {
	fmt.Println("Open System Settings > General > Default web browser and select another browser.")
	return nil
}
