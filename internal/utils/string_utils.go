package utils

import "strings"

func CutStringAtChar(s string, c string) string {
	return s[:strings.Index(s, c)]
}
