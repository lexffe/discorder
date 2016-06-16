package ui

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/termbox-go"
)

// Unity3d like UI transform (minus scale, pivot and rotation)
type Transform struct {
	AnchorMin common.Vector2F
	AnchorMax common.Vector2F

	Position common.Vector2F
	Size     common.Vector2F

	Top, Bottom, Left, Right int

	Parent   *Transform
	Children []Entity
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
		ret.Y = yOffsetMin + parentRect.Y + float32(t.Top)
		ret.H = yOffsetMax - yOffsetMin - float32(t.Bottom+t.Top)
	}

	if t.AnchorMax.X == t.AnchorMin.X {
		ret.X = t.Position.X + parentRect.X + (t.AnchorMin.X * parentRect.W)
		ret.W = t.Size.X
	} else {
		xOffsetMin := parentRect.W * t.AnchorMin.X
		xOffsetMax := parentRect.W * t.AnchorMax.X
		ret.X = xOffsetMin + parentRect.X + float32(t.Left)
		ret.W = xOffsetMax - xOffsetMin - (float32(t.Right) + float32(t.Left))
	}
	return ret
}

func (t *Transform) AddChildren(children ...Entity) {
	for _, v := range children {
		childTransform := v.GetTransform()
		childTransform.Parent = t
	}
	if t.Children == nil {
		t.Children = make([]Entity, len(children))
		copy(t.Children, children)
	} else {
		t.Children = append(t.Children, children...)
	}
}

func (t *Transform) AddFirst(child Entity) {
	child.GetTransform().Parent = t
	if t.Children == nil {
		t.Children = []Entity{child}
	} else {
		t.Children = append([]Entity{child}, t.Children...)
	}
}

func (t *Transform) RemoveChild(child Entity, destroy bool) {
	if t.Children == nil || len(t.Children) < 1 {
		return
	}

	if destroy {
		child.Destroy()
	}

	for k, v := range t.Children {
		if v == child {
			t.Children = append(t.Children[:k], t.Children[k+1:]...)
			break
		}

	}
}

// Revmoves and optinally clears children
func (t *Transform) ClearChildren(destroy bool) {
	for _, v := range t.Children {
		if destroy {
			v.Destroy()
		}
		v.GetTransform().Parent = nil
	}
	t.Children = []Entity{}
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
