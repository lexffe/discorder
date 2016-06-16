package ui

type Manager struct {
	Windows     []Entity
	TextInputs  []*TextInput
	ActiveInput *TextInput
}

func NewManager() *Manager {
	return &Manager{
		Windows:    make([]Entity, 0),
		TextInputs: make([]*TextInput, 0),
	}
}

// Windows
func (u *Manager) AddWindow(e Entity) {
	u.Windows = append(u.Windows, e)
}

func (u *Manager) AddWindowFront(e Entity) {
	u.Windows = append([]Entity{e}, u.Windows...)
}

func (u *Manager) CurrentWindow() Entity {
	if len(u.Windows) > 0 {
		return u.Windows[len(u.Windows)-1]
	}

	return nil
}

func (u *Manager) RemoveWindow(e Entity) bool {
	for i, v := range u.Windows {
		if v == e {
			u.Windows = append(u.Windows[:i], u.Windows[i+1:]...)
			return true
		}
	}
	return false
}

// Inputs
func (u *Manager) AddInput(input *TextInput, setActive bool) {
	u.TextInputs = append(u.TextInputs, input)
	if setActive {
		u.SetActiveInput(input)
	}
}

func (u *Manager) SetActiveInput(input *TextInput) {
	if u.ActiveInput != nil {
		u.ActiveInput.Active = false
	}
	input.Active = true
	u.ActiveInput = input
}

func (u *Manager) RemoveInput(input *TextInput, reAssignActive bool) {
	for k, v := range u.TextInputs {
		if v == input {
			u.TextInputs = append(u.TextInputs[:k], u.TextInputs[k+1:]...)
			break
		}
	}
	if reAssignActive && len(u.TextInputs) > 0 {
		u.SetActiveInput(u.TextInputs[len(u.TextInputs)-1])
	}
}
