//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32                 = syscall.NewLazyDLL("user32.dll")
	gdi32                  = syscall.NewLazyDLL("gdi32.dll")
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	shell32                = syscall.NewLazyDLL("shell32.dll")
	advapi32               = syscall.NewLazyDLL("advapi32.dll")
	dwmapi                 = syscall.NewLazyDLL("dwmapi.dll")
	registerClass          = user32.NewProc("RegisterClassExW")
	createWindow           = user32.NewProc("CreateWindowExW")
	defWindowProc          = user32.NewProc("DefWindowProcW")
	destroyWindow          = user32.NewProc("DestroyWindow")
	showWindow             = user32.NewProc("ShowWindow")
	getMessage             = user32.NewProc("GetMessageW")
	translateMessage       = user32.NewProc("TranslateMessage")
	dispatchMessage        = user32.NewProc("DispatchMessageW")
	postQuitMessage        = user32.NewProc("PostQuitMessage")
	postMessage            = user32.NewProc("PostMessageW")
	registerWindowMessage  = user32.NewProc("RegisterWindowMessageW")
	findWindow             = user32.NewProc("FindWindowW")
	enumChildWindows       = user32.NewProc("EnumChildWindows")
	getClassName           = user32.NewProc("GetClassNameW")
	getWindowText          = user32.NewProc("GetWindowTextW")
	getWindowRect          = user32.NewProc("GetWindowRect")
	getClientRect          = user32.NewProc("GetClientRect")
	getForegroundWindow    = user32.NewProc("GetForegroundWindow")
	setForegroundWindow    = user32.NewProc("SetForegroundWindow")
	isWindowVisible        = user32.NewProc("IsWindowVisible")
	setWindowPos           = user32.NewProc("SetWindowPos")
	setWindowRgn           = user32.NewProc("SetWindowRgn")
	getCursorPos           = user32.NewProc("GetCursorPos")
	loadCursor             = user32.NewProc("LoadCursorW")
	setCursor              = user32.NewProc("SetCursor")
	setCapture             = user32.NewProc("SetCapture")
	releaseCapture         = user32.NewProc("ReleaseCapture")
	setTimer               = user32.NewProc("SetTimer")
	killTimer              = user32.NewProc("KillTimer")
	monitorFromWindow      = user32.NewProc("MonitorFromWindow")
	getMonitorInfo         = user32.NewProc("GetMonitorInfoW")
	getDpiForWindow        = user32.NewProc("GetDpiForWindow")
	setDpiContext          = user32.NewProc("SetProcessDpiAwarenessContext")
	setProcessDPIAware     = user32.NewProc("SetProcessDPIAware")
	beginPaint             = user32.NewProc("BeginPaint")
	endPaint               = user32.NewProc("EndPaint")
	invalidateRect         = user32.NewProc("InvalidateRect")
	getDC                  = user32.NewProc("GetDC")
	releaseDC              = user32.NewProc("ReleaseDC")
	fillRect               = user32.NewProc("FillRect")
	drawText               = user32.NewProc("DrawTextW")
	messageBox             = user32.NewProc("MessageBoxW")
	createPopupMenu        = user32.NewProc("CreatePopupMenu")
	appendMenu             = user32.NewProc("AppendMenuW")
	trackPopupMenu         = user32.NewProc("TrackPopupMenu")
	destroyMenu            = user32.NewProc("DestroyMenu")
	openClipboard          = user32.NewProc("OpenClipboard")
	closeClipboard         = user32.NewProc("CloseClipboard")
	emptyClipboard         = user32.NewProc("EmptyClipboard")
	setClipboardData       = user32.NewProc("SetClipboardData")
	createIconIndirect     = user32.NewProc("CreateIconIndirect")
	destroyIcon            = user32.NewProc("DestroyIcon")
	gdiFlush               = gdi32.NewProc("GdiFlush")
	createSolidBrush       = gdi32.NewProc("CreateSolidBrush")
	createPen              = gdi32.NewProc("CreatePen")
	createFont             = gdi32.NewProc("CreateFontW")
	selectObject           = gdi32.NewProc("SelectObject")
	deleteObject           = gdi32.NewProc("DeleteObject")
	createCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	createCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	createDIBSection       = gdi32.NewProc("CreateDIBSection")
	createBitmap           = gdi32.NewProc("CreateBitmap")
	deleteDC               = gdi32.NewProc("DeleteDC")
	bitBlt                 = gdi32.NewProc("BitBlt")
	roundRect              = gdi32.NewProc("RoundRect")
	setBkMode              = gdi32.NewProc("SetBkMode")
	setTextColor           = gdi32.NewProc("SetTextColor")
	createRoundRectRgn     = gdi32.NewProc("CreateRoundRectRgn")
	saveDC                 = gdi32.NewProc("SaveDC")
	restoreDC              = gdi32.NewProc("RestoreDC")
	intersectClipRect      = gdi32.NewProc("IntersectClipRect")
	getModuleHandle        = kernel32.NewProc("GetModuleHandleW")
	createMutex            = kernel32.NewProc("CreateMutexW")
	closeHandle            = kernel32.NewProc("CloseHandle")
	getLocalTime           = kernel32.NewProc("GetLocalTime")
	globalAlloc            = kernel32.NewProc("GlobalAlloc")
	globalLock             = kernel32.NewProc("GlobalLock")
	globalUnlock           = kernel32.NewProc("GlobalUnlock")
	globalFree             = kernel32.NewProc("GlobalFree")
	notifyIcon             = shell32.NewProc("Shell_NotifyIconW")
	shellExecute           = shell32.NewProc("ShellExecuteW")
	appBarMessage          = shell32.NewProc("SHAppBarMessage")
	regGetValue            = advapi32.NewProc("RegGetValueW")
	regCreateKey           = advapi32.NewProc("RegCreateKeyExW")
	regSetValue            = advapi32.NewProc("RegSetValueExW")
	regDeleteValue         = advapi32.NewProc("RegDeleteValueW")
	regCloseKey            = advapi32.NewProc("RegCloseKey")
	dwmGetWindowAttribute  = dwmapi.NewProc("DwmGetWindowAttribute")
)

