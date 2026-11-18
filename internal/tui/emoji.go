package tui

import "strings"

// emoji is one entry of the ":name" completion table.
type emoji struct{ name, char string }

// quickReactions are the six reactions WhatsApp offers first.
var quickReactions = []string{"👍", "❤️", "😂", "😮", "😢", "🙏"}

// emojiTable lists the emoji offered after ":" in the message box, most
// common first. Typing emoji in a terminal is awkward, so this covers the
// everyday ones by their usual short names; anything else can still be
// pasted or entered with the system emoji picker (Win+. on Windows).
var emojiTable = []emoji{
	{"smile", "😄"}, {"grin", "😁"}, {"joy", "😂"}, {"rofl", "🤣"}, {"lol", "😆"},
	{"slight_smile", "🙂"}, {"wink", "😉"}, {"blush", "😊"}, {"innocent", "😇"}, {"heart_eyes", "😍"},
	{"star_struck", "🤩"}, {"kiss", "😘"}, {"yum", "😋"}, {"tongue", "😛"}, {"zany", "🤪"},
	{"hug", "🤗"}, {"thinking", "🤔"}, {"shush", "🤫"}, {"neutral", "😐"}, {"expressionless", "😑"},
	{"smirk", "😏"}, {"unamused", "😒"}, {"eyeroll", "🙄"}, {"grimace", "😬"}, {"relieved", "😌"},
	{"pensive", "😔"}, {"sleepy", "😪"}, {"sleeping", "😴"}, {"mask", "😷"}, {"sick", "🤒"},
	{"nauseated", "🤢"}, {"hot", "🥵"}, {"cold", "🥶"}, {"dizzy", "😵"}, {"mind_blown", "🤯"},
	{"cowboy", "🤠"}, {"party", "🥳"}, {"cool", "😎"}, {"sunglasses", "😎"}, {"nerd", "🤓"},
	{"confused", "😕"}, {"worried", "😟"}, {"frown", "🙁"}, {"open_mouth", "😮"}, {"wow", "😮"},
	{"astonished", "😲"}, {"flushed", "😳"}, {"pleading", "🥺"}, {"scared", "😨"}, {"cry", "😢"},
	{"sob", "😭"}, {"scream", "😱"}, {"sweat", "😓"}, {"sweat_smile", "😅"}, {"tired", "😫"},
	{"yawn", "🥱"}, {"angry", "😠"}, {"rage", "😡"}, {"cursing", "🤬"}, {"devil", "😈"},
	{"skull", "💀"}, {"poop", "💩"}, {"clown", "🤡"}, {"ghost", "👻"}, {"alien", "👽"},
	{"robot", "🤖"}, {"see_no_evil", "🙈"}, {"hear_no_evil", "🙉"}, {"speak_no_evil", "🙊"},
	{"salute", "🫡"}, {"melting", "🫠"}, {"facepalm", "🤦"}, {"shrug", "🤷"},

	{"thumbsup", "👍"}, {"+1", "👍"}, {"like", "👍"}, {"thumbsdown", "👎"}, {"-1", "👎"},
	{"ok_hand", "👌"}, {"pinch", "🤏"}, {"victory", "✌️"}, {"peace", "✌️"}, {"crossed_fingers", "🤞"},
	{"rock", "🤘"}, {"call_me", "🤙"}, {"point_left", "👈"}, {"point_right", "👉"}, {"point_up", "👆"},
	{"point_down", "👇"}, {"raised_hand", "✋"}, {"wave", "👋"}, {"clap", "👏"}, {"raised_hands", "🙌"},
	{"open_hands", "👐"}, {"handshake", "🤝"}, {"pray", "🙏"}, {"thanks", "🙏"}, {"namaste", "🙏"},
	{"writing", "✍️"}, {"muscle", "💪"}, {"fist", "✊"}, {"punch", "👊"}, {"eyes", "👀"},
	{"brain", "🧠"}, {"tongue_out", "👅"}, {"lips", "👄"},

	{"heart", "❤️"}, {"red_heart", "❤️"}, {"orange_heart", "🧡"}, {"yellow_heart", "💛"}, {"green_heart", "💚"},
	{"blue_heart", "💙"}, {"purple_heart", "💜"}, {"black_heart", "🖤"}, {"white_heart", "🤍"}, {"broken_heart", "💔"},
	{"two_hearts", "💕"}, {"sparkling_heart", "💖"}, {"love", "😍"}, {"kiss_mark", "💋"}, {"100", "💯"},
	{"anger", "💢"}, {"boom", "💥"}, {"sweat_drops", "💦"}, {"dash", "💨"}, {"zzz", "💤"},

	{"fire", "🔥"}, {"sparkles", "✨"}, {"star", "⭐"}, {"glowing_star", "🌟"}, {"zap", "⚡"},
	{"sun", "☀️"}, {"moon", "🌙"}, {"cloud", "☁️"}, {"rain", "🌧️"}, {"snowflake", "❄️"},
	{"rainbow", "🌈"}, {"umbrella", "☔"}, {"earth", "🌍"}, {"rose", "🌹"}, {"flower", "🌸"},
	{"sunflower", "🌻"}, {"tree", "🌳"}, {"leaf", "🍃"}, {"dog", "🐶"}, {"cat", "🐱"},
	{"monkey", "🐒"}, {"lion", "🦁"}, {"tiger", "🐯"}, {"unicorn", "🦄"}, {"bee", "🐝"},
	{"butterfly", "🦋"}, {"snake", "🐍"}, {"goat", "🐐"}, {"chicken", "🐔"}, {"fish", "🐟"},

	{"tada", "🎉"}, {"confetti", "🎊"}, {"balloon", "🎈"}, {"gift", "🎁"}, {"cake", "🎂"},
	{"birthday", "🎂"}, {"trophy", "🏆"}, {"medal", "🏅"}, {"crown", "👑"}, {"gem", "💎"},
	{"money", "💰"}, {"dollar", "💵"}, {"rupee", "💸"}, {"chart_up", "📈"}, {"chart_down", "📉"},
	{"rocket", "🚀"}, {"plane", "✈️"}, {"car", "🚗"}, {"bike", "🚲"}, {"train", "🚆"},
	{"house", "🏠"}, {"office", "🏢"}, {"hospital", "🏥"}, {"mosque", "🕌"}, {"kaaba", "🕋"},

	{"coffee", "☕"}, {"tea", "🍵"}, {"chai", "🍵"}, {"beer", "🍺"}, {"wine", "🍷"},
	{"pizza", "🍕"}, {"burger", "🍔"}, {"fries", "🍟"}, {"biryani", "🍛"}, {"curry", "🍛"},
	{"rice", "🍚"}, {"noodles", "🍜"}, {"egg", "🍳"}, {"apple", "🍎"}, {"mango", "🥭"},
	{"banana", "🍌"}, {"ice_cream", "🍦"}, {"chocolate", "🍫"}, {"cookie", "🍪"}, {"popcorn", "🍿"},

	{"check", "✅"}, {"done", "✅"}, {"tick", "✔️"}, {"x", "❌"}, {"cross", "❌"},
	{"warning", "⚠️"}, {"no_entry", "⛔"}, {"question", "❓"}, {"exclamation", "❗"}, {"info", "ℹ️"},
	{"bulb", "💡"}, {"idea", "💡"}, {"bell", "🔔"}, {"pin", "📌"}, {"paperclip", "📎"},
	{"link", "🔗"}, {"lock", "🔒"}, {"key", "🔑"}, {"search", "🔍"}, {"hourglass", "⏳"},
	{"clock", "🕐"}, {"alarm", "⏰"}, {"calendar", "📅"}, {"memo", "📝"}, {"book", "📖"},
	{"books", "📚"}, {"mail", "📧"}, {"phone", "📱"}, {"call", "📞"}, {"laptop", "💻"},
	{"computer", "🖥️"}, {"keyboard", "⌨️"}, {"camera", "📷"}, {"video", "🎥"}, {"mic", "🎤"},
	{"headphones", "🎧"}, {"music", "🎵"}, {"guitar", "🎸"}, {"game", "🎮"}, {"dice", "🎲"},
	{"soccer", "⚽"}, {"cricket", "🏏"}, {"basketball", "🏀"}, {"gym", "🏋️"}, {"run", "🏃"},
	{"bug", "🐛"}, {"wrench", "🔧"}, {"hammer", "🔨"}, {"gear", "⚙️"}, {"package", "📦"},
	{"ship", "🚢"}, {"recycle", "♻️"}, {"up", "⬆️"}, {"down", "⬇️"}, {"left", "⬅️"},
	{"right", "➡️"}, {"red_circle", "🔴"}, {"green_circle", "🟢"}, {"blue_circle", "🔵"}, {"yellow_circle", "🟡"},
}

