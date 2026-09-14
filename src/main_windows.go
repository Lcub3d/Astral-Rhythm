//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const widgetClass = "QiyaoShichen.Widget.v1"
const cardClass = "QiyaoShichen.Card.v1"

type App struct {
	Preview                                      PreviewState
	PreviewInput                                 previewInput
	PreviewClock                                 string
	L                                            Localizer
	HWND, Popup, Taskbar, Icon                   uintptr
	Exe, Base, Source                            string
	Data                                         Data
	Settings                                     Settings
	DPI                                          int
	Colors                                       Palette
	Fonts                                        map[string]uintptr
	TaskRect, WidgetRect                         Rect
	Monitor                                      MonitorInfo
	Anchor                                       int
	Vertical, Shown, CardShown, Pinned, MenuOpen bool
	MouseDown, Dragging                          bool
	DragStart                                    Point
	DragGap                                      int
	HoverSince, LeaveSince                       time.Time
	Blocks                                       []TextBlock
	Panels                                       []VisualPanel
	DrawIcons                                    []VisualIcon
	Progress                                     Rect
	Buttons                                      []Button
	Body                                         Rect
	CardWidth, CardHeight, ContentHeight, Scroll int
	TaskbarLight                                 bool
	MouseOverCard                                bool
	CardHoverID                                  int
	TaskbarCreated                               uint32
	LastClock, LastStamp                         string
	TickCount                                    int
	Status                                       string
	StatusUntil                                  time.Time
	Closing                                      bool
}

var app *App
var traySearchRect Rect
var traySearchFound bool
var enumTrayCallback = syscall.NewCallback(func(h, l uintptr) uintptr {
	if className(h) == "TrayNotifyWnd" {
		r := windowRect(h)
		if r.W() > 0 && r.H() > 0 && visible(h) {
			traySearchRect = r
			traySearchFound = true
			return 0
		}
	}
	return 1
})