const (
	WM_DESTROY        = 0x0002
	WM_SIZE           = 0x0005
	WM_PAINT          = 0x000F
	WM_CLOSE          = 0x0010
	WM_ERASEBKGND     = 0x0014
	WM_SETTINGCHANGE  = 0x001A
	WM_MOUSEACTIVATE  = 0x0021
	WM_DISPLAYCHANGE  = 0x007E
	WM_NCHITTEST      = 0x0084
	WM_TIMER          = 0x0113
	WM_MOUSEMOVE      = 0x0200
	WM_LBUTTONDOWN    = 0x0201
	WM_LBUTTONUP      = 0x0202
	WM_RBUTTONUP      = 0x0205
	WM_MOUSEWHEEL     = 0x020A
	WM_CAPTURECHANGED = 0x0215
	WM_DPICHANGED     = 0x02E0
	WM_APP            = 0x8000
	WM_TRAY           = WM_APP + 1
	WM_SHOWCARD       = WM_APP + 2
	WM_RESETPOS       = WM_APP + 3
	WM_RENDER         = WM_APP + 4
	WS_POPUP          = 0x80000000
	WS_EX_TOPMOST     = 0x8
	WS_EX_TOOLWINDOW  = 0x80
	WS_EX_LAYERED     = 0x00080000
	WS_EX_NOACTIVATE  = 0x08000000
	SW_HIDE           = 0
	SW_SHOWNOACTIVATE = 4
	SWP_NOACTIVATE    = 0x0010
	SWP_SHOWWINDOW    = 0x0040
	SWP_NOMOVE        = 0x0002
	SWP_NOSIZE        = 0x0001
	DT_CENTER         = 0x1
	DT_VCENTER        = 0x4
	DT_WORDBREAK      = 0x10
	DT_SINGLELINE     = 0x20
	DT_CALCRECT       = 0x400
	DT_NOPREFIX       = 0x800
	HKCU              = 0x80000001
)

var hwndTopmost = ^uintptr(0)