// emojiQuery returns the name typed after a ":" that starts a word, up to
// the cursor: ":smi" -> "smi". At least two letters, so "10:30" and ":)"
// don't open the dropdown.
func emojiQuery(value string, pos int) (query string, start int, ok bool) {
	runes := []rune(value)
	before := runes[:max(0, min(pos, len(runes)))]
	i := len(before) - 1
	for i >= 0 && isEmojiNameRune(before[i]) {
		i--
	}
	if i < 0 || before[i] != ':' || len(before)-i-1 < 2 {
		return "", 0, false
	}
	if i > 0 && before[i-1] != ' ' && before[i-1] != '\n' {
		return "", 0, false
	}
	return strings.ToLower(string(before[i+1:])), i, true
}

func isEmojiNameRune(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '+' || r == '-'
}

// matchEmoji returns emoji whose name starts with q, then ones that
// merely contain it.
func matchEmoji(q string) []emoji {
	var out, rest []emoji
	for _, e := range emojiTable {
		switch {
		case strings.HasPrefix(e.name, q):
			out = append(out, e)
		case strings.Contains(e.name, q):
			rest = append(rest, e)
		}
		if len(out) == maxSuggestions {
			return out
		}
	}
	return append(out, rest...)[:min(maxSuggestions, len(out)+len(rest))]
}
