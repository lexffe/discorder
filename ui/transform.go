package ui

import (
	"github.com/jonas747/discorder/common"
	"github.com/nsf/termbox-go"
)

// Unit3d like UI transform (minus scale, pivot and rotation)
type Transform struct {
	AnchorMin common.Vector2F
	AnchorMax common.Vector2F

	Position common.Vector2F
	Size     common.Vector2F

	Top, Bottom, Left, Right int

	Parent *Transform
}

// Incorrect.. will fix as i come by the silly mistakes
func (t *Transform) GetRect() common.Rect {
	parentRect := common.Rect{}
	if t.Parent != nil {
		parentRect = t.Parent.GetRect()
	} else {
		termSizeX, termSizeY := termbox.Size()
		parentRect = common.Rect{0, 0, float32(termSizeX), float32(termSizeY)}
	}

	ret := common.Rect{}

	if t.AnchorMax.Y == t.AnchorMin.Y {
		ret.Y = t.Position.Y + parentRect.Y + (t.AnchorMin.Y * parentRect.H)
		ret.H = t.Size.Y
	} else {
		yOffsetMin := parentRect.H * t.AnchorMin.Y
		yOffsetMax := parentRect.H * t.AnchorMax.Y
		ret.Y = yOffsetMax + parentRect.Y + float32(t.Top)
		ret.H = yOffsetMax - yOffsetMin - float32(t.Bottom)
	}

	if t.AnchorMax.X == t.AnchorMin.X {
		ret.X = t.Position.X + parentRect.X + (t.AnchorMin.X * parentRect.W)
		ret.W = t.Size.X
	} else {
		xOffsetMin := parentRect.W * t.AnchorMin.X
		xOffsetMax := parentRect.W * t.AnchorMax.X
		ret.X = xOffsetMin + parentRect.X + float32(t.Left)
		ret.W = xOffsetMax - xOffsetMin - float32(t.Right)
	}

	return ret
}

/*

Anchor min&max
|
V
 ________
|		 |
|		 |
|________|

w = 4
h = 3

x = 0
y = 0

pivot = top left



 globalviewentity
 	|		|
 texview	helpentity
 			|
 			helpwindow


*/
