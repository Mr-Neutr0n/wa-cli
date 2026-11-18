package tui

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/codebyoketch/wa-cli/internal/whatsapp"
)

// This file holds the overlay: a list drawn in place of the messages,
// used for help, search, forwarding and chat info.

// --- overlays ---

type overlayKind int

const (
	overlayHelp overlayKind = iota
	overlaySearch
	overlayForward
	overlayInfo
)

type overlayItem struct {
	label  string // main text
	detail string // dim text after it
	jid    string // chat this row is about
	id     string // message ID (search results)
	ts     int64
}

// overlay is a list drawn in place of the messages. With a query line it
// filters as you type (search, forward); without it just scrolls (help,
// info).
type overlay struct {
	kind    overlayKind
	title   string
	query   string
	typing  bool // has a query line
	items   []overlayItem
	sel     int
	top     int
	scope   string // search: the chat searched, "" = everywhere
	seq     int    // search: drops results for an older query
	note    string // shown when there are no items
	forward whatsapp.MessageView
}

func (o *overlay) help() string {
	switch o.kind {
	case overlaySearch:
		return "type to search · ↑/↓ move · enter go to message · esc close"
	case overlayForward:
		return "type to filter · ↑/↓ move · enter forward · esc cancel"
	}
	return "↑/↓ scroll · esc close"
}

func (o *overlay) key(m *model, msg tea.KeyMsg) tea.Cmd {
	page := max(1, m.calc().msgH-4)
	switch msg.Type {
	case tea.KeyEsc:
		m.overlay = nil
		return nil
	case tea.KeyUp:
		o.sel--
	case tea.KeyDown:
		o.sel++
	case tea.KeyPgUp:
		o.sel -= page
	case tea.KeyPgDown:
		o.sel += page
	case tea.KeyHome:
		o.sel = 0
	case tea.KeyEnd:
		o.sel = len(o.items) - 1
	case tea.KeyEnter:
		return o.choose(m)
	case tea.KeyBackspace:
		if r := []rune(o.query); o.typing && len(r) > 0 {
			return o.setQuery(m, string(r[:len(r)-1]))
		}
	default:
		text := keyText(msg)
		switch {
		case text == "":
		case o.typing:
			return o.setQuery(m, o.query+oneLine(text))
		case text == "q" || text == "?":
			m.overlay = nil
		}
	}
	o.sel = max(0, min(o.sel, len(o.items)-1))
	return nil
}

func (o *overlay) setQuery(m *model, q string) tea.Cmd {
	o.query = q
	o.sel, o.top = 0, 0
	switch o.kind {
	case overlayForward:
		o.items = forwardTargets(m, q)
	case overlaySearch:
		o.seq++
		if len([]rune(strings.TrimSpace(q))) < 2 {
			o.items, o.note = nil, "Type at least two letters."
			return nil
		}
		o.note = "Searching…"
		return searchCmd(m.ctx, m.be, o.scope, strings.TrimSpace(q), o.seq)
	}
	return nil
}

func (o *overlay) choose(m *model) tea.Cmd {
	if o.sel < 0 || o.sel >= len(o.items) {
		return nil
	}
	it := o.items[o.sel]
	switch o.kind {
	case overlayForward:
		m.overlay = nil
		ctx, be, target := m.ctx, m.be, o.forward.Message
		return tea.Batch(
			m.info("Forwarding to "+it.label+"…"),
			do(func() error { return be.Forward(ctx, it.jid, target) }),
		)
	case overlaySearch:
		m.overlay = nil
		return m.jumpTo(it.jid, it.id)
	}
	return nil
}

// view returns exactly h lines.
func (o *overlay) view(m *model, w, h int) []string {
	out := make([]string, 0, h)
	head := " " + headerStyle.Render(oneLine(o.title))
	if o.typing {
		head += "  " + oneLine(o.query) + cursorStyle.Render(" ")
	}
	out = append(out, ansi.Truncate(head, w, ""), "")

	rows := max(1, h-2)
	if o.kind == overlayHelp || o.kind == overlayInfo {
		// Nothing to select: the arrows simply scroll.
		o.sel = max(0, min(o.sel, len(o.items)-rows))
		o.top = o.sel
	}
	if o.sel < o.top {
		o.top = o.sel
	}
	if o.sel >= o.top+rows {
		o.top = o.sel - rows + 1
	}
	o.top = max(0, min(o.top, len(o.items)-rows))

	if len(o.items) == 0 && o.note != "" {
		out = append(out, "   "+dimStyle.Render(o.note))
	}
	for i := o.top; i < len(o.items) && len(out) < h; i++ {
		it := o.items[i]
		selectable := o.kind == overlaySearch || o.kind == overlayForward
		// Rows hold names and message text from the network: never draw
		// them unsanitized, whatever built the row.
		label := truncate(oneLine(it.label), w-6)
		detail := ""
		if room := w - 8 - ansi.StringWidth(label); it.detail != "" && room > 4 {
			detail = "  " + truncate(oneLine(it.detail), room)
		}
		switch {
		case selectable && i == o.sel:
			out = append(out, " "+selectedStyle.Render(" › "+label+" ")+dimStyle.Render(detail))
		case it.jid == "" && it.id == "" && it.detail == "" && strings.HasSuffix(it.label, ":"):
			out = append(out, " "+headerStyle.Render(label)) // a section title
		default:
			out = append(out, "    "+chatNameStyle.Render(label)+dimStyle.Render(detail))
		}
	}
	for len(out) < h {
		out = append(out, "")
	}
	return out[:h]
}

