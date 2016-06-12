package discorder

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"strconv"
)

type Tab struct {
	*ui.BaseEntity
	app           *App
	Name          string
	Index         int
	ViewContainer ui.Entity
	MessageView   *MessageView
	Active        bool
	SendChannel   string

	IndicatorMarked bool
	Indicator       *ui.Text
}

func NewTab(app *App, index int) *Tab {
	container := ui.NewAutoLayoutContainer()
	container.LayoutType = ui.LayoutTypeHorizontal
	container.ForceExpandHeight = true

	mw := NewMessageView(app)
	container.Transform.AddChildren(mw)
	t := &Tab{
		BaseEntity:    &ui.BaseEntity{},
		app:           app,
		Name:          strconv.FormatInt(int64(index), 10),
		Index:         index,
		ViewContainer: container,
		MessageView:   mw,
		Indicator:     ui.NewText(),
	}

	t.Transform.AddChildren(container)
	container.Transform.AnchorMax = common.NewVector2I(1, 1)
	t.Indicator.Text = t.Name

	app.ApplyThemeToText(t.Indicator, "tab_normal")

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
	if active {
		t.app.ApplyThemeToText(t.Indicator, "tab_selected")
	} else {
		t.app.ApplyThemeToText(t.Indicator, "tab_normal")
	}
	t.MessageView.DisplayMessagesDirty = true
	t.IndicatorMarked = false
}

func (t *Tab) SetName(name string) {
	t.Name = name
	t.Indicator.Text = name
}

func (t *Tab) Destroy() { t.DestroyChildren() }

// Implement sort for tab slices
type TabSlice []*Tab

func (t TabSlice) Len() int {
	return len([]*Tab(t))
}

func (t TabSlice) Less(a, b int) bool {
	return t[a].Index < t[b].Index
}

func (t TabSlice) Swap(i, j int) {
	temp := t[i]
	t[i] = t[j]
	t[j] = temp
}
