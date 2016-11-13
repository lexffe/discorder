package ui

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/go-runewidth"
	"github.com/jonas747/termbox-go"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	TextModeOverflow = iota
	TextModeHide
	TextModeWrap
)

type Text struct {
	*BaseEntity
	Disabled bool // won't draw then

	Text string

	SkipLines int

	Mode int

	// If attribs is empty, uses style instead
	attribs []*AttribPair
	Style   AttribPair

	Layer    int
	Userdata interface{}

	builtLines []string
	lastBuild  string
}

func NewText() *Text {
	t := &Text{
		BaseEntity: &BaseEntity{},
	}
	return t
}

func (t *Text) GetDrawLayer() int {
	return t.Layer
}

func (t *Text) SetAttribs(attribs map[int]AttribPair) {
	highest := 0
	for key := range attribs {
		if key > highest {
			highest = key
		}
	}

	t.attribs = make([]*AttribPair, highest+1)
	for key, pair := range attribs {
		c := pair
		t.attribs[key] = &c
	}
}

func (t *Text) Draw() {
	if t.Disabled {
		return
	}

	if len(t.builtLines) < 1 || t.lastBuild != t.Text {
		t.BuildLines()
	}

	rect := t.Transform.GetRect()

	var attribs []*AttribPair
	if t.attribs != nil && len(t.attribs) > 0 {
		attribs = t.attribs
	} else {
		attribs = []*AttribPair{&t.Style}
	}

	// The actual drawing happens here
	y := 0
	i := 0
	skip := t.SkipLines
	height := int(rect.H)

	var curAttribs AttribPair

	for _, line := range t.builtLines {
		if y >= height && height != 0 {
			break
		}

		if skip > 0 {
			//log.Println("SKipped a line", skip, line)
			skip--
			i += utf8.RuneCountInString(line)
			y++
			continue
		}
		x := 0
		for _, char := range line {
			if i < len(attribs) {
				newAttribs := attribs[i]
				if newAttribs != nil {
					curAttribs = *newAttribs
				}
			}

			charWidth := runewidth.RuneWidth(char)
			if charWidth == 0 {
				continue
			}

			termbox.SetCell(x+int(rect.X), y+int(rect.Y), char, curAttribs.FG, curAttribs.BG)
			x += charWidth
			i++
		}
		y++
	}

	// cellSlice := GenCellSlice(t.Text, attribs)
	// SetCells(cellSlice, int(rect.X), int(rect.Y), int(rect.W), int(rect.H))
}

func (t *Text) HeightRequired() int {
	if t.Disabled {
		return 0
	}

	rect := t.Transform.GetRect()
	return HeightRequired(t.Text, int(rect.W))
}

// Implement LayoutElement
func (t *Text) GetRequiredSize() common.Vector2F {
	//rect := t.Transform.GetRect()
	return common.NewVector2I(runewidth.StringWidth(t.Text), t.HeightRequired())
}

func (t *Text) IsLayoutDynamic() bool {
	return false
}

func (t *Text) Destroy() { t.DestroyChildren() }

func (t *Text) BuildLines() {
	rect := t.Transform.GetRect()
	t.builtLines = BuildTextLines(t.Text, int(rect.W))
	t.lastBuild = t.Text
}

func BuildTextLines(in string, width int) []string {
	lines := strings.Split(in, "\n")
	secondPass := make([]string, 0)

	if width < 1 {
		// Uh yeah lets just not
		return []string{""}
	}
	for _, line := range lines {
		for {
			split, rest := StrSplit(line, width)
			secondPass = append(secondPass, split)
			if rest == "" {
				break
			}
			line = rest
		}
	}

	return secondPass
}

func StrSplit(s string, width int) (split, rest string) {
	// Possibly split up s
	if runewidth.StringWidth(s) > width {
		_, beforeIndex := RuneByPhysPosition(s, width)
		firstPart := s[:beforeIndex]

		// Split at newline if possible
		foundWhiteSpace := false
		lastIndex := strings.LastIndex(firstPart, "\n")
		if lastIndex == -1 {
			// No newline, check for any possible whitespace then
			lastIndex = strings.LastIndexFunc(firstPart, func(r rune) bool {
				return unicode.In(r, unicode.White_Space)
			})
			if lastIndex == -1 {
				lastIndex = beforeIndex
			} else {
				foundWhiteSpace = true
			}
		} else {
			foundWhiteSpace = true
		}

		// Remove the whitespace we split at if any
		if foundWhiteSpace {
			_, rLen := utf8.DecodeRuneInString(s[lastIndex:])
			rest = s[lastIndex+rLen:]
		} else {
			rest = s[lastIndex:]
		}

		split = s[:lastIndex]
	} else {
		split = s
	}

	return
}

// Returns the rune at the physical position, that means it takes
// The width of runes into account
func RuneByPhysPosition(s string, runePos int) (rune, int) {
	sLen := runewidth.StringWidth(s)
	if sLen <= runePos || runePos < 0 {
		panic("runePos is out of bounds")
	}

	i := 0
	lastK := 0
	lastR := rune(0)
	for k, r := range s {
		if i == runePos {
			return r, k
		}
		i += runewidth.RuneWidth(r)
		lastR = r
		lastK = k
	}
	return lastR, lastK
}
