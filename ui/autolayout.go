package ui

import (
	"github.com/jonas747/discorder/common"
)

type LayoutType int

const (
	LayoutTypeVertical LayoutType = iota
	LayoutTypeHorizontal
)

type AutoLayoutContainer struct {
	*BaseEntity
	ForceExpandWidth, ForceExpandHeight bool
	LayoutType                          LayoutType
	LayoutDynamic                       bool
	Spacing                             int
}

func NewAutoLayoutContainer() *AutoLayoutContainer {
	return &AutoLayoutContainer{
		BaseEntity: &BaseEntity{},
	}
}

func (a *AutoLayoutContainer) BuildLayout() {

	rect := a.Transform.GetRect()

	required := float32(0)
	numDynammic := 0
	elements := make([]LayoutElement, 0)
	// Get number of dynamic elements and calulate leftover space for them
	RunFuncCondTraverse(a, func(e Entity) bool {
		if e == a {
			return true
		}
		cast, ok := e.(LayoutElement)
		if !ok {
			return false
		}
		transform := cast.GetTransform()

		if a.LayoutType == LayoutTypeVertical && a.ForceExpandWidth {
			transform.Size.X = rect.W
		} else if a.LayoutType == LayoutTypeHorizontal && a.ForceExpandHeight {
			transform.Size.Y = rect.H
		}

		requiredSize := cast.GetRequiredSize()
		dynamic := cast.IsLayoutDynamic()
		if dynamic {
			numDynammic++
		}

		if a.LayoutType == LayoutTypeVertical {
			transform.AnchorMin.Y = 0
			transform.AnchorMax.Y = 0
			if !dynamic {
				required += requiredSize.Y
			}
		} else {
			transform.AnchorMin.X = 0
			transform.AnchorMax.X = 0
			if !dynamic {
				required += requiredSize.X
			}
		}

		elements = append(elements, cast)
		return false
	})

	spaceLeft := float32(0)
	if a.LayoutType == LayoutTypeVertical {
		spaceLeft = rect.H - required
	} else {
		spaceLeft = rect.W - required
	}

	spacePerDynamic := spaceLeft / float32(numDynammic)

	counter := float32(0)
	// Apply
	for _, v := range elements {
		requiredSize := v.GetRequiredSize()
		transform := v.GetTransform()

		if a.LayoutType == LayoutTypeVertical {
			transform.Position = common.NewVector2F(transform.Position.X, counter)
			if v.IsLayoutDynamic() {
				transform.Size.Y = spacePerDynamic
				counter += spacePerDynamic + float32(a.Spacing)
			} else {
				transform.Size.Y = requiredSize.Y
				counter += requiredSize.Y + float32(a.Spacing)
			}
		} else {
			transform.Position = common.NewVector2F(counter, transform.Position.Y)
			if v.IsLayoutDynamic() {
				transform.Size.X = spacePerDynamic
				counter += spacePerDynamic + float32(a.Spacing)
			} else {
				transform.Size.X = requiredSize.X
				counter += requiredSize.X + float32(a.Spacing)
			}
		}
	}
}

func (a *AutoLayoutContainer) Update() {
	a.BuildLayout()
}

func (a *AutoLayoutContainer) Destroy() { a.DestroyChildren() }

func (a *AutoLayoutContainer) GetRequiredSize() common.Vector2F {
	rect := a.Transform.GetRect()
	return common.NewVector2F(rect.W, rect.H)
}

func (a *AutoLayoutContainer) IsLayoutDynamic() bool {
	return a.LayoutDynamic
}

type LayoutElement interface {
	GetRequiredSize() common.Vector2F
	GetTransform() *Transform
	IsLayoutDynamic() bool
}

type Container struct {
	*BaseEntity
	ProxySize     LayoutElement
	Dynamic       bool
	AllowZeroSize bool
}

// A bare bones container
func NewContainer() *Container {
	return &Container{
		BaseEntity: &BaseEntity{},
	}
}

func (c *Container) GetRequiredSize() common.Vector2F {
	if c.ProxySize != nil {
		size := c.ProxySize.GetRequiredSize()
		if !c.AllowZeroSize {
			if size.X == 0 {
				size.X = 1
			} else if size.Y == 0 {
				size.Y = 1
			}
		}
		return size
	}

	if c.Dynamic {
		return common.NewVector2I(0, 0)
	}

	rect := c.Transform.GetRect()
	return common.NewVector2F(rect.W, rect.H)
}

func (c *Container) IsLayoutDynamic() bool {
	return c.Dynamic
}

func (c *Container) Destroy() { c.DestroyChildren() }
