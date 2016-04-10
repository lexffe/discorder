package ui

import (
	"github.com/nsf/termbox-go"
)

type Entity interface {
	Children(recursive bool) []Entity
	Init()
	Destroy()
}

type BaseEntity struct {
	entities []Entity
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

func (b *BaseEntity) AddChild(child Entity) {
	if b.entities == nil {
		b.entities = []Entity{child}
	} else {
		b.entities = append(b.entities, child)
	}
}

func (b *BaseEntity) RemoveChild(child Entity, destroy bool) {
	if b.entities == nil || len(b.entities) < 1 {
		return
	}

	if destroy {
		child.Destroy()
	}

	index := -1
	for k, v := range b.entities {
		if v == child {
			index = k
			break
		}

	}

	if index != -1 {
		if index == len(b.entities)-1 {
			b.entities = b.entities[:index-1]
		} else {
			b.entities = append(b.entities[:index], b.entities[index+1:]...)
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

type InputHandler interface {
	HandleInput(event termbox.Event)
}

type Drawable interface {
	GetDrawLayer() int
	Draw()
}
