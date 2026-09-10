//go:build windows && amd64

package main

import "unsafe"

// Compile-time assertions for the Windows x64 ABI structures passed to Win32.
var _ [80 - unsafe.Sizeof(WindowClass{})]byte
var _ [unsafe.Sizeof(WindowClass{}) - 80]byte
var _ [48 - unsafe.Sizeof(Message{})]byte
var _ [unsafe.Sizeof(Message{}) - 48]byte
var _ [72 - unsafe.Sizeof(PaintStruct{})]byte
var _ [unsafe.Sizeof(PaintStruct{}) - 72]byte
var _ [976 - unsafe.Sizeof(NotifyIconData{})]byte
var _ [unsafe.Sizeof(NotifyIconData{}) - 976]byte
var _ [40 - unsafe.Sizeof(MonitorInfo{})]byte
var _ [unsafe.Sizeof(MonitorInfo{}) - 40]byte
var _ [32 - unsafe.Sizeof(IconInfo{})]byte
var _ [unsafe.Sizeof(IconInfo{}) - 32]byte
