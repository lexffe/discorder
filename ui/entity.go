package ui

import (
	"github.com/nsf/termbox-go"
)

type Entity interface {
	Children(recursive bool) []Entity
	Destroy()
	GetTransform() *Transform
}

type BaseEntity struct {
	Transform Transform
}

// Runs f recursively and in order on e and its children
// by in order it means it runs on parent first and children last
func RunFunc(e Entity, f func(e Entity)) {
	f(e)
	children := e.Children(false) // We wanna make sure we do it in the proper order
	if children != nil {
		for _, child := range children {
			RunFunc(child, f)
		}
	}
}

// Same as above but only runs on children if f returns true
func RunFuncConditional(e Entity, f func(e Entity) bool) {
	traverseChildren := f(e)
	if traverseChildren {
		children := e.Children(false) // We wanna make sure we do it in the proper order
		if children != nil {
			for _, child := range children {
				RunFuncConditional(child, f)
			}
		}
	}
}

func (b *BaseEntity) GetTransform() *Transform {
	return &b.Transform
}

// Maybe reuse the slice...? probably miniscule performance hit to not...
func (b *BaseEntity) Children(recursive bool) []Entity {
	if b.Transform.Children == nil || len(b.Transform.Children) < 1 {
		return nil
	}

	ret := make([]Entity, len(b.Transform.Children))
	copy(ret, b.Transform.Children)

	if recursive {
		for _, entity := range b.Transform.Children {
			children := entity.Children(true)
			if children != nil {
				ret = append(ret, children...)
			}
		}
	}

	return ret
}

func (b *BaseEntity) DestroyChildren() {
	b.Transform.ClearChildren(true)
}

type SimpleEntity struct {
	*BaseEntity
}

func NewSimpleEntity() *SimpleEntity {
	return &SimpleEntity{
		BaseEntity: &BaseEntity{},
	}
}

func (s *SimpleEntity) Destroy() { s.DestroyChildren() }

// Misc handlers

type InputHandler interface {
	HandleInput(event termbox.Event)
}

type UpdateHandler interface {
	Update() // Ran before drawing, for example add or remove children
}

type LateUpdateHandler interface {
	LateUpdate() // Ran after update, shouldnt change the size of the element
}

type DrawHandler interface {
	GetDrawLayer() int
	Draw()
}

type Scrollable interface {
	Scroll(dir Direction, amount int)
}
