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
	normalized := html.UnescapeString(input)
	normalized = strings.ReplaceAll(normalized, "\u00a0", " ")
	normalized = whitespaceRegexp.ReplaceAllString(normalized, " ")

	return strings.TrimSpace(normalized)
}

func StripHTMLAndNormalizeFeedText(input string) string {
	withoutHTML := htmlTagRegexp.ReplaceAllString(input, " ")

	return NormalizeFeedText(withoutHTML)
}
