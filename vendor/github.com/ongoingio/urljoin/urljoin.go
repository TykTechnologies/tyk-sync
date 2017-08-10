package urljoin

import (
	"strings"
)

// Join joins any number of URL parts into a single URL, correctly adding
// slashes. It doesn't add a trailing slash, but keeps one if present.
func Join(parts ...string) string {
	l := len(parts)
	if l == 1 {
		return parts[0]
	}
	ps := make([]string, l)
	for i, part := range parts {
		if i == 0 {
			ps[i] = strings.TrimRight(part, "/")
		} else {
			ps[i] = strings.TrimLeft(part, "/")
		}
	}
	return strings.Join(ps, "/")
}
