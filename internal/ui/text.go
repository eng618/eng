package ui

// Truncate shortens s to at most maxLen runes, appending "..." when cut.
// It is rune-aware (never splits multibyte characters) and total-length
// safe: non-positive maxLen yields "", tiny limits yield a prefix without
// ellipsis. This is the single shared implementation; do not reimplement
// truncation elsewhere.
func Truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen < 3 {
		if maxLen < 0 {
			return ""
		}
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}
