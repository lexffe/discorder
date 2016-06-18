package ui

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/termbox-go"
	"sort"
	"strings"
)

type MenuItem struct {
	Name       string
	IsCategory bool
	Decorative bool // Not selectable

	IsInput          bool
	InputType        DataType
	InputDefaultText string

	Marked      bool
	Highlighted bool
	Info        string

	UserData interface{}

	Children []*MenuItem

	Text  *Text
	Input *TextInput

	matches int
	path    string
}

func (mi *MenuItem) GetDisplayName(searching bool) string {
	if searching {
		return mi.path + mi.Name
	}

	if mi.IsCategory {
		return "[Dir] " + mi.Name
	}

	return mi.Name
}

// Runs f recursively on all children
func (mi *MenuItem) RunFunc(f func(mi *MenuItem) bool) bool {
	cont := f(mi)
	if !cont {
		return false
	}
	if mi.Children != nil && len(mi.Children) > 0 {
		for _, v := range mi.Children {
			cont := v.RunFunc(f)
			if !cont {
				return false
			}
		}
	}
	return true
}

type MenuItemSlice []*MenuItem

func (mi MenuItemSlice) Len() int {
	return len([]*MenuItem(mi))
}

func (mi MenuItemSlice) Less(a, b int) bool {
	return mi[a].matches > mi[b].matches
}

func (mi MenuItemSlice) Swap(i, j int) {
	temp := mi[i]
	mi[i] = mi[j]
	mi[j] = temp
}

type MenuWindow struct {
	*BaseEntity
	Window      *Window
	LowerWindow *Window

	MainContainer     *AutoLayoutContainer
	TopContainer      *Container
	MenuItemContainer *AutoLayoutContainer
	LowerContainer    *Container

	InfoText    *Text
	SearchInput *TextInput

	// Style
	StyleNormal         AttribPair
	StyleMarked         AttribPair
	StyleSelected       AttribPair
	StyleMarkedSelected AttribPair
	StyleInputNormal    AttribPair

	Options         []*MenuItem
	FilteredOptions []*MenuItem
	Selectables     []int
	Highlighted     int

	CurDir []string

	texts []*Text
	Dirty bool

	Layer int

	OnSelect             func(*MenuItem)
	lastSearch           string
	shouldResetHighlight bool // if true resets highlight on next update

	manager *Manager
}

func NewMenuWindow(layer int, manager *Manager, searchEnabled bool) *MenuWindow {
	mw := &MenuWindow{
		BaseEntity:        &BaseEntity{},
		Window:            NewWindow(manager),
		LowerWindow:       NewWindow(nil),
		MainContainer:     NewAutoLayoutContainer(),
		TopContainer:      NewContainer(),
		SearchInput:       NewTextInput(manager, layer+1),
		MenuItemContainer: NewAutoLayoutContainer(),
		LowerContainer:    NewContainer(),
		InfoText:          NewText(),
		Layer:             layer,
		Dirty:             true,
		manager:           manager,
	}

	mw.Window.Transform.AnchorMax = common.NewVector2F(1, 1)
	mw.Window.Layer = mw.Layer
	mw.Transform.AddChildren(mw.Window)

	mw.MainContainer.ForceExpandWidth = true
	mw.TopContainer.Dynamic = true
	mw.TopContainer.AllowZeroSize = true

	mw.MenuItemContainer.Transform.AnchorMax = common.NewVector2I(1, 1)
	mw.TopContainer.Transform.AddChildren(mw.MenuItemContainer)
	mw.MenuItemContainer.ForceExpandWidth = true

	mw.Window.Transform.AddChildren(mw.MainContainer)
	mw.MainContainer.Transform.AnchorMax = common.NewVector2F(1, 1)

	mw.LowerWindow.Transform.AddChildren(mw.InfoText)
	mw.LowerWindow.Transform.AnchorMax = common.NewVector2I(1, 1)

	mw.LowerWindow.Layer = layer

	mw.MainContainer.Transform.AddChildren(mw.TopContainer)

	mw.MainContainer.Transform.AddChildren(mw.SearchInput)
	mw.SearchInput.HideCursorWhenEmpty = true
	if searchEnabled {
		manager.SetActiveInput(mw.SearchInput)
	}

	mw.MainContainer.Transform.AddChildren(mw.LowerContainer)

	mw.LowerContainer.Transform.AddChildren(mw.LowerWindow)
	mw.LowerContainer.Transform.AnchorMax = common.NewVector2I(1, 1)
	mw.LowerContainer.ProxySize = mw.InfoText
	mw.LowerContainer.AllowZeroSize = false

	mw.InfoText.Text = "Information"
	mw.InfoText.Transform.AnchorMax = common.NewVector2I(1, 1)
	mw.InfoText.Layer = layer
	manager.AddWindow(mw)
	return mw
}

// Makes sure index is within len(options)
func (mw *MenuWindow) CheckBounds(index int) int {
	if index < 0 {
		return 0
	}
	if index >= len(mw.FilteredOptions) {
		return len(mw.FilteredOptions) - 1
	}
	return index
}