// --- help ---

func (m *model) openHelp() tea.Cmd {
	lines := []string{
		"Chat list:",
		"type            search chats and contacts (a contact starts a new chat)",
		"↑ ↓ PgUp PgDn   move; the highlighted chat is previewed, not marked read",
		"Enter / Tab     open the chat and start typing",
		"Ctrl+P          pin or unpin",
		"Ctrl+A          archive or unarchive",
		"Ctrl+T          mute or unmute",
		"Ctrl+U          mark unread / read",
		"Ctrl+R          show or hide archived chats",
		"Ctrl+F          search every chat's messages",
		"",
		"Writing:",
		"Enter           send",
		"Shift+Enter     new line (also Ctrl+Enter, Alt+Enter, Ctrl+J)",
		"paste           multi-line text stays one message",
		"@name           mention someone in a group",
		":name           pick an emoji, e.g. :thumbsup",
		"Ctrl+O          attach a file by path - or just drop a file on the terminal",
		"Alt+V           attach the picture on the clipboard (a screenshot)",
		"Ctrl+D          send the attached file as a document instead",
		"↑               select the last message (then act on it)",
		"PgUp PgDn       scroll the chat",
		"Ctrl+F          search this chat",
		"Ctrl+G          chat / group info",
		"Esc / Tab       back to the chat list",
		"",
		"A selected message:",
		"Enter           view photo full screen / open media or link / retry / reply",
		"r               reply",
		"e               react (then 1-6, arrows, or any emoji)",
		"E               edit (your own, within 15 minutes)",
		"d               delete (for everyone, or just here)",
		"f               forward",
		"c               copy the text",
		"o               open with the default app",
		"s               save to Downloads/WhatsApp",
		"Esc             back to writing",
		"",
		"Photo viewer:",
		"m               switch drawing mode (sixel / kitty / iterm / blocks)",
		"o               open in the system's image viewer (true full resolution)",
		"s               save",
		"any other key   close",
		"",
		"Ctrl+C quits. Settings: wa config (sendReadReceipts, sendPresence, mouse, imageProtocol).",
	}
	o := &overlay{kind: overlayHelp, title: "Keys"}
	for _, l := range lines {
		key, rest, ok := strings.Cut(l, "  ")
		if !ok {
			o.items = append(o.items, overlayItem{label: l})
			continue
		}
		o.items = append(o.items, overlayItem{label: padRight(key, 15), detail: strings.TrimSpace(rest)})
	}
	m.overlay = o
	return nil
}

// --- forward ---

func forwardTargets(m *model, q string) []overlayItem {
	q = strings.ToLower(strings.TrimSpace(q))
	var out []overlayItem
	for _, c := range m.chats {
		if q == "" || strings.Contains(c.lower, q) {
			out = append(out, overlayItem{label: c.Title, jid: c.JID})
			if len(out) == 200 {
				break
			}
		}
	}
	return out
}

func (m *model) openForward(v whatsapp.MessageView) tea.Cmd {
	m.overlay = &overlay{kind: overlayForward, title: "Forward to:", typing: true, forward: v, note: "No matching chat."}
	m.overlay.items = forwardTargets(m, "")
	return nil
}

// --- search ---

type searchResultsMsg struct {
	seq   int
	items []overlayItem
	err   error
}

func searchCmd(ctx context.Context, be Backend, chatJID, query string, seq int) tea.Cmd {
	return func() tea.Msg {
		found, err := be.Search(ctx, chatJID, query, 100)
		if err != nil {
			return searchResultsMsg{seq: seq, err: err}
		}
		now := time.Now()
		items := make([]overlayItem, 0, len(found))
		for _, v := range found {
			who := oneLine(v.Sender)
			if v.FromMe {
				who = "You"
			}
			items = append(items, overlayItem{
				label:  oneLine(v.Body),
				detail: strings.TrimSpace(who + " · " + listTime(v.Timestamp, now)),
				jid:    v.ChatJID,
				id:     v.ID,
				ts:     v.Timestamp,
			})
		}
		return searchResultsMsg{seq: seq, items: items}
	}
}