func (a *App) S(n int) int {
	if n < 0 {
		return -((-n*a.DPI + 48) / 96)
	}
	return (n*a.DPI + 48) / 96
}
func (a *App) font(size int, bold bool) uintptr {
	px := a.S(size)
	k := fmt.Sprintf("%d/%t", px, bold)
	if f := a.Fonts[k]; f != 0 {
		return f
	}
	weight := 400
	if bold {
		weight = 600
	}
	face := a.L.Text("Microsoft YaHei UI", "Segoe UI")
	f, _, _ := createFont.Call(u(-px), 0, 0, 0, u(weight), 0, 0, 0, 1, 0, 0, 4, 0, uintptr(unsafe.Pointer(w(face))))
	a.Fonts[k] = f
	return f
}
func (a *App) clearFonts() {
	for _, f := range a.Fonts {
		deleteObject.Call(f)
	}
	a.Fonts = map[string]uintptr{}
	clearArtCache()
	clearGlyphMasks()
}
func (a *App) theme() {
	a.TaskbarLight = systemLight()
	a.Colors = darkPalette
	if a.Settings.Theme == "light" || (a.Settings.Theme == "system" && systemLight()) {
		a.Colors = lightPalette
	}
}
func rectFill(dc uintptr, r Rect, c uint32) {
	b, _, _ := createSolidBrush.Call(color(c))
	fillRect.Call(dc, uintptr(unsafe.Pointer(&r)), b)
	deleteObject.Call(b)
}
func rounded(dc uintptr, r Rect, c, border uint32, rad int) {
	brush, _, _ := createSolidBrush.Call(color(c))
	pen, _, _ := createPen.Call(0, 1, color(border))
	ob, _, _ := selectObject.Call(dc, brush)
	op, _, _ := selectObject.Call(dc, pen)
	roundRect.Call(dc, u(int(r.Left)), u(int(r.Top)), u(int(r.Right)), u(int(r.Bottom)), u(rad), u(rad))
	selectObject.Call(dc, ob)
	selectObject.Call(dc, op)
	deleteObject.Call(brush)
	deleteObject.Call(pen)
}
func (a *App) text(dc uintptr, text string, r Rect, size int, bold bool, c uint32, flags int) {
	old, _, _ := selectObject.Call(dc, a.font(size, bold))
	setBkMode.Call(dc, 1)
	setTextColor.Call(dc, color(c))
	drawText.Call(dc, uintptr(unsafe.Pointer(w(text))), ^uintptr(0), uintptr(unsafe.Pointer(&r)), u(flags|DT_NOPREFIX))
	selectObject.Call(dc, old)
}
func (a *App) measure(dc uintptr, text string, width, size int, bold bool) int {
	r := R(0, 0, width, 0)
	old, _, _ := selectObject.Call(dc, a.font(size, bold))
	drawText.Call(dc, uintptr(unsafe.Pointer(w(text))), ^uintptr(0), uintptr(unsafe.Pointer(&r)), DT_CALCRECT|DT_WORDBREAK|DT_NOPREFIX)
	selectObject.Call(dc, old)
	return max(r.H(), a.S(size+3))
}
func roundedRegion(h uintptr, width, height, rad int) {
	r, _, _ := createRoundRectRgn.Call(0, 0, u(width+1), u(height+1), u(rad), u(rad))
	if r != 0 {
		ok, _, _ := setWindowRgn.Call(h, r, 1)
		if ok == 0 {
			deleteObject.Call(r)
		}
	}
}
func (a *App) paint(h uintptr, isCard bool) {
	var ps PaintStruct
	beginPaint.Call(h, uintptr(unsafe.Pointer(&ps)))
	endPaint.Call(h, uintptr(unsafe.Pointer(&ps)))
	a.render(h, isCard)
}
func (a *App) render(h uintptr, isCard bool) {
	if h == 0 || a.Closing {
		return
	}
	r := clientRect(h)
	if r.W() < 1 || r.H() < 1 {
		return
	}
	surface, e := newNativeSurface(r.W(), r.H())
	if e != nil {
		logError(e)
		return
	}
	defer surface.Close()
	if isCard {
		a.paintCard(surface.PixelBuffer)
	} else {
		a.paintWidget(surface.PixelBuffer)
	}
	if e = surface.Present(h); e != nil {
		logError(e)
	}
}
func (a *App) positionCard() {
	a.makeLayout()
	x := int(a.WidgetRect.Left) - a.S(16)
	y := int(a.TaskRect.Top) - a.CardHeight - a.S(7)
	if a.Vertical {
		x = int(a.TaskRect.Left) - a.CardWidth - a.S(7)
		if a.TaskRect.Left <= a.Monitor.Monitor.Left+int32(a.S(5)) {
			x = int(a.TaskRect.Right) + a.S(7)
		}
		y = int(a.WidgetRect.Bottom) - a.CardHeight
	} else if a.TaskRect.Top <= a.Monitor.Monitor.Top+int32(a.S(5)) {
		y = int(a.TaskRect.Bottom) + a.S(7)
	}
	x = clamp(x, int(a.Monitor.Work.Left)+a.S(8), int(a.Monitor.Work.Right)-a.CardWidth-a.S(8))
	y = clamp(y, int(a.Monitor.Work.Top)+a.S(8), int(a.Monitor.Work.Bottom)-a.CardHeight-a.S(8))
	setWindowPos.Call(a.Popup, hwndTopmost, u(x), u(y), u(a.CardWidth), u(a.CardHeight), SWP_NOACTIVATE|SWP_SHOWWINDOW)
	// Per-pixel rounded corners; no hard window region.
	invalidate(a.Popup)
}
func (a *App) showCard(pin bool) {
	if a.MenuOpen || a.Dragging {
		return
	}
	a.Pinned = pin
	if !a.CardShown {
		a.Preview.Reset()
		a.Scroll = 0
	}
	a.CardShown = true
	a.LeaveSince = time.Time{}
	a.positionCard()
	invalidate(a.HWND)
}
func (a *App) hideCard() {
	a.Preview.Reset()
	if a.CardShown {
		showWindow.Call(a.Popup, SW_HIDE)
	}
	a.CardShown = false
	a.Pinned = false
	a.LeaveSince = time.Time{}
	invalidate(a.HWND)
}
func (a *App) fullScreen() bool {
	if !a.Settings.HideFullscreen {
		return false
	}
	fg, _, _ := getForegroundWindow.Call()
	if fg == 0 || fg == a.HWND || fg == a.Popup || fg == a.Taskbar {
		return false
	}
	switch className(fg) {
	case "Progman", "WorkerW", "Shell_TrayWnd", "Shell_SecondaryTrayWnd":
		return false
	}
	if !visible(fg) {
		return false
	}
	fm := monitorInfo(fg)
	if fm.Monitor != a.Monitor.Monitor {
		return false
	}
	r := windowRect(fg)
	if dwmGetWindowAttribute.Find() == nil {
		var ext Rect
		ok, _, _ := dwmGetWindowAttribute.Call(fg, 9, uintptr(unsafe.Pointer(&ext)), u(int(unsafe.Sizeof(ext))))
		if ok == 0 && ext.W() > 0 {
			r = ext
		}
	}
	m := a.Monitor.Monitor
	return r.Left <= m.Left+1 && r.Top <= m.Top+1 && r.Right >= m.Right-1 && r.Bottom >= m.Bottom-1
}
func (a *App) positionWidget() {
	task, _, _ := findWindow.Call(uintptr(unsafe.Pointer(w("Shell_TrayWnd"))), 0)
	if task == 0 {
		a.Shown = false
		showWindow.Call(a.HWND, SW_HIDE)
		if !a.Pinned {
			a.hideCard()
		}
		return
	}
	a.Taskbar = task
	a.Monitor = monitorInfo(task)
	dpi := windowDPI(task)
	if dpi != a.DPI {
		a.DPI = dpi
		a.clearFonts()
	}
	r := windowRect(task)
	a.Vertical = r.H() > r.W()
	intersection := R(max(int(r.Left), int(a.Monitor.Monitor.Left)), max(int(r.Top), int(a.Monitor.Monitor.Top)), 0, 0)
	intersection.Right = min32(r.Right, a.Monitor.Monitor.Right)
	intersection.Bottom = min32(r.Bottom, a.Monitor.Monitor.Bottom)
	hidden := !visible(task) || (!a.Vertical && intersection.H() < a.S(10)) || (a.Vertical && intersection.W() < a.S(10))
	if a.fullScreen() {
		a.Shown = false
		showWindow.Call(a.HWND, SW_HIDE)
		a.hideCard()
		return
	}
	if hidden {
		a.Shown = false
		showWindow.Call(a.HWND, SW_HIDE)
		if !a.Pinned && (!a.CardShown || !windowRect(a.Popup).Contains(cursorPoint())) {
			a.hideCard()
		}
		return
	}
	a.TaskRect = r
	traySearchFound = false
	enumChildWindows.Call(task, enumTrayCallback, 0)
	anchor := int(r.Right) - a.S(265)
	if a.Vertical {
		anchor = int(r.Bottom) - a.S(200)
	}
	if traySearchFound {
		if !a.Vertical && traySearchRect.Left > r.Left+int32(r.W()/3) && traySearchRect.Left < r.Right {
			anchor = int(traySearchRect.Left)
		}
		if a.Vertical && traySearchRect.Top > r.Top+int32(r.H()/3) && traySearchRect.Top < r.Bottom {
			anchor = int(traySearchRect.Top)
		}
	}
	a.Anchor = anchor
	width := a.widgetWidth()
	height := a.S(a.Settings.FontSize*2 + 14)
	x, y := 0, 0
	if !a.Vertical {
		height = min(height, r.H()-a.S(4))
		x = anchor - width - a.S(a.Settings.Gap)
		y = int(r.Top) + (r.H()-height)/2
		x = clamp(x, int(r.Left)+a.S(8), int(r.Right)-width-a.S(8))
	} else {
		width = min(width, r.W()-a.S(4))
		x = int(r.Left) + (r.W()-width)/2
		y = anchor - height - a.S(a.Settings.Gap)
		y = clamp(y, int(r.Top)+a.S(8), int(r.Bottom)-height-a.S(8))
	}
	wr := R(x, y, width, height)
	changed := a.WidgetRect != wr
	a.WidgetRect = wr
	a.Shown = true
	setWindowPos.Call(a.HWND, hwndTopmost, u(x), u(y), u(width), u(height), SWP_NOACTIVATE|SWP_SHOWWINDOW)
	if changed {
		// No plate and no rounded region on the transparent taskbar button.
		invalidate(a.HWND)
		if a.CardShown {
			a.positionCard()
		}
	}
	if a.CardShown {
		setWindowPos.Call(a.Popup, hwndTopmost, 0, 0, 0, 0, SWP_NOACTIVATE|SWP_NOMOVE|SWP_NOSIZE)
	}
}
func min32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
func (a *App) save() {
	a.Settings.Normalize()
	if e := SaveSettings(a.Settings); e != nil {
		logError(e)
	}
}
func (a *App) tableStamp() string {
	var b strings.Builder
	for _, p := range []string{filepath.Join(a.Base, "表格", hourFile), filepath.Join(a.Base, "表格", dayFile), filepath.Join(a.Base, "data.json"), filepath.Join(a.Base, "locales", "en.json")} {
		if s, e := os.Stat(p); e == nil {
			fmt.Fprintf(&b, "%s:%d:%d;", p, s.Size(), s.ModTime().UnixNano())
		}
	}
	return b.String()
}
func (a *App) reload(explicit bool) {
	data, source, e := LoadData(a.Base)
	if e != nil {
		logError(e)
		if explicit {
			errorbox(a.HWND, a.L.Text("读取失败，已保留当前有效数据。\n\n", "Read failed. The last valid data has been kept.\n\n")+e.Error()+a.L.Text("\n\n请保存并关闭正在编辑的表格后重试。", "\n\nSave and close the workbook, then try again."))
		} else {
			a.Status = a.L.Text("表格暂未读入；继续使用上一份有效数据", "Workbook unavailable; keeping the last valid data")
			a.StatusUntil = time.Now().Add(15 * time.Second)
		}
		return
	}
	a.L = loadLocalizer(a.Settings.Language, a.Base)
	a.Data = data
	a.Source = source
	a.LastStamp = a.tableStamp()
	a.LastClock = ""
	invalidate(a.HWND)
	if a.CardShown {
		a.positionCard()
	}
	if explicit {
		a.Status = a.L.Text("已重新读取：", "Reloaded: ") + a.L.Content(source)
		a.StatusUntil = time.Now().Add(4 * time.Second)
		a.showCard(true)
	}
}
func (a *App) tick() {
	if a.Closing || a.MenuOpen {
		return
	}
	a.pollPreviewDismissal()
	a.TickCount++
	a.positionWidget()
	now := wallClockNow()
	clock := now.Format("2006-01-02 15:04")
	if clock != a.LastClock {
		a.LastClock = clock
		invalidate(a.HWND)
		a.tray(1)
		if a.CardShown {
			a.positionCard()
		}
	}
	if a.TickCount%10 == 0 {
		old := a.Colors
		oldTaskbarLight := a.TaskbarLight
		a.theme()
		if old != a.Colors || oldTaskbarLight != a.TaskbarLight {
			invalidate(a.HWND)
			if a.CardShown {
				a.positionCard()
			}
		}
	}
	if a.TickCount%25 == 0 {
		if stamp := a.tableStamp(); stamp != a.LastStamp {
			a.reload(false)
		}
	}
	if a.MenuOpen || a.MouseDown {
		return
	}
	p := cursorPoint()
	t := time.Now()
	onWidget := a.Shown && a.WidgetRect.Contains(p)
	onCard := a.CardShown && windowRect(a.Popup).Contains(p)
	if onCard != a.MouseOverCard {
		a.MouseOverCard = onCard
		a.CardHoverID = 0
		if a.CardShown {
			invalidate(a.Popup)
		}
	}
	if onWidget {
		if a.HoverSince.IsZero() {
			a.HoverSince = t
			invalidate(a.HWND)
		}
		a.LeaveSince = time.Time{}
		if !a.CardShown && t.Sub(a.HoverSince) >= 350*time.Millisecond {
			a.showCard(false)
		}
	} else {
		if !a.HoverSince.IsZero() {
			invalidate(a.HWND)
		}
		a.HoverSince = time.Time{}
	}
	if onWidget || onCard {
		a.LeaveSince = time.Time{}
	} else if a.CardShown && !a.Pinned && !a.Preview.Active() {
		if a.LeaveSince.IsZero() {
			a.LeaveSince = t
		}
		if t.Sub(a.LeaveSince) >= 450*time.Millisecond {
			a.hideCard()
		}
	}
}
func (a *App) createTrayIcon() uintptr {
	v, e := artBitmap("assets/icons/app.png", 32, 32)
	if e != nil {
		return 0
	}
	maskData := make([]byte, 32*32/8)
	mask, _, _ := createBitmap.Call(32, 32, 1, 1, uintptr(unsafe.Pointer(&maskData[0])))
	if mask == 0 {
		return 0
	}
	defer deleteObject.Call(mask)
	ii := IconInfo{IsIcon: 1, Mask: mask, Color: v.Bitmap}
	icon, _, _ := createIconIndirect.Call(uintptr(unsafe.Pointer(&ii)))
	return icon
}
func (a *App) tray(action int) {
	n := NotifyIconData{}
	n.Size = uint32(unsafe.Sizeof(n))
	n.Window = a.HWND
	n.ID = 1
	n.Flags = 1 | 2 | 4
	n.Callback = WM_TRAY
	n.Icon = a.Icon
	now := wallClockNow()
	tip := a.L.Name() + " · " + a.L.DayShort(a.Data.DayAt(now)) + " · " + a.L.Content(a.Data.HourAt(now).Name) + a.L.Text("\n左键查看，右键设置", "\nClick for details; right-click for settings")
	v, _ := syscall.UTF16FromString(tip)
	copy(n.Tip[:], v)
	ok, _, _ := notifyIcon.Call(u(action), uintptr(unsafe.Pointer(&n)))
	if action == 0 && ok == 0 {
		logError("Shell_NotifyIcon: tray icon is temporarily unavailable")
	}
}
func (a *App) allArrangements() {
	p := filepath.Join(filepath.Dir(settingsPath()), a.L.Text("全部安排.txt", "All guidance.txt"))
	_ = os.MkdirAll(filepath.Dir(p), 0700)
	if e := os.WriteFile(p, append([]byte{0xEF, 0xBB, 0xBF}, []byte(a.L.AllText(a.Data))...), 0600); e != nil {
		errorbox(a.HWND, e.Error())
		return
	}
	openPath(a.HWND, p)
}
func (a *App) copyCurrent() {
	if e := copyText(a.Popup, a.L.CurrentText(a.Data, wallClockNow())); e != nil {
		a.Status = e.Error()
		errorbox(a.Popup, e.Error())
	} else {
		a.Status = a.L.Text("已复制当前七曜、时辰和全部建议", "Copied the current day star, shichen and guidance")
	}
	a.StatusUntil = time.Now().Add(3 * time.Second)
	invalidate(a.Popup)
}
func (a *App) cardClick(p Point) {
	for _, b := range a.Buttons {
		if a.buttonRect(b).Contains(p) && (!b.InBody || a.Body.Contains(p)) {
			if a.Preview.Toggle(b.ID) {
				a.LeaveSince = time.Time{}
				a.positionCard()
				return
			}
			switch b.ID {
			case 1:
				a.Pinned = !a.Pinned
				invalidate(a.Popup)
			case 2:
				a.hideCard()
				a.HoverSince = time.Time{}
			case 3:
				a.copyCurrent()
			case 4:
				a.allArrangements()
			case 5:
				a.Settings.HourDetails = !a.Settings.HourDetails
				a.save()
				a.positionCard()
			case 6:
				a.Settings.DayDetails = !a.Settings.DayDetails
				a.save()
				a.positionCard()
			}
			return
		}
	}
}
func menuItem(m uintptr, id int, text string, checked bool) {
	flags := uintptr(0)
	if checked {
		flags = 8
	}
	appendMenu.Call(m, flags, u(id), uintptr(unsafe.Pointer(w(text))))
}
func menuSep(m uintptr) { appendMenu.Call(m, 0x800, 0, 0) }
func (a *App) contextMenu() {
	if a.MenuOpen {
		return
	}
	a.MenuOpen = true
	if !a.Pinned {
		a.hideCard()
	}
	menu, _, _ := createPopupMenu.Call()
	defer destroyMenu.Call(menu)
	menuItem(menu, 1, a.L.Text("查看当前详情", "Show current details"), false)
	menuItem(menu, 7, a.L.Text("固定悬浮卡", "Pin card"), a.Pinned)
	menuItem(menu, 2, a.L.Text("复制当前建议", "Copy current guidance"), false)
	menuItem(menu, 3, a.L.Text("查看全部安排", "View all guidance"), false)
	menuSep(menu)
	pos, _, _ := createPopupMenu.Call()
	menuItem(pos, 10, a.L.Text("向左移动 10 像素", "Move left by 10 px"), false)
	menuItem(pos, 11, a.L.Text("向右移动 10 像素", "Move right by 10 px"), false)
	if a.Vertical {
		destroyMenu.Call(pos)
		pos, _, _ = createPopupMenu.Call()
		menuItem(pos, 10, a.L.Text("向上移动 10 像素", "Move up by 10 px"), false)
		menuItem(pos, 11, a.L.Text("向下移动 10 像素", "Move down by 10 px"), false)
	}
	menuItem(pos, 12, a.L.Text("恢复默认位置", "Reset position"), false)
	appendMenu.Call(menu, 0x10, pos, uintptr(unsafe.Pointer(w(a.L.Text("位置（也可直接拖动）", "Position (or drag to move)")))))
	fontm, _, _ := createPopupMenu.Call()
	for _, v := range []int{10, 11, 12, 13, 14, 16, 18} {
		menuItem(fontm, 100+v, a.L.PixelLabel(v), a.Settings.FontSize == v)
	}
	appendMenu.Call(menu, 0x10, fontm, uintptr(unsafe.Pointer(w(a.L.Text("任务栏字号", "Taskbar font size")))))
	themem, _, _ := createPopupMenu.Call()
	menuItem(themem, 20, a.L.Text("跟随系统", "Follow system"), a.Settings.Theme == "system")
	menuItem(themem, 21, a.L.Text("深色", "Dark"), a.Settings.Theme == "dark")
	menuItem(themem, 22, a.L.Text("浅色", "Light"), a.Settings.Theme == "light")
	appendMenu.Call(menu, 0x10, themem, uintptr(unsafe.Pointer(w(a.L.Text("外观", "Appearance")))))
	langm, _, _ := createPopupMenu.Call()
	menuItem(langm, 60, "跟随系统 / System", a.Settings.Language == "system")
	menuItem(langm, 61, "简体中文", a.Settings.Language == "zh-CN")
	menuItem(langm, 62, "English", a.Settings.Language == "en")
	appendMenu.Call(menu, 0x10, langm, uintptr(unsafe.Pointer(w("语言 / Language"))))
	menuItem(menu, 30, a.L.Text("显示时辰藏干与含义", "Show shichen stems and meaning"), a.Settings.HourDetails)
	menuItem(menu, 33, a.L.Text("显示曜星更多信息", "Show more day-star details"), a.Settings.DayDetails)
	menuItem(menu, 31, a.L.Text("全屏时隐藏", "Hide during fullscreen"), a.Settings.HideFullscreen)
	auto := autostartEnabled(a.Exe)
	menuItem(menu, 32, a.L.Text("开机启动（当前用户）", "Start at sign-in (current user)"), auto)
	menuSep(menu)
	menuItem(menu, 40, a.L.Text("打开原始表格文件夹", "Open workbook folder"), false)
	menuItem(menu, 41, a.L.Text("重新读取表格", "Reload workbooks and translations"), false)
	menuItem(menu, 42, a.L.Text("打开设置文件夹", "Open settings folder"), false)
	menuItem(menu, 43, a.L.Text("五行配色规则", "Five-element color palette"), false)
	menuSep(menu)
	menuItem(menu, 50, a.L.Text("关于与使用说明", "About and help"), false)
	menuItem(menu, 99, a.L.Text("退出", "Exit"), false)
	previous, _, _ := getForegroundWindow.Call()
	setForegroundWindow.Call(a.HWND)
	p := cursorPoint()
	cmd, _, _ := trackPopupMenu.Call(menu, 0x100|0x80|0x2, u(int(p.X)), u(int(p.Y)), 0, a.HWND, 0)
	postMessage.Call(a.HWND, 0, 0, 0)
	if previous != 0 && previous != a.HWND {
		setForegroundWindow.Call(previous)
	}
	a.MenuOpen = false
	a.HoverSince = time.Now()
	switch int(cmd) {
	case 1:
		a.showCard(true)
	case 7:
		if a.Pinned {
			a.Pinned = false
			a.LeaveSince = time.Now()
		} else {
			a.showCard(true)
		}
	case 2:
		a.copyCurrent()
	case 3:
		a.allArrangements()
	case 10:
		a.Settings.Gap += 10
		a.save()
		a.positionWidget()
	case 11:
		a.Settings.Gap -= 10
		a.save()
		a.positionWidget()
	case 12:
		a.Settings.Gap = DefaultSettings().Gap
		a.save()
		a.positionWidget()
	case 20, 21, 22:
		a.Settings.Theme = map[int]string{20: "system", 21: "dark", 22: "light"}[int(cmd)]
		a.theme()
		a.save()
		invalidate(a.HWND)
		if a.CardShown {
			a.positionCard()
		}
	case 60, 61, 62:
		a.setLanguage(map[int]string{60: "system", 61: "zh-CN", 62: "en"}[int(cmd)])
	case 30, 33:
		if cmd == 30 {
			a.Settings.HourDetails = !a.Settings.HourDetails
		} else {
			a.Settings.DayDetails = !a.Settings.DayDetails
		}
		a.save()
		if a.CardShown {
			a.positionCard()
		}
	case 31:
		a.Settings.HideFullscreen = !a.Settings.HideFullscreen
		a.save()
		a.positionWidget()
	case 32:
		if e := setAutostart(a.Exe, !auto); e != nil {
			errorbox(a.HWND, e.Error())
		} else {
			a.Settings.Autostart = !auto
			a.save()
		}
	case 40:
		openPath(a.HWND, filepath.Join(a.Base, "表格"))
	case 41:
		a.reload(true)
	case 42:
		_ = os.MkdirAll(filepath.Dir(settingsPath()), 0700)
		openPath(a.HWND, filepath.Dir(settingsPath()))
	case 43:
		msgbox(a.HWND, a.L.Text("五行配色", "Five-element palette"), a.L.ColorRules())
	case 50:
		msgbox(a.HWND, a.L.Title(), a.L.About(a.Source))
	case 99:
		destroyWindow.Call(a.HWND)
	default:
		if cmd >= 110 && cmd <= 118 {
			a.Settings.FontSize = int(cmd) - 100
			a.save()
			invalidate(a.HWND)
			a.positionWidget()
		}
	}
}
func widgetProc(h uintptr, msg uint32, wp, lp uintptr) (ret uintptr) {
	defer func() {
		if e := recover(); e != nil {
			logError(fmt.Sprintf("widget message %x: %v", msg, e))
			ret, _, _ = defWindowProc.Call(h, u(int(msg)), wp, lp)
		}
	}()
	a := app
	if a == nil {
		r, _, _ := defWindowProc.Call(h, u(int(msg)), wp, lp)
		return r
	}
	if a.TaskbarCreated != 0 && msg == a.TaskbarCreated {
		a.tray(0)
		a.positionWidget()
		return 0
	}
	switch msg {
	case WM_RENDER:
		a.render(h, false)
		return 0
	case WM_PAINT:
		a.paint(h, false)
		return 0
	case WM_ERASEBKGND:
		return 1
	case WM_MOUSEACTIVATE:
		return 3 // MA_NOACTIVATE: hovering/clicking the widget does not activate it.
	case WM_TIMER:
		a.tick()
		return 0
	case WM_LBUTTONDOWN:
		a.MouseDown = true
		a.Dragging = false
		a.DragStart = cursorPoint()
		a.DragGap = a.Settings.Gap
		setCapture.Call(h)
		return 0
	case WM_MOUSEMOVE:
		if a.MouseDown {
			p := cursorPoint()
			delta := int(p.X - a.DragStart.X)
			if a.Vertical {
				delta = int(p.Y - a.DragStart.Y)
			}
			if abs(delta) > a.S(3) {
				a.Dragging = true
				a.hideCard()
			}
			if a.Dragging {
				a.Settings.Gap = a.DragGap - delta*96/a.DPI
				a.Settings.Normalize()
				a.positionWidget()
			}
		}
		return 0
	case WM_LBUTTONUP:
		if a.MouseDown {
			dragged := a.Dragging
			a.MouseDown = false
			a.Dragging = false
			releaseCapture.Call()
			if dragged {
				a.save()
			} else if a.CardShown && (a.Pinned || a.Preview.Active()) {
				a.hideCard()
				a.HoverSince = time.Now().Add(24 * time.Hour)
			} else {
				a.showCard(true)
			}
		}
		return 0
	case WM_CAPTURECHANGED:
		if a.MouseDown {
			a.MouseDown = false
			a.Dragging = false
			a.save()
		}
		return 0
	case WM_RBUTTONUP:
		a.contextMenu()
		return 0
	case WM_MOUSEWHEEL:
		delta := int(int16((wp >> 16) & 0xffff))
		if delta > 0 {
			a.Settings.Gap += 5
		} else if delta < 0 {
			a.Settings.Gap -= 5
		}
		a.save()
		a.positionWidget()
		return 0
	case WM_TRAY:
		switch uint32(lp) & 0xffff {
		case WM_LBUTTONUP:
			a.showCard(true)
		case WM_RBUTTONUP:
			a.contextMenu()
		}
		return 0
	case WM_SETLANG:
		if wp <= 2 {
			a.setLanguage([]string{"system", "zh-CN", "en"}[wp])
		}
		return 0
	case WM_SHOWCARD:
		a.positionWidget()
		a.showCard(true)
		return 0
	case WM_RESETPOS:
		a.Settings.Gap = DefaultSettings().Gap
		a.save()
		a.positionWidget()
		a.showCard(true)
		return 0
	case WM_DPICHANGED, WM_DISPLAYCHANGE, WM_SETTINGCHANGE:
		if a.Settings.Language == "system" && a.L.Language != ResolveLanguage("system", systemLanguage()) {
			a.setLanguage("system")
		}
		a.theme()
		a.positionWidget()
		invalidate(a.HWND)
		if a.CardShown {
			a.positionCard()
		}
		return 0
	case WM_CLOSE:
		destroyWindow.Call(h)
		return 0
	case WM_DESTROY:
		a.Closing = true
		killTimer.Call(h, 1)
		a.save()
		a.tray(2)
		if a.Popup != 0 {
			destroyWindow.Call(a.Popup)
		}
		a.clearFonts()
		if a.Icon != 0 {
			destroyIcon.Call(a.Icon)
		}
		postQuitMessage.Call(0)
		return 0
	}
	r, _, _ := defWindowProc.Call(h, u(int(msg)), wp, lp)
	return r
}
func cardProc(h uintptr, msg uint32, wp, lp uintptr) (ret uintptr) {
	defer func() {
		if e := recover(); e != nil {
			logError(fmt.Sprintf("card message %x: %v", msg, e))
			ret, _, _ = defWindowProc.Call(h, u(int(msg)), wp, lp)
		}
	}()
	a := app
	if a == nil {
		r, _, _ := defWindowProc.Call(h, u(int(msg)), wp, lp)
		return r
	}
	switch msg {
	case 0x20: // WM_SETCURSOR
		if a.previewControlCursor() {
			return 1
		}
	case WM_RENDER:
		a.render(h, true)
		return 0
	case WM_PAINT:
		a.paint(h, true)
		return 0
	case WM_ERASEBKGND:
		return 1
	case WM_MOUSEACTIVATE:
		return 3
	case WM_LBUTTONUP:
		a.cardClick(Point{int32(int16(lp & 0xffff)), int32(int16((lp >> 16) & 0xffff))})
		return 0
	case WM_MOUSEMOVE:
		p := Point{int32(int16(lp & 0xffff)), int32(int16((lp >> 16) & 0xffff))}
		hovered := 0
		for _, b := range a.Buttons {
			if a.Body.Contains(p) && a.buttonRect(b).Contains(p) {
				hovered = b.ID
				break
			}
		}
		if hovered != a.CardHoverID || !a.MouseOverCard {
			a.CardHoverID = hovered
			a.MouseOverCard = true
			invalidate(h)
		}
		return 0
	case WM_MOUSEWHEEL:
		delta := int(int16((wp >> 16) & 0xffff))
		a.Scroll = clamp(a.Scroll-delta*a.S(42)/120, 0, max(0, a.ContentHeight-a.Body.H()))
		invalidate(h)
		return 0
	case WM_RBUTTONUP:
		a.contextMenu()
		return 0
	case WM_CLOSE:
		a.hideCard()
		return 0
	}
	r, _, _ := defWindowProc.Call(h, u(int(msg)), wp, lp)
	return r
}
func register(name string, callback uintptr, instance, icon, cursor uintptr, shadow bool) error {
	wc := WindowClass{WndProc: callback, Instance: instance, Icon: icon, IconSmall: icon, Cursor: cursor, ClassName: w(name)}
	wc.Size = uint32(unsafe.Sizeof(wc))
	if shadow {
		wc.Style = 0x20000
	}
	r, _, e := registerClass.Call(uintptr(unsafe.Pointer(&wc)))
	if r == 0 {
		return fmt.Errorf("%s: %w", activeLocale().Text("注册窗口失败", "Cannot register window"), e)
	}
	return nil
}
func selfTest(exeDir string, l Localizer) {
	d, source, e := LoadData(exeDir)
	lines := []string{l.Name() + " " + appVersion, l.Text("数据来源：", "Data source: ") + l.Content(source)}
	if e != nil {
		lines = append(lines, l.Text("读取错误：", "Read error: ")+e.Error())
	}
	if err := d.Validate(); err != nil {
		lines = append(lines, l.Text("数据校验失败：", "Data validation failed: ")+err.Error())
	} else {
		lines = append(lines, l.Text("7个曜星、12个时辰数据校验通过", "Data valid: 7 day stars, 12 shichen"))
	}
	ok := true
	for i := 0; i < 1440; i++ {
		t := time.Date(2026, 9, 9, i/60, i%60, 0, 0, time.UTC)
		if HourIndex(t) != ((i+60)/120)%12 {
			ok = false
		}
	}
	lines = append(lines, fmt.Sprintf(l.Text("全天1440分钟时辰映射：%t", "All 1,440 minute mappings: %t"), ok))
	lines = append(lines, l.Text("语言：", "Language: ")+l.Language)
	p := filepath.Join(exeDir, "self-test.txt")
	if err := os.WriteFile(p, append([]byte{0xEF, 0xBB, 0xBF}, []byte(strings.Join(lines, "\r\n"))...), 0600); err != nil {
		errorbox(0, err.Error())
	}
}
func main() {
	runtime.LockOSThread()
	defer func() {
		if e := recover(); e != nil {
			logError(e)
			errorbox(0, fmt.Sprintf(activeLocale().Text("程序出现错误：%v\n日志位于 %%APPDATA%%\\Astral\\error.log", "Unexpected error: %v\nLog: %%APPDATA%%\\Astral\\error.log"), e))
		}
	}()
	exe, e := os.Executable()
	if e != nil {
		errorbox(0, e.Error())
		return
	}
	exe, _ = filepath.Abs(exe)
	base := filepath.Dir(exe)
	settings := ReadSettings()
	languageRequested := false
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--language=") {
			value := strings.TrimPrefix(arg, "--language=")
			if value != "system" && value != "zh-CN" && value != "en" {
				errorbox(0, "Language must be system, zh-CN or en.")
				return
			}
			settings.Language = value
			languageRequested = true
		}
	}
	startupLocale = loadLocalizer(settings.Language, base)
	for _, arg := range os.Args[1:] {
		if arg == "--self-test" {
			selfTest(base, startupLocale)
			return
		}
	}
	reset := false
	preview := false
	for _, arg := range os.Args[1:] {
		if arg == "--preview" {
			preview = true
		}
		if arg == "--reset-position" {
			reset = true
		}
	}
	mutex, _, err := createMutex.Call(0, 0, uintptr(unsafe.Pointer(w(`Local\QiyaoShichen.v1`))))
	if mutex != 0 {
		defer closeHandle.Call(mutex)
	}
	if err == syscall.Errno(183) {
		h, _, _ := findWindow.Call(uintptr(unsafe.Pointer(w(widgetClass))), 0)
		if h != 0 {
			if !isCurrentWindowTitle(windowTitle(h)) {
				msgbox(0, startupLocale.Text("请先退出旧版", "Close the previous version"), startupLocale.Text("检测到旧版七曜时辰仍在运行。\n请右键旧版控件 → 退出，再启动七曜工作法。", "An older version of this app is still running.\nRight-click its widget and choose Exit before starting Astral Rhythm."))
				return
			}
			if languageRequested {
				postMessage.Call(h, WM_SETLANG, u(map[string]int{"system": 0, "zh-CN": 1, "en": 2}[settings.Language]), 0)
			}
			m := WM_SHOWCARD
			if reset {
				m = WM_RESETPOS
			}
			postMessage.Call(h, u(m), 0, 0)
		}
		return
	}
	enableDPI()
	data, source, loadErr := LoadData(base)
	if loadErr != nil {
		logError(loadErr)
	}
	if reset {
		settings.Gap = DefaultSettings().Gap
	}
	app = &App{L: startupLocale, Exe: exe, Base: base, Data: data, Source: source, Settings: settings, DPI: 96, Fonts: map[string]uintptr{}}
	a := app
	a.Monitor = monitorInfo(0)
	a.theme()
	a.Icon = a.createTrayIcon()
	a.LastStamp = a.tableStamp()
	inst, _, _ := getModuleHandle.Call(0)
	hand, _, _ := loadCursor.Call(0, 32649)
	arrow, _, _ := loadCursor.Call(0, 32512)
	if e = register(widgetClass, syscall.NewCallback(widgetProc), inst, a.Icon, hand, false); e != nil {
		errorbox(0, e.Error())
		return
	}
	if e = register(cardClass, syscall.NewCallback(cardProc), inst, a.Icon, arrow, false); e != nil {
		errorbox(0, e.Error())
		return
	}
	style := uintptr(WS_EX_TOOLWINDOW | WS_EX_NOACTIVATE | WS_EX_TOPMOST | WS_EX_LAYERED)
	a.HWND, _, e = createWindow.Call(style, uintptr(unsafe.Pointer(w(widgetClass))), uintptr(unsafe.Pointer(w(a.L.Title()))), WS_POPUP, 0, 0, 52, 38, 0, 0, inst, 0)
	if a.HWND == 0 {
		errorbox(0, a.L.Text("创建控件失败：", "Cannot create widget: ")+e.Error())
		return
	}
	a.Popup, _, e = createWindow.Call(style, uintptr(unsafe.Pointer(w(cardClass))), uintptr(unsafe.Pointer(w(a.L.Name()+a.L.Text(" · 当前建议", " · Current guidance")))), WS_POPUP, 0, 0, 410, 580, a.HWND, 0, inst, 0)
	if a.Popup == 0 {
		errorbox(0, a.L.Text("创建提示卡失败：", "Cannot create card: ")+e.Error())
		return
	}
	t, _, _ := registerWindowMessage.Call(uintptr(unsafe.Pointer(w("TaskbarCreated"))))
	a.TaskbarCreated = uint32(t)
	if a.Settings.Autostart {
		if e := setAutostart(a.Exe, true); e != nil {
			logError(e)
		}
	}
	a.tray(0)
	a.positionWidget()
	a.save()
	setTimer.Call(a.HWND, 1, 200, 0)
	if preview {
		a.showCard(true)
	}
	if loadErr != nil {
		msgbox(a.HWND, a.L.Text("表格读取提示", "Workbook loading"), a.L.Text("原始表格暂时无法读取，当前使用内置数据。\n\n", "Cannot read workbooks; using built-in data.\n\n")+loadErr.Error()+a.L.Text("\n\n右键控件 → 重新读取表格，可再次尝试。", "\n\nRight-click the widget and choose Reload workbooks to retry."))
	}
	var msg Message
	for {
		r, _, err := getMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(r) == -1 {
			logError(err)
			break
		}
		if r == 0 {
			break
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}
