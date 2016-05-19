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
	Transform                           *Transform
	ForceExpandWidth, ForceExpandHeight bool
	LayoutType                          LayoutType
}

func NewAutoLayoutContainer() *AutoLayoutContainer {
	return &AutoLayoutContainer{
		BaseEntity: &BaseEntity{},
		Transform:  &Transform{},
	}
}

func (a *AutoLayoutContainer) BuildLayout() {

	rect := a.Transform.GetRect()

	required := float32(0)
	numDynammic := 0
	elements := make([]LayoutElement, 0)
	RunFuncConditional(a, func(e Entity) bool {
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

		if a.LayoutType == LayoutTypeVertical {
			transform.AnchorMin.Y = 0
			transform.AnchorMax.Y = 0
			required += requiredSize.Y
			if requiredSize.Y == 0 {
				numDynammic++
			}
		} else {
			transform.AnchorMin.X = 0
			transform.AnchorMin.X = 0
			required += requiredSize.X
			if requiredSize.X == 0 {
				numDynammic++
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
	for _, v := range elements {
		requiredSize := v.GetRequiredSize()
		transform := v.GetTransform()

		if a.LayoutType == LayoutTypeVertical {
			transform.Position = common.NewVector2F(0, counter)
			if requiredSize.Y == 0 {
				transform.Size.Y = spacePerDynamic
				counter += spacePerDynamic
			} else {
				transform.Size.Y = requiredSize.Y
				counter += requiredSize.Y
			}
		} else {
			transform.Position = common.NewVector2F(counter, 0)
			if requiredSize.X == 0 {
				transform.Size.X = spacePerDynamic
				counter += spacePerDynamic
			} else {
				transform.Size.X = requiredSize.X
				counter += requiredSize.X
			}
		}
	}
}

func (a *AutoLayoutContainer) Update() {
	a.BuildLayout()
}

func (a *AutoLayoutContainer) Destroy() { a.DestroyChildren() }

type LayoutElement interface {
	GetRequiredSize() common.Vector2F
	//PreferredSize() common.Vector2F
	GetTransform() *Transform
}

type Container struct {
	*BaseEntity
	Transform *Transform
	ProxySize LayoutElement
}

// A bare bones container
func NewContainer() *Container {
	return &Container{
		BaseEntity: &BaseEntity{},
		Transform:  &Transform{},
	}
}

func (c *Container) GetRequiredSize() common.Vector2F {
	if c.ProxySize != nil {
		return c.ProxySize.GetRequiredSize()
	}

	rect := c.Transform.GetRect()
	return common.NewVector2F(rect.W, rect.H)
}

func (c *Container) GetTransform() *Transform {
	return c.Transform
}

func (c *Container) Destroy() { c.DestroyChildren() }
