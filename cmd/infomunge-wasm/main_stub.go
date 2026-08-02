//go:build !js

package main

// The real entrypoint is compiled only for the JavaScript/WASM target. Keep
// the package buildable in ordinary host-wide Go checks so its adapter logic
// can have normal unit-test coverage.
func main() {}