func (o *overlay) searchResults(msg searchResultsMsg) {
	if o.kind != overlaySearch || msg.seq != o.seq {
		return
	}
	o.items, o.sel, o.top = msg.items, 0, 0
	o.note = "Nothing found."
	if msg.err != nil {
		o.note = msg.err.Error()
	}
}

func (m *model) openSearch(scope string) tea.Cmd {
	title := "Search all chats:"
	if scope != "" {
		title = "Search " + m.open.Title + ":"
	}
	m.overlay = &overlay{kind: overlaySearch, title: title, typing: true, scope: scope, note: "Type at least two letters."}
	return nil
}

// jumpTo shows a specific message, switching chats and loading older
// pages as needed.
func (m *model) jumpTo(chatJID, id string) tea.Cmd {
	var cmds []tea.Cmd
	if chatJID != m.openJID {
		c := m.chat(chatJID)
		if c == nil {
			return m.info("That chat is no longer in the list.")
		}
		item := *c
		m.selectedJID = chatJID
		if m.filter != "" {
			m.filter = ""
			m.refilter()
		}
		m.keepSelectionVisible()
		cmds = append(cmds, m.showChat(item))
	}
	m.focus = focusMessages
	m.jump = &jumpTarget{id: id, pages: 40}
	cmds = append(cmds, m.continueJump())
	return tea.Batch(cmds...)
}

// jumpTarget is a message being looked for while older pages load.
type jumpTarget struct {
	id    string
	pages int // how many more pages to try
}

// continueJump is called whenever messages have loaded: select the target
// if it's there, otherwise load one more page.
func (m *model) continueJump() tea.Cmd {
	if m.jump == nil {
		return nil
	}
	if _, ok := m.msgIdx[m.jump.id]; ok {
		m.selID = m.jump.id
		m.jump = nil
		m.revealSelected()
		return m.wantPreviews()
	}
	if len(m.msgs) == 0 {
		return nil // the first page is still loading
	}
	if !m.haveOlder || m.jump.pages <= 0 {
		m.jump = nil
		return m.info("Couldn't find that message in local history.")
	}
	m.jump.pages--
	m.loadingOlder = true
	oldest := m.msgs[0].view.Message
	return loadPageCmd(m.ctx, m.be, m.openJID, m.loadSeq, &oldest, pageSize*2, false)
}

// --- chat info ---

type infoLoadedMsg struct {
	jid     string
	members []whatsapp.Member
	err     error
}

func (m *model) openInfo() tea.Cmd {
	if m.openJID == "" {
		return nil
	}
	o := &overlay{kind: overlayInfo, title: m.open.Title}
	user, _, _ := strings.Cut(m.openJID, "@")
	if m.open.IsGroup {
		o.items = append(o.items, overlayItem{label: "Group", detail: m.openJID})
		o.note = ""
		o.items = append(o.items, overlayItem{label: "Loading members…"})
		m.overlay = o
		ctx, be, jid := m.ctx, m.be, m.openJID
		return func() tea.Msg {
			members, err := be.Members(ctx, jid)
			return infoLoadedMsg{jid: jid, members: members, err: err}
		}
	}
	o.items = append(o.items, overlayItem{label: "Number", detail: "+" + user})
	if whatsapp.IsLIDJID(m.openJID) {
		o.items[0] = overlayItem{label: "ID", detail: m.openJID}
	}
	if sub := m.subtitle(); sub != "" {
		o.items = append(o.items, overlayItem{label: "Status", detail: sub})
	}
	flags := []string{}
	if m.open.Pinned {
		flags = append(flags, "pinned")
	}
	if m.open.Muted || m.open.MutedUntil != 0 {
		flags = append(flags, "muted")
	}
	if m.open.Archived {
		flags = append(flags, "archived")
	}
	if m.open.Ephemeral > 0 {
		flags = append(flags, "disappearing messages on")
	}
	if len(flags) > 0 {
		o.items = append(o.items, overlayItem{label: "Chat", detail: strings.Join(flags, ", ")})
	}
	m.overlay = o
	return nil
}

func (o *overlay) infoLoaded(m *model, msg infoLoadedMsg) {
	if o.kind != overlayInfo || msg.jid != m.openJID {
		return
	}
	o.items = o.items[:1]
	if msg.err != nil {
		o.items = append(o.items, overlayItem{label: "Couldn't load members", detail: msg.err.Error()})
		return
	}
	o.items = append(o.items, overlayItem{label: "", detail: ""}, overlayItem{label: itoa(len(msg.members)+1) + " members:"})
	for _, mem := range msg.members {
		detail := ""
		if mem.IsAdmin {
			detail = "admin"
		}
		o.items = append(o.items, overlayItem{label: oneLine(mem.Name), detail: detail})
	}
}
