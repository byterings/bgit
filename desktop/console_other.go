//go:build !windows

package main

// attachParentConsole is only needed on Windows, where a windowed binary is
// not connected to the console that started it.
func attachParentConsole() {}