func (mw *MenuWindow) GetIndex(item *MenuItem) int {
	for k, v := range mw.Options {
		if item == v {
			return k
		}
	}

	return -1
}

func (mw *MenuWindow) RemoveMarked(index int) {
	index = mw.CheckBounds(index)
	mw.FilteredOptions[index].Marked = false
	mw.ApplyStyleToItem(mw.FilteredOptions[index])
}

func (mw *MenuWindow) AddMarked(index int) {
	index = mw.CheckBounds(index)
	mw.FilteredOptions[index].Marked = true
	mw.ApplyStyleToItem(mw.FilteredOptions[index])
}

func (mw *MenuWindow) SetHighlighted(index int) {
	if len(mw.FilteredOptions) < 1 {
		return
	}
	// Remove previous selection
	if mw.Highlighted < len(mw.FilteredOptions) && mw.Highlighted >= 0 {
		curHighlighted := mw.FilteredOptions[mw.Highlighted]
		curHighlighted.Highlighted = false
		mw.ApplyStyleToItem(curHighlighted)
	}

	index = mw.CheckBounds(index)
	highlighted := mw.FilteredOptions[index]
	highlighted.Highlighted = true
	if highlighted.IsInput {
		mw.manager.SetActiveInput(highlighted.Input)
	}

	mw.Highlighted = index
	mw.ApplyStyleToItem(highlighted)

	mw.InfoText.Text = highlighted.Info
}

func (mw *MenuWindow) ApplyStyleToItem(item *MenuItem) {
	if item.Text == nil {
		return
	}

	switch {
	case item.Highlighted && item.Marked:
		item.Text.Style = mw.StyleMarkedSelected
	case item.Highlighted:
		item.Text.Style = mw.StyleSelected
	case item.Marked:
		item.Text.Style = mw.StyleMarked
	default:
		if item.IsInput {
			item.Text.Style = mw.StyleInputNormal
		} else {
			item.Text.Style = mw.StyleNormal
		}
	}
}

func (mw *MenuWindow) GetHighlighted() *MenuItem {
	if len(mw.FilteredOptions) < 1 {
		return nil
	}
	index := mw.CheckBounds(mw.Highlighted)
	return mw.FilteredOptions[index]
}

func (mw *MenuWindow) SetOptionsString(options []string) {
	newOptions := make([]*MenuItem, len(options))
	for k, v := range options {
		newOptions[k] = &MenuItem{
			Name:        v,
			Marked:      false,
			Highlighted: false,
		}
	}
	mw.SetOptions(newOptions)
}

func (mw *MenuWindow) SetOptions(options []*MenuItem) {
	mw.Options = options
	mw.Dirty = true
	mw.shouldResetHighlight = true
}

func (mw *MenuWindow) OptionsHeight() int {
	h := 0
	rect := mw.Transform.GetRect()
	for _, v := range mw.FilteredOptions {
		if v.IsInput {
			h++
		} else {
			h += HeightRequired(v.GetDisplayName(mw.lastSearch != ""), int(rect.W))
		}
	}
	return h
}

func (mw *MenuWindow) Rebuild() {
	//mw.ClearChildren()
	//mw.AddChild(mw.Window)
	mw.MenuItemContainer.Transform.ClearChildren(true)
	options := mw.FilteredOptions

	mw.texts = make([]*Text, len(options))

	for k, option := range options {
		var t *Text
		if option.IsInput {
			input := NewTextInput(mw.manager, mw.Layer)
			t = input.Text
			mw.MenuItemContainer.Transform.AddChildren(input)
			input.MinHeight = 1
			option.Input = input
			input.DataType = option.InputType
			input.TextBuffer = option.InputDefaultText
		} else {
			t = NewText()
			t.Text = option.GetDisplayName(mw.lastSearch != "")
			t.Layer = mw.Layer
			mw.MenuItemContainer.Transform.AddChildren(t)
		}
		option.Text = t
		mw.texts[k] = t

		switch {
		case option.Highlighted && option.Marked:
			t.Style = mw.StyleMarkedSelected
		case option.Highlighted:
			t.Style = mw.StyleSelected
		case option.Marked:
			t.Style = mw.StyleMarked
		default:
			if option.IsInput {
				t.Style = mw.StyleInputNormal
			} else {
				t.Style = mw.StyleNormal
			}
		}
	}
}

func (mw *MenuWindow) FilterOptions() {
	// Get the options in the current dir
	var options []*MenuItem
	if mw.lastSearch != "" {
		options = SearchFilter(mw.lastSearch, mw.Options, "")
	} else {
		options = FilterOptionsByPath(mw.CurDir, mw.Options)
	}

	mw.Selectables = make([]int, 0)
	for k, option := range options {
		if !option.Decorative {
			mw.Selectables = append(mw.Selectables, k)
		}
	}

	mw.FilteredOptions = options
}

func FilterOptionsByPath(path []string, options []*MenuItem) []*MenuItem {
	if len(path) < 1 {
		return options
	}

	for _, v := range options {
		if v.Name == path[0] {
			return FilterOptionsByPath(path[1:], v.Children)
		}
	}
	return nil
}

