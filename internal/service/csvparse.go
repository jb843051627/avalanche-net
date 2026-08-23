package service

// splitCSVLine 按逗号切分并去空白。
func splitCSVLine(row string) []string {
	raw := splitString(row, ',')
	out := make([]string, 0, len(raw))
	for _, p := range raw {
		p = trimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// splitDash 按 '-' 切分深度区间。
func splitDash(s string) []string { return splitString(s, '-') }

func splitString(s string, sep byte) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}
