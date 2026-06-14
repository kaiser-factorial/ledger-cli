package ui

import (
	"fmt"
	"strings"
)

// SparkChars are the characters used for sparklines (low to high)
var SparkChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// Sparkline generates a simple ASCII sparkline from a slice of values.
// It returns a string of Unicode block characters.
func Sparkline(values []int) string {
	if len(values) == 0 {
		return ""
	}

	max := 0
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	if max == 0 {
		return strings.Repeat(string(SparkChars[0]), len(values))
	}

	var sb strings.Builder
	for _, v := range values {
		idx := (v * (len(SparkChars) - 1)) / max
		if idx < 0 {
			idx = 0
		}
		if idx >= len(SparkChars) {
			idx = len(SparkChars) - 1
		}
		sb.WriteRune(SparkChars[idx])
	}
	return sb.String()
}

// SparklineWithLegend returns a sparkline with a small legend
func SparklineWithLegend(values []int, label string) string {
	if len(values) == 0 {
		return ""
	}
	spark := Sparkline(values)
	total := 0
	for _, v := range values {
		total += v
	}
	return fmt.Sprintf("%s %s (total: %d)", spark, label, total)
}