func SearchFilter(searchBy string, in []*MenuItem, path string) []*MenuItem {
	if searchBy == "" {
		return in
	}

	filtered := make([]*MenuItem, 0)
	for _, option := range in {
		if option.IsCategory && len(option.Children) > 0 {
			filtered = append(filtered, SearchFilter(searchBy, option.Children, path+option.Name+"/")...)
			continue
		} else if option.IsCategory {
			continue
		}
		option.path = path
		fullName := path + option.Name
		option.matches = StringSearch(searchBy, fullName)
		if option.matches > 0 {
			filtered = append(filtered, option)
		}
	}
	sort.Sort(MenuItemSlice(filtered))
	return filtered
}

func StringSearch(search, content string) int {
	matches := 0
	curSecondaryIndex := 0

	// Case insensitive
	content = strings.ToLower(content)
	search = strings.ToLower(search)

	for _, v := range search {
		foundIndex := strings.IndexRune(content[curSecondaryIndex:], v)
		if foundIndex != -1 {
			matches++
			curSecondaryIndex += foundIndex + 1
		} else {
			return 0
		}
	}
	return matches
}

func fieldsFunc(r rune) bool {
	return r == ' ' || r == '_' || r == '-'
}

// func (mw *MenuWindow) OnLayoutChanged() {
// 	mw.Rebuild()
// }

func (mw *MenuWindow) Destroy() {
	mw.manager.RemoveWindow(mw)
	mw.DestroyChildren()
}

func (mw *MenuWindow) Update() {
	shouldResetHighlight := false
	if mw.lastSearch != mw.SearchInput.TextBuffer {
		mw.lastSearch = mw.SearchInput.TextBuffer
		mw.Dirty = true
		shouldResetHighlight = true
	}

	if mw.Dirty {
		if shouldResetHighlight || mw.shouldResetHighlight {
			if mw.Highlighted < len(mw.FilteredOptions) && mw.Highlighted >= 0 {
				mw.FilteredOptions[mw.Highlighted].Highlighted = false
			}
		}
		mw.FilterOptions()
		mw.Rebuild()
		if shouldResetHighlight || mw.shouldResetHighlight {
			if len(mw.Selectables) > 0 {
				mw.SetHighlighted(mw.Selectables[0])
			}
			mw.shouldResetHighlight = false
		}
	}

	requiredHeight := mw.OptionsHeight()
	rect := mw.MenuItemContainer.Transform.GetRect()
	_, termSizeY := termbox.Size()

	// Calculate scroll
	if requiredHeight > termSizeY || requiredHeight > int(rect.H) {
		// If window is taller then scroll
		heightPerOption := float64(requiredHeight) / float64(len(mw.FilteredOptions))
		scroll := int(heightPerOption*(float64(len(mw.FilteredOptions)-(mw.Highlighted)))) - (requiredHeight - int(rect.H/2))
		mw.MenuItemContainer.Transform.Top = scroll
		mw.MenuItemContainer.Transform.Bottom = -scroll
	} else {
		mw.MenuItemContainer.Transform.Top = 0
		mw.MenuItemContainer.Transform.Bottom = -0
	}
	mw.Dirty = false
}

func (mw *MenuWindow) CheckBoundsSelectedable(index int) int {
	if index < 0 {
		return 0
	}
	if index >= len(mw.Selectables) {
		return len(mw.Selectables) - 1
	}
	return index
}

func (mw *MenuWindow) Scroll(dir Direction, amount int) {
	if len(mw.Selectables) < 1 {
		return
	}

	selectedableIndex := 0
	for k, v := range mw.Selectables {
		if v == mw.Highlighted {
			selectedableIndex = k
			break
		}
	}
	switch dir {
	case DirUp:
		selectedableIndex -= amount
		mw.SetHighlighted(mw.Selectables[mw.CheckBoundsSelectedable(selectedableIndex)])
	case DirDown:
		selectedableIndex += amount
		mw.SetHighlighted(mw.Selectables[mw.CheckBoundsSelectedable(selectedableIndex)])
	case DirEnd:
		mw.SetHighlighted(mw.Selectables[len(mw.Selectables)-1])
	case DirStart:
		mw.SetHighlighted(mw.Selectables[0])
	}
}

func (mw *MenuWindow) Select() {
	highlighted := mw.GetHighlighted()
	if highlighted == nil {
		return
	}

	if highlighted.IsCategory {
		mw.CurDir = append(mw.CurDir, highlighted.Name)
		mw.Dirty = true
		mw.shouldResetHighlight = true
	} else if mw.OnSelect != nil {
		mw.OnSelect(highlighted)
	}
}

func (mw *MenuWindow) Back() bool {
	if len(mw.CurDir) > 0 {
		mw.CurDir = mw.CurDir[:len(mw.CurDir)-1]
		mw.Dirty = true
		mw.shouldResetHighlight = true
		return true
	}

	return false
}

func (mw *MenuWindow) RunFunc(f func(item *MenuItem) bool) {
	for _, v := range mw.Options {
		cont := v.RunFunc(f)
		if !cont {
			return
		}
	}
}
