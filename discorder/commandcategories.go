package discorder

import (
	"github.com/jonas747/discorder/ui"
)

type CommandCategory struct {
	Name        string
	Description string
	Children    []*CommandCategory
}

var CommandCategories = []*CommandCategory{
	{
		Name:        "Discord",
		Description: "Discord utilities",
	}, {
		Name:        "Windows",
		Description: "Server browser, help etc...",
	}, {
		Name:        "Utils",
		Description: "Discorder utilities",
	},
}

func GetCategoryFromPath(other []string, categories []*CommandCategory) *CommandCategory {
	if len(other) < 1 {
		return nil
	}

	for _, v := range categories {
		if other[0] == v.Name {
			if len(other) == 1 {
				return v
			}

			return GetCategoryFromPath(other[1:], v.Children)
		}
	}

	return nil
}

func (cc *CommandCategory) GenMenu(app *App, cmds []Command, categories []*CommandCategory) *ui.MenuItem {
	item := &ui.MenuItem{
		Name:       cc.Name,
		Info:       cc.Description,
		IsCategory: true,
		Children:   make([]*ui.MenuItem, 0),
	}

	// Add sub categories
	for _, sub := range cc.Children {
		subItem := sub.GenMenu(app, cmds, categories)
		item.Children = append(item.Children, subItem)
	}

	// Add commands
	for _, cmd := range cmds {
		if GetCategoryFromPath(cmd.GetCategory(), categories) == cc {
			item.Children = append(item.Children, app.GenMenuItemFromCommand(cmd))
		}
	}
	return item
}
