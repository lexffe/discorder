package discorder

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discorder/ui"
	"github.com/jonas747/discordgo"
	"strings"
	"unicode/utf8"
)

type MentionAutoCompletion struct {
	*ui.BaseEntity
	App                     *App
	input                   *ui.TextInput
	mentionMatches          []*discordgo.Member
	mentionSelect           int
	lastBufferCheck         string
	dirty                   bool
	isAutocompletingMention bool
}

func NewMentionAutoCompletion(app *App, input *ui.TextInput) *MentionAutoCompletion {
	return &MentionAutoCompletion{
		BaseEntity: &ui.BaseEntity{},
		App:        app,
		input:      input,
	}
}

func (ma *MentionAutoCompletion) OnCommand(cmd Command, args Arguments) {
	if cmd.GetName() != "autocomplete_selection" {
		return
	}

	amount, _ := args.Int("amount")

	ma.mentionSelect += amount
	if ma.mentionSelect < 0 {
		ma.mentionSelect = 0
	} else if ma.mentionSelect >= len(ma.mentionMatches) {
		ma.mentionSelect = 0
	}
	ma.dirty = true
}

func (ma *MentionAutoCompletion) Update() {

	if ma.lastBufferCheck != ma.input.TextBuffer || ma.dirty {
		if ma.lastBufferCheck != ma.input.TextBuffer {
			ma.lastBufferCheck = ma.input.TextBuffer
			ma.mentionSelect = 0 // Reset the selection if input changed
		}

		ma.Check()

		ma.Transform.ClearChildren(true)
		if ma.isAutocompletingMention {
			rect := ma.Transform.GetRect()
			curX := float32(0)
			for k, v := range ma.mentionMatches {
				t := ui.NewText()

				t.Transform.AnchorMax.Y = 1
				t.Transform.Position.X = curX

				t.Text = v.User.Username
				size := utf8.RuneCountInString(t.Text)
				t.Transform.Size.X = float32(size)

				if k == ma.mentionSelect {
					ma.App.ApplyThemeToText(t, "element_selected")
				} else {
					ma.App.ApplyThemeToText(t, "element_normal")
				}

				ma.Transform.AddChildren(t)
				curX += float32(size) + 1
				if curX > rect.W {
					break
				}
			}
		}
	}
}

func (ma *MentionAutoCompletion) Check() {
	split := strings.Split(ma.input.TextBuffer, " ")
	currentIndex := ma.FindMatchSubIndex(split)
	//log.Println(currentIndex, split[currentIndex])
	if len(split[currentIndex]) > 0 && split[currentIndex][0] == '@' {
		if !ma.isAutocompletingMention {
			ma.isAutocompletingMention = true
			ma.mentionMatches = make([]*discordgo.Member, 0)
			ma.mentionSelect = 0
		}

		ma.FindMatchingMentions(currentIndex)
	} else {
		ma.isAutocompletingMention = false
	}
}

func (ma *MentionAutoCompletion) FindMatchingMentions(subIndex int) {
	split := strings.Split(ma.input.TextBuffer, " ")

	if len(split[subIndex]) < 2 {
		ma.mentionMatches = []*discordgo.Member{}
		return
	}
	strToSearchFor := split[subIndex][1:]

	matches := make([]*discordgo.Member, 0)

	talkingChannel, err := ma.App.session.State.Channel(ma.App.ViewManager.ActiveTab.SendChannel)
	if err != nil {
		return // Invalid channel or channels not loaded
	}

	selectedGuild, err := ma.App.session.State.Guild(talkingChannel.GuildID)
	if err != nil {
		return // Invalid guid... still warming up then probably
	}

	for _, member := range selectedGuild.Members {
		if strings.Contains(strings.ToLower(member.User.Username), strings.ToLower(strToSearchFor)) {
			matches = append(matches, member)
		}
	}
	ma.mentionMatches = matches
}

func (ma *MentionAutoCompletion) PerformAutocompleteMention() bool {
	if ma.mentionSelect > len(ma.mentionMatches)-1 {
		return false
	}

	split := strings.Split(ma.input.TextBuffer, " ")
	currentIndex := ma.FindMatchSubIndex(split)

	selected := ma.mentionMatches[ma.mentionSelect]

	split[currentIndex] = "<@" + selected.User.ID + ">"
	ma.input.CursorLocation = len(split) - 1
	str := ""
	for k, v := range split {
		if k <= currentIndex {
			ma.input.CursorLocation += utf8.RuneCountInString(v)
		}
		str += v + " "
	}
	str = str[:len(str)-1]

	ma.input.TextBuffer = str
	return true
}

func (ma *MentionAutoCompletion) FindMatchSubIndex(split []string) int {
	i := 0
	currentIndex := 0
	for k, v := range split {
		i += utf8.RuneCountInString(v)
		if i >= ma.input.CursorLocation-k {
			currentIndex = k
			break
		}
	}
	return currentIndex
}
func (ma *MentionAutoCompletion) Destroy() { ma.DestroyChildren() }

// Implements LayoutElement
func (ma *MentionAutoCompletion) GetRequiredSize() common.Vector2F {
	size := 0
	if ma.isAutocompletingMention {
		size = 1
	}
	return common.NewVector2I(0, size)
}

func (ma *MentionAutoCompletion) IsLayoutDynamic() bool {
	return false
}
