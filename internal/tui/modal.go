package tui

import tea "github.com/charmbracelet/bubbletea"

// A modal is a temporary UI that takes over the keyboard while it is
// open: the photo viewer, a question, a one-line prompt, an overlay or
// the reaction picker. Each closes itself (by clearing its field on the
// model) when it is done.
type modal interface {
	// key handles a key press while the modal is open.
	key(m *model, msg tea.KeyMsg) tea.Cmd
	// help is the line of key hints shown while it is open.
	help() string
}

// modal returns what owns the keyboard, or nil when the focused area
// does. Only the viewer can open while another modal is up (a photo
// finishes loading), so it comes first.
func (m *model) modal() modal {
	switch {
	case m.viewer != nil:
		return m.viewer
	case m.confirm != nil:
		return m.confirm
	case m.prompt != nil:
		return m.prompt
	case m.overlay != nil:
		return m.overlay
	case m.react != nil:
		return m.react
	}
	return nil
}
