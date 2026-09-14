//go:build windows

package main

import "time"

// The UI uses a native, per-pixel alpha surface for both windows. The taskbar
// button has no visual plate; the two separate cards reproduce the reference.
func (a *App) paintWidget(dst *PixelBuffer) {
	a.paintWidgetAt(dst, wallClockNow())
}

// A shared clock input makes native documentation renders reproducible.
func (a *App) paintWidgetAt(dst *PixelBuffer, now time.Time) {
	day, hour := a.Data.DayAt(now), a.Data.HourAt(now)
	light := a.TaskbarLight
	// An alpha of 1/255 keeps the entire label pair clickable/draggable under
	// Win32 hit testing. It is not a drawn plate: no tint, border or shadow.
	// Fully transparent alpha=0 would make the gaps between letters click-through.
	for i := 3; i < len(dst.Pix); i += 4 {
		dst.Pix[i] = 1
	}
	r := dst.Bounds()
	half := r.H() / 2
	sz := min(a.Settings.FontSize, max(8, (half-2)*96/a.DPI))
	isz := min(a.S(17), half-a.S(2))
	x := a.S(3)
	tx := x + isz + a.S(6)
	if r.W() < a.S(61) {
		x = a.S(2)
		isz = min(isz, a.S(13))
		tx = x + isz + a.S(3)
		sz = min(sz, max(8, (r.W()-tx-a.S(2))*96/(2*a.DPI)))
	}
	canvasArt(dst, assetName("day", day.Weekday, light, false), R(x, (half-isz)/2, isz, isz), r)
	canvasArt(dst, assetName("hour", hour.Index, light, false), R(x, half+(r.H()-half-isz)/2, isz, isz), r)
	a.canvasText(dst, a.L.DayShort(day), R(tx, 0, r.W()-tx-a.S(2), half), sz, true, DayTextColor(day.Weekday, light), 0, r)
	a.canvasText(dst, a.L.Content(hour.Name), R(tx, half, r.W()-tx-a.S(2), r.H()-half), sz, true, HourTextColor(hour.Index, light), 0, r)
}
func (a *App) makeLayout() {
	now := wallClockNow()
	dc, _, _ := getDC.Call(a.Popup)
	defer releaseDC.Call(a.Popup, dc)
	o := CardOptions{Locale: a.L, DPI: a.DPI, MaxWidth: max(a.S(240), a.Monitor.Work.W()-a.S(20)), MaxHeight: max(a.S(160), a.Monitor.Work.H()-a.S(24)), Light: a.isLight(), DayDetails: a.Settings.DayDetails, HourDetails: a.Settings.HourDetails}
	a.applyPreviewOptions(&o, now)
	layout := CompactLayout(a.Data.DayAt(now), a.Data.HourAt(now), o, func(s string, size int, bold bool) int { return a.textWidth(dc, s, size, bold) })
	a.CardWidth, a.CardHeight, a.ContentHeight = layout.Width, layout.Height, layout.ContentHeight
	a.Blocks, a.Panels, a.DrawIcons, a.Buttons, a.Body = layout.Blocks, layout.Panels, layout.Icons, layout.Buttons, layout.Body
	a.Scroll = clamp(a.Scroll, 0, max(0, a.ContentHeight-a.Body.H()))
}
func (a *App) shifted(r Rect) Rect {
	r.Top += a.Body.Top - int32(a.Scroll)
	r.Bottom += a.Body.Top - int32(a.Scroll)
	return r
}
func (a *App) buttonRect(b Button) Rect {
	if b.InBody {
		return a.shifted(b.Rect)
	}
	return b.Rect
}
func (a *App) paintCard(dst *PixelBuffer) {
	r := dst.Bounds()
	c := a.Colors
	light := a.isLight()
	clip := intersectRect(a.Body, R(1, 1, r.W()-2, r.H()-2))
	dst.RoundRect(r, float64(a.S(14)), c.Panel, r)
	for _, p := range a.Panels {
		rr := a.shifted(p.Rect)
		if rr.Bottom <= a.Body.Top || rr.Top >= a.Body.Bottom {
			continue
		}
		if previewPanelKind(p.Kind) {
			a.paintPreviewPanel(dst, p, rr, clip)
			continue
		}
		switch p.Kind {
		case "day", "hour":
			v := visualDesign.Days[0]
			if p.Kind == "day" {
				v = visualDesign.Days[p.Index]
			} else {
				v = visualDesign.Hours[p.Index]
			}
			border := blendRGB(v.Accent(light), v.Surface(light), .87)
			canvasPanel(dst, assetName(p.Kind, p.Index, light, true), rr, a.S(11), v.Surface(light), border, clip)
		case "day-badge":
			v := visualDesign.Days[p.Index]
			col := blendRGB(v.Accent(light), v.Surface(light), .89)
			dst.RoundRect(rr, float64(a.S(8)), col, clip)
		}
	}
	for _, v := range a.DrawIcons {
		rr := a.shifted(v.Rect)
		if rr.Bottom > a.Body.Top && rr.Top < a.Body.Bottom {
			canvasArt(dst, v.Path, rr, clip)
		}
	}
	for _, b := range a.Blocks {
		rr := a.shifted(b.Rect)
		if rr.Bottom > a.Body.Top && rr.Top < a.Body.Bottom {
			a.canvasText(dst, b.Text, rr, b.Size, b.Bold, b.Color, 0, clip)
		}
	}
	a.paintButtons(dst, clip)
	if a.ContentHeight > a.Body.H() {
		track := a.Body.H() - a.S(20)
		thumb := max(a.S(26), track*a.Body.H()/a.ContentHeight)
		top := a.S(10) + (track-thumb)*a.Scroll/(a.ContentHeight-a.Body.H())
		dst.RoundRect(R(r.W()-a.S(5), top, a.S(2), thumb), float64(a.S(1)), c.Muted, r)
	}
	dst.RoundFrame(r, float64(a.S(14)), 1, c.Border, r)
}
func (a *App) paintButtons(dst *PixelBuffer, clip Rect) {
	a.paintCardControls(dst, clip)
}
