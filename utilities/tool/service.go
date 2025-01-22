package tool

import (
	"bytes"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func (s *service) DeaccentVietnameseString(value string) string {
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, "đ", "d")
	valueBytes := make([]byte, len(value))

	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	_, _, _ = t.Transform(valueBytes, []byte(value), true)
	return string(bytes.TrimRight(valueBytes, "\x00"))
}

func (s *service) GetStartOfDate(date time.Time) time.Time {
	year, month, day := date.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, date.Location())
}

func (s *service) GetEndOfDate(date time.Time) time.Time {
	year, month, day := date.Date()
	return time.Date(year, month, day, 23, 59, 59, 999, date.Location())
}
