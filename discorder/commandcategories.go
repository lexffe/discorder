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
	&CommandCategory{
		Name:        "Discord",
		Description: "Discord utilities",
	}, &CommandCategory{
		Name:        "Windows",
		Description: "Server browser, help etc...",
	}, &CommandCategory{
		Name:        "Misc",
		Description: "Misc commands",
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

func (cc *CommandCategory) GenMenu(cmds []*Command, categories []*CommandCategory) *ui.MenuItem {
	item := &ui.MenuItem{
		Name:       cc.Name,
		Info:       cc.Description,
		IsCategory: true,
		Children:   make([]*ui.MenuItem, 0),
	}

	// Add sub categories
	for _, sub := range cc.Children {
		subItem := sub.GenMenu(cmds, categories)
		item.Children = append(item.Children, subItem)
	}

	// Add commands
	for _, cmd := range cmds {
		if GetCategoryFromPath(cmd.Category, categories) == cc {
			item.Children = append(item.Children, cmd.GenMenuItem())
		}
	}
	return item
}