type WindowClass struct {
	Size, Style                        uint32
	WndProc                            uintptr
	ClsExtra, WndExtra                 int32
	Instance, Icon, Cursor, Background uintptr
	MenuName, ClassName                *uint16
	IconSmall                          uintptr
}
type Message struct {
	Window         uintptr
	Message        uint32
	WParam, LParam uintptr
	Time           uint32
	Pt             Point
	Private        uint32
}
type PaintStruct struct {
	DC                 uintptr
	Erase              int32
	Paint              Rect
	Restore, IncUpdate int32
	Reserved           [32]byte
}
type MonitorInfo struct {
	Size          uint32
	Monitor, Work Rect
	Flags         uint32
}
type AppBarData struct {
	Size           uint32
	Window         uintptr
	Callback, Edge uint32
	Rect           Rect
	Param          uintptr
}
type SystemTime struct{ Year, Month, Weekday, Day, Hour, Minute, Second, Milliseconds uint16 }
type NotifyIconData struct {
	Size                uint32
	Window              uintptr
	ID, Flags, Callback uint32
	Icon                uintptr
	Tip                 [128]uint16
	State, StateMask    uint32
	Info                [256]uint16
	Timeout             uint32
	InfoTitle           [64]uint16
	InfoFlags           uint32
	Guid                [16]byte
	BalloonIcon         uintptr
}
type BitmapInfo struct {
	Size                   uint32
	Width, Height          int32
	Planes, BitCount       uint16
	Compression, ImageSize uint32
	XPels, YPels           int32
	Used, Important        uint32
	Colors                 [1]uint32
}
type IconInfo struct {
	IsIcon      uint32
	HotX, HotY  uint32
	Mask, Color uintptr
}

func w(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(strings.ReplaceAll(s, "\x00", ""))
	return p
}
func u(v int) uintptr { return uintptr(v) }
func color(hex uint32) uintptr {
	return uintptr(((hex >> 16) & 255) | (hex & 0x00ff00) | ((hex & 255) << 16))
}
func className(h uintptr) string {
	var b [256]uint16
	n, _, _ := getClassName.Call(h, uintptr(unsafe.Pointer(&b[0])), u(len(b)))
	return syscall.UTF16ToString(b[:n])
}
func windowRect(h uintptr) Rect {
	var r Rect
	getWindowRect.Call(h, uintptr(unsafe.Pointer(&r)))
	return r
}
func clientRect(h uintptr) Rect {
	var r Rect
	getClientRect.Call(h, uintptr(unsafe.Pointer(&r)))
	return r
}
func cursorPoint() Point     { var p Point; getCursorPos.Call(uintptr(unsafe.Pointer(&p))); return p }
func visible(h uintptr) bool { r, _, _ := isWindowVisible.Call(h); return r != 0 }
func invalidate(h uintptr) {
	if h != 0 {
		postMessage.Call(h, WM_RENDER, 0, 0)
	}
}
func msgbox(h uintptr, title, text string) {
	messageBox.Call(h, uintptr(unsafe.Pointer(w(text))), uintptr(unsafe.Pointer(w(title))), 0x40)
}
func errorbox(h uintptr, text string) {
	messageBox.Call(h, uintptr(unsafe.Pointer(w(text))), uintptr(unsafe.Pointer(w(activeLocale().Name()))), 0x10)
}
func wallClockNow() time.Time {
	var s SystemTime
	getLocalTime.Call(uintptr(unsafe.Pointer(&s)))
	return time.Date(int(s.Year), time.Month(s.Month), int(s.Day), int(s.Hour), int(s.Minute), int(s.Second), int(s.Milliseconds)*1e6, time.UTC)
}
func enableDPI() {
	if setDpiContext.Find() == nil {
		r, _, _ := setDpiContext.Call(^uintptr(3))
		if r != 0 {
			return
		}
	}
	setProcessDPIAware.Call()
}
func windowDPI(h uintptr) int {
	if getDpiForWindow.Find() == nil {
		r, _, _ := getDpiForWindow.Call(h)
		if r >= 96 && r <= 768 {
			return int(r)
		}
	}
	return 96
}
func monitorInfo(h uintptr) MonitorInfo {
	m, _, _ := monitorFromWindow.Call(h, 2)
	mi := MonitorInfo{}
	mi.Size = uint32(unsafe.Sizeof(mi))
	getMonitorInfo.Call(m, uintptr(unsafe.Pointer(&mi)))
	return mi
}
func systemLight() bool {
	var val uint32
	size := uint32(4)
	r, _, _ := regGetValue.Call(HKCU, uintptr(unsafe.Pointer(w(`Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`))), uintptr(unsafe.Pointer(w("SystemUsesLightTheme"))), 0x10, 0, uintptr(unsafe.Pointer(&val)), uintptr(unsafe.Pointer(&size)))
	return r == 0 && val != 0
}

const runKey = `Software\Microsoft\Windows\CurrentVersion\Run`

