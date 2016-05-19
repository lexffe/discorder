package ui

import (
	"github.com/nsf/termbox-go"
)

type Entity interface {
	Children(recursive bool) []Entity
	Destroy()
}

type BaseEntity struct {
	entities []Entity
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

// Maybe reuse the slice...? probably miniscule performance hit to not...
func (b *BaseEntity) Children(recursive bool) []Entity {
	if b.entities == nil || len(b.entities) < 1 {
		return nil
	}

	ret := make([]Entity, len(b.entities))
	copy(ret, b.entities)
	if recursive {
		for _, entity := range b.entities {
			children := entity.Children(true)
			if children != nil {
				ret = append(ret, children...)
			}
		}
	}

	return ret
}

func (b *BaseEntity) AddChild(children ...Entity) {

	if b.entities == nil {
		b.entities = make([]Entity, len(children))
		copy(b.entities, children)
	} else {
		b.entities = append(b.entities, children...)
	}
}

func (b *BaseEntity) RemoveChild(child Entity, destroy bool) {
	if b.entities == nil || len(b.entities) < 1 {
		return
	}

	if destroy {
		child.Destroy()
	}

	for k, v := range b.entities {
		if v == child {
			b.entities = append(b.entities[:k], b.entities[k+1:]...)
			break
		}

	}
}

// Only clears the list, does not call Destroy() on them or anythin
func (b *BaseEntity) ClearChildren() {
	b.entities = make([]Entity, 0)
}

func (b *BaseEntity) DestroyChildren() {
	for _, v := range b.entities {
		if v != nil {
			v.Destroy()
		}
	}
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
