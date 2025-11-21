// Package engine provides text utility functions for UI rendering.
// This file implements text wrapping and measurement for procedurally generated content.
package engine

import (
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// WrapText wraps text to fit within maxWidth pixels using the given font face.
// Returns array of lines that fit within the width constraint.
// Handles word wrapping at word boundaries and hyphenation for very long words.
func WrapText(text string, maxWidth int, face font.Face) []string {
	if text == "" {
		return []string{}
	}
	if maxWidth <= 0 {
		return []string{text}
	}

	lines := []string{}
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		newLine, overflow := tryAddWordToLine(currentLine, word, maxWidth, face)
		if !overflow {
			currentLine = newLine
		} else {
			lines = appendCurrentLine(lines, currentLine)
			currentLine = handleOverflowWord(word, maxWidth, face, &lines)
		}
	}

	lines = appendCurrentLine(lines, currentLine)
	return ensureNonNilSlice(lines)
}

// tryAddWordToLine attempts to add a word to the current line.
// Returns the new line and whether it overflows the max width.
func tryAddWordToLine(currentLine, word string, maxWidth int, face font.Face) (string, bool) {
	testLine := buildTestLine(currentLine, word)
	width := MeasureText(testLine, face)
	return testLine, width > maxWidth
}

// buildTestLine builds a test line by adding a word to the current line.
func buildTestLine(currentLine, word string) string {
	if currentLine == "" {
		return word
	}
	return currentLine + " " + word
}

// handleOverflowWord handles a word that doesn't fit on the current line.
// Returns the new current line after processing the overflow.
func handleOverflowWord(word string, maxWidth int, face font.Face, lines *[]string) string {
	wordWidth := MeasureText(word, face)
	if wordWidth > maxWidth {
		return hyphenateAndAppend(word, maxWidth, face, lines)
	}
	return word
}

// hyphenateAndAppend hyphenates a long word and appends all but the last segment.
func hyphenateAndAppend(word string, maxWidth int, face font.Face, lines *[]string) string {
	hyphenatedLines := hyphenateWord(word, maxWidth, face)
	for i := 0; i < len(hyphenatedLines)-1; i++ {
		*lines = append(*lines, hyphenatedLines[i])
	}
	if len(hyphenatedLines) > 0 {
		return hyphenatedLines[len(hyphenatedLines)-1]
	}
	return ""
}

// appendCurrentLine appends the current line to lines if non-empty.
func appendCurrentLine(lines []string, currentLine string) []string {
	if currentLine != "" {
		return append(lines, currentLine)
	}
	return lines
}

// ensureNonNilSlice ensures the slice is non-nil.
func ensureNonNilSlice(lines []string) []string {
	if len(lines) == 0 {
		return []string{}
	}
	return lines
}

// hyphenateWord breaks a very long word across multiple lines with hyphens.
// Returns array of lines where all but the last end with a hyphen.
func hyphenateWord(word string, maxWidth int, face font.Face) []string {
	if word == "" {
		return []string{}
	}

	lines := []string{}
	current := ""

	for _, ch := range word {
		test := current + string(ch)
		testWithHyphen := test + "-"

		// Check if adding this character (plus hyphen) exceeds width
		width := MeasureText(testWithHyphen, face)

		if width > maxWidth && current != "" {
			// Break here, add hyphen to current
			lines = append(lines, current+"-")
			current = string(ch)
		} else {
			// Add character to current
			current = test
		}
	}

	// Add final part (no hyphen on last segment)
	if current != "" {
		lines = append(lines, current)
	}

	// Edge case: if even a single character + hyphen exceeds width,
	// we still need to return something
	if len(lines) == 0 {
		return []string{word}
	}

	return lines
}

// MeasureText measures the width in pixels of text rendered with the given font face.
// Returns the width as an integer pixel value.
func MeasureText(text string, face font.Face) int {
	if text == "" {
		return 0
	}

	// Use font.MeasureString which returns fixed.Int26_6
	bounds := font.MeasureString(face, text)

	// Convert from fixed.Int26_6 to pixels
	// fixed.Int26_6 represents fixed-point values with 26 integer bits and 6 fractional bits, so divide by 64 to get pixel units
	width := int(bounds / fixed.I(1))

	return width
}

// WrapTextWithHeight wraps text and calculates the total height needed.
// Returns the wrapped lines and the total height in pixels needed to render them.
// Useful for dynamic tooltip sizing.
func WrapTextWithHeight(text string, maxWidth int, face font.Face, lineSpacing int) ([]string, int) {
	lines := WrapText(text, maxWidth, face)

	if len(lines) == 0 {
		return lines, 0
	}

	// Calculate height based on number of lines
	// Each line needs the font height plus line spacing (except the last line)
	metrics := face.Metrics()
	lineHeight := int(metrics.Height / fixed.I(1))

	totalHeight := lineHeight * len(lines)
	if len(lines) > 1 {
		totalHeight += lineSpacing * (len(lines) - 1)
	}

	return lines, totalHeight
}
