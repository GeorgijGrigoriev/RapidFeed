package utils

import (
	"html"
	"regexp"
	"strings"
)

var (
	htmlTagRegexp    = regexp.MustCompile(`<[^>]*>`)
	whitespaceRegexp = regexp.MustCompile(`\s+`)
)

func NormalizeFeedText(input string) string {
	normalized := strings.ReplaceAll(input, "\u00a0", " ")
	normalized = whitespaceRegexp.ReplaceAllString(normalized, " ")

	return strings.TrimSpace(normalized)
}

func StripHTMLAndNormalizeFeedText(input string) string {
	unescaped := html.UnescapeString(input)
	withoutHTML := htmlTagRegexp.ReplaceAllString(unescaped, " ")

	return NormalizeFeedText(withoutHTML)
}
