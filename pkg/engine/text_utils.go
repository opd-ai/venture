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
		// Invalid width, return single line
		return []string{text}
	}

	lines := []string{}
	words := strings.Fields(text) // Split by whitespace
	currentLine := ""

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		// Measure text width
		width := MeasureText(testLine, face)

		if width <= maxWidth {
			// Word fits, add to current line
			currentLine = testLine
		} else {
			// Word doesn't fit
			if currentLine != "" {
				// Save current line and start new one
				lines = append(lines, currentLine)
			}

			// Check if single word is too long for maxWidth
			wordWidth := MeasureText(word, face)
			if wordWidth > maxWidth {
				// Hyphenate very long word
				hyphenatedLines := hyphenateWord(word, maxWidth, face)
				for i, line := range hyphenatedLines {
					if i < len(hyphenatedLines)-1 {
						// All but last line get appended immediately
						lines = append(lines, line)
					} else {
						// Last line becomes new current line
						currentLine = line
					}
				}
			} else {
				// Word fits on its own line
				currentLine = word
			}
		}
	}

	// Add final line if any
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	// Return at least empty array, never nil
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
	// fixed.Int26_6 uses 26 bits for fractional part (6 bits), so divide by 64
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