func readAutostart(name string) string {
	var data [2048]uint16
	size := uint32(len(data) * 2)
	r, _, _ := regGetValue.Call(HKCU, uintptr(unsafe.Pointer(w(runKey))), uintptr(unsafe.Pointer(w(name))), 0x2, 0, uintptr(unsafe.Pointer(&data[0])), uintptr(unsafe.Pointer(&size)))
	if r != 0 {
		return ""
	}
	return strings.Trim(syscall.UTF16ToString(data[:]), "\" ")
}
func autostartEnabled(exe string) bool { return strings.EqualFold(readAutostart(appID), exe) }
func setAutostart(exe string, enabled bool) error {
	var key uintptr
	r, _, _ := regCreateKey.Call(HKCU, uintptr(unsafe.Pointer(w(runKey))), 0, 0, 0, 0x0002, 0, uintptr(unsafe.Pointer(&key)), 0)
	if r != 0 {
		return fmt.Errorf("cannot open current-user startup settings (%d)", r)
	}
	defer regCloseKey.Call(key)
	if enabled {
		v, _ := syscall.UTF16FromString(`"` + exe + `"`)
		r, _, _ = regSetValue.Call(key, uintptr(unsafe.Pointer(w(appID))), 0, 1, uintptr(unsafe.Pointer(&v[0])), u(len(v)*2))
	} else {
		r, _, _ = regDeleteValue.Call(key, uintptr(unsafe.Pointer(w(appID))))
		if r == 2 {
			r = 0
		}
	}
	if r != 0 {
		return fmt.Errorf("cannot save startup settings (%d)", r)
	}
	// Remove only this application's recognized legacy Run entry, never other apps.
	legacy := readAutostart("QiyaoShichen")
	if legacy != "" && strings.EqualFold(filepath.Base(legacy), "QiyaoShichen.exe") {
		r, _, _ = regDeleteValue.Call(key, uintptr(unsafe.Pointer(w("QiyaoShichen"))))
		if r != 0 && r != 2 {
			return fmt.Errorf("cannot remove legacy startup entry (%d)", r)
		}
	}
	return nil
}
func openPath(h uintptr, p string) {
	r, _, _ := shellExecute.Call(h, uintptr(unsafe.Pointer(w("open"))), uintptr(unsafe.Pointer(w(p))), 0, 0, 1)
	if r <= 32 {
		errorbox(h, activeLocale().Text("无法打开：", "Cannot open: ")+p+activeLocale().Text("\n请手动打开这个文件或文件夹。", "\nPlease open the file or folder manually."))
	}
}
func copyText(h uintptr, s string) error {
	r, _, _ := openClipboard.Call(h)
	if r == 0 {
		return fmt.Errorf("%s", activeLocale().Text("剪贴板暂时被占用，请稍后再试", "The clipboard is busy. Please try again."))
	}
	defer closeClipboard.Call()
	v, _ := syscall.UTF16FromString(s)
	mem, _, _ := globalAlloc.Call(0x0002, u(len(v)*2))
	if mem == 0 {
		return fmt.Errorf("%s", activeLocale().Text("无法分配剪贴板内存", "Cannot allocate clipboard memory."))
	}
	p, _, _ := globalLock.Call(mem)
	if p == 0 {
		globalFree.Call(mem)
		return fmt.Errorf("%s", activeLocale().Text("无法写入剪贴板", "Cannot write to the clipboard."))
	}
	// p is valid external memory returned by GlobalLock; the HGLOBAL remains owned here.
	copy(unsafe.Slice((*uint16)(unsafe.Pointer(p)), len(v)), v)
	globalUnlock.Call(mem)
	emptyClipboard.Call()
	r, _, _ = setClipboardData.Call(13, mem)
	if r == 0 {
		globalFree.Call(mem)
		return fmt.Errorf("%s", activeLocale().Text("无法更新剪贴板", "Cannot update the clipboard."))
	}
	return nil
}
func logError(e any) {
	p := filepath.Join(filepath.Dir(settingsPath()), "error.log")
	_ = os.MkdirAll(filepath.Dir(p), 0700)
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err == nil {
		defer f.Close()
		fmt.Fprintf(f, "%s %v\n", time.Now().Format(time.RFC3339), e)
	}
}

func windowTitle(h uintptr) string {
	var b [256]uint16
	n, _, _ := getWindowText.Call(h, uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)))
	if n >= uintptr(len(b)) {
		n = uintptr(len(b) - 1)
	}
	return syscall.UTF16ToString(b[:n])
}
