package main

type Point struct{ X, Y int32 }
type Rect struct{ Left, Top, Right, Bottom int32 }

func (r Rect) W() int { return int(r.Right - r.Left) }
func (r Rect) H() int { return int(r.Bottom - r.Top) }
func (r Rect) Contains(p Point) bool {
	return p.X >= r.Left && p.X < r.Right && p.Y >= r.Top && p.Y < r.Bottom
}
func R(x, y, w, h int) Rect { return Rect{int32(x), int32(y), int32(x + w), int32(y + h)} }

func clamp(n, lo, hi int) int {
	if hi < lo {
		return lo
	}
	if n < lo {
		return lo
	}
	if n > hi {
		return hi
	}
	return n
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
