package discorder

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"strconv"
)

type Tab struct {
	*ui.BaseEntity

	Name          string
	Index         int
	ViewContainer ui.Entity
	MessageView   *MessageView
	Active        bool
	SendChannel   string

	Indicator *TabIndicator
}

func NewTab(app *App, index int) *Tab {
	container := ui.NewAutoLayoutContainer()
	container.LayoutType = ui.LayoutTypeHorizontal
	container.ForceExpandHeight = true

	mw := NewMessageView(app)
	container.Transform.AddChildren(mw)
	t := &Tab{
		BaseEntity:    &ui.BaseEntity{},
		Name:          strconv.FormatInt(int64(index), 10),
		Index:         index,
		ViewContainer: container,
		MessageView:   mw,
	}

	t.Transform.AddChildren(container)
	container.Transform.AnchorMax = common.NewVector2I(1, 1)

	return t
}

func (t *Tab) GetRequiredSize() common.Vector2F {
	return common.NewVector2F(0, 0)
}

func (t *Tab) IsLayoutDynamic() bool {
	return t.Active
}

func (t *Tab) SetActive(active bool) {
	t.Active = active
}

func (t *Tab) Destroy() { t.DestroyChildren() }

type TabIndicator struct {
	*ui.BaseEntity
}
