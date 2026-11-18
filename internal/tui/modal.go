package tui

import (
	"os"
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/codebyoketch/wa-cli/internal/whatsapp"
)

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

// --- a question answered with one key ---

type confirmPrompt struct {
	text   string
	keys   map[string]func() tea.Cmd
	opened time.Time
}

// confirmDelay is how long a question must have been on screen before an
// answer counts: long enough to rule out a key that was already on its
// way (a paste, a held key), far shorter than reading the question takes.
const confirmDelay = 350 * time.Millisecond

func (c *confirmPrompt) help() string { return "enter confirm · esc cancel" }

func (c *confirmPrompt) key(m *model, msg tea.KeyMsg) tea.Cmd {
	m.confirm = nil
	if msg.Paste || time.Since(c.opened) < confirmDelay {
		return nil
	}
	if act, ok := c.keys[strings.ToLower(keyText(msg))]; ok {
		return act()
	}
	return nil
}

// --- a one-line text prompt ---

type linePrompt struct {
	label  string
	ed     editor
	submit func(m *model, text string) tea.Cmd
}

func (p *linePrompt) help() string { return "enter confirm · esc cancel" }

func (p *linePrompt) key(m *model, msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEsc:
		m.prompt = nil
	case tea.KeyEnter:
		m.prompt = nil
		return p.submit(m, strings.TrimSpace(p.ed.Value()))
	case tea.KeyBackspace:
		p.ed.Backspace()
	case tea.KeyDelete:
		p.ed.Delete()
	case tea.KeyLeft:
		p.ed.Left()
	case tea.KeyRight:
		p.ed.Right()
	case tea.KeyHome:
		p.ed.Start()
	case tea.KeyEnd:
		p.ed.End()
	case tea.KeyCtrlU:
		p.ed.Reset()
	default:
		if text := keyText(msg); text != "" {
			p.ed.Insert(oneLine(text))
		}
	}
	return nil
}

func (m *model) openAttachPrompt() tea.Cmd {
	if m.openJID == "" {
		return nil
	}
	ed := newEditor()
	ed.maxRows = 1
	ed.SetWidth(max(10, m.calc().paneW-22))
	m.prompt = &linePrompt{
		label: "File to send:",
		ed:    ed,
		submit: func(m *model, text string) tea.Cmd {
			if text == "" {
				return nil
			}
			path, ok := asFilePath(text)
			if !ok {
				return m.setStatus("No such file: "+text, true)
			}
			return m.setAttachment(path)
		},
	}
	return m.info("Type or paste a file's full path (or just drop the file on the terminal).")
}

// setAttachment queues a file to go out with the next Enter.
func (m *model) setAttachment(path string) tea.Cmd {
	info, err := os.Stat(path)
	if err != nil {
		return m.fail(err)
	}
	m.attach = &attachment{path: path, size: info.Size()}
	m.focus = focusInput
	return nil
}

// fileKind says how a file will be sent.
func fileKind(path string) string {
	switch whatsapp.AttachmentKind(path) {
	case whatsapp.KindImage:
		return "a photo (full quality)"
	case whatsapp.KindVideo:
		return "a video"
	case whatsapp.KindVoice:
		return "a voice message"
	case whatsapp.KindAudio:
		return "audio"
	}
	return "a document"
}

// --- reaction picker ---

type reactPicker struct {
	target whatsapp.MessageView
	idx    int
}

func newReactPicker(v whatsapp.MessageView) *reactPicker { return &reactPicker{target: v} }

func (r *reactPicker) send(m *model, emoji string) tea.Cmd {
	m.react = nil
	ctx, be, target := m.ctx, m.be, r.target.Message
	return do(func() error { return be.React(ctx, target, emoji) })
}

func (r *reactPicker) help() string {
	return "←/→ choose · enter react · type any emoji · del remove yours · esc cancel"
}

func (r *reactPicker) key(m *model, msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEsc:
		m.react = nil
	case tea.KeyLeft:
		r.idx = (r.idx + len(quickReactions) - 1) % len(quickReactions)
	case tea.KeyRight, tea.KeyTab:
		r.idx = (r.idx + 1) % len(quickReactions)
	case tea.KeyEnter:
		return r.send(m, quickReactions[r.idx])
	case tea.KeyBackspace, tea.KeyDelete:
		return r.send(m, "")
	default:
		text := strings.TrimSpace(keyText(msg))
		if text == "" {
			return nil
		}
		if len(text) == 1 && text[0] >= '1' && int(text[0]-'1') < len(quickReactions) {
			return r.send(m, quickReactions[text[0]-'1'])
		}
		// Anything that isn't plain letters is taken as the emoji itself
		// (typed with the emoji picker, or pasted).
		for _, c := range text {
			if c < unicode.MaxASCII {
				return nil
			}
		}
		return r.send(m, text)
	}
	return nil
}

func (r *reactPicker) view(w int) string {
	var b strings.Builder
	b.WriteString(" " + headerStyle.Render("React") + " ")
	for i, e := range quickReactions {
		if i == r.idx {
			b.WriteString(selectedStyle.Render(" " + e + " "))
		} else {
			b.WriteString(" " + e + " ")
		}
	}
	return ansi.Truncate(b.String(), w, "")
}
