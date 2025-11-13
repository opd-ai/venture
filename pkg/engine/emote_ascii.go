package engine

import (
	"fmt"
	"math/rand"
)

// EmoteASCII generates procedural ASCII art for chat emotes.
// Each expression type has multiple variations generated from a seed.
type EmoteASCII struct {
	// Art is the ASCII art string (multi-line)
	Art string
	// Width is the character width
	Width int
	// Height is the number of lines
	Height int
}

// GenerateASCIIEmote creates procedural ASCII art for an expression type.
// Uses seed for deterministic generation so same seed produces same art.
func GenerateASCIIEmote(expressionType ExpressionType, seed int64) *EmoteASCII {
	rng := rand.New(rand.NewSource(seed))
	
	switch expressionType {
	case ExpressionWave:
		return generateWaveEmote(rng)
	case ExpressionCheer:
		return generateCheerEmote(rng)
	case ExpressionDance:
		return generateDanceEmote(rng)
	case ExpressionLaugh:
		return generateLaughEmote(rng)
	case ExpressionCry:
		return generateCryEmote(rng)
	case ExpressionSit:
		return generateSitEmote(rng)
	case ExpressionPoint:
		return generatePointEmote(rng)
	case ExpressionSalute:
		return generateSaluteEmote(rng)
	case ExpressionShrug:
		return generateShrugEmote(rng)
	case ExpressionThumbsUp:
		return generateThumbsUpEmote(rng)
	case ExpressionFacepalm:
		return generateFacepalmEmote(rng)
	case ExpressionSleep:
		return generateSleepEmote(rng)
	default:
		return &EmoteASCII{Art: "¯\\_(ツ)_/¯", Width: 9, Height: 1}
	}
}

// generateWaveEmote creates a waving ASCII art
func generateWaveEmote(rng *rand.Rand) *EmoteASCII {
	variants := []string{
		"  o/\n |/\n/ \\",
		"\\o/\n |\n/ \\",
		"  o_/\n |/\n/ \\",
	}
	art := variants[rng.Intn(len(variants))]
	return &EmoteASCII{Art: art, Width: 3, Height: 3}
}

// generateCheerEmote creates a cheering ASCII art
func generateCheerEmote(rng *rand.Rand) *EmoteASCII {
	variants := []string{
		"\\o/\n |\n/ \\",
		"\\O/\n <\n/ \\",
		"\\o/!!\n |\n/ \\",
	}
	art := variants[rng.Intn(len(variants))]
	return &EmoteASCII{Art: art, Width: 5, Height: 3}
}

// generateDanceEmote creates a dancing ASCII art
func generateDanceEmote(rng *rand.Rand) *EmoteASCII {
	variants := []string{
		"~o~\n<|>\n/ \\",
		"♪o♪\n<|>\n >",
		"♫o♫\n |/\n/ <",
	}
	art := variants[rng.Intn(len(variants))]
	return &EmoteASCII{Art: art, Width: 3, Height: 3}
}

// generateLaughEmote creates a laughing ASCII art
func generateLaughEmote(rng *rand.Rand) *EmoteASCII {
	variants := []string{
		"XD",
		"(≧▽≦)",
		"ヾ(≧▽≦*)o",
		"(^▽^)",
	}
	art := variants[rng.Intn(len(variants))]
	height := 1
	for i := 0; i < len(art); i++ {
		if art[i] == '\n' {
			height++
		}
	}
	return &EmoteASCII{Art: art, Width: len(art), Height: height}
}

// generateCryEmote creates a crying ASCII art
func generateCryEmote(rng *rand.Rand) *EmoteASCII {
	variants := []string{
		"(╥﹏╥)",
		"(T_T)",
		"(ToT)",
		"｡゜(｀Д´)゜｡",
	}
	art := variants[rng.Intn(len(variants))]
	height := 1
	for i := 0; i < len(art); i++ {
		if art[i] == '\n' {
			height++
		}
	}
	return &EmoteASCII{Art: art, Width: len(art), Height: height}
}

// generateSitEmote creates a sitting ASCII art
func generateSitEmote(rng *rand.Rand) *EmoteASCII {
	variants := []string{
		"_o_\n| |\n===",
		"_O_\n | \n===",
	}
	art := variants[rng.Intn(len(variants))]
	return &EmoteASCII{Art: art, Width: 3, Height: 3}
}

// generatePointEmote creates a pointing ASCII art
func generatePointEmote(rng *rand.Rand) *EmoteASCII {
	variants := []string{
		" o\n-|\n/ \\",
		" o→\n |\n/ \\",
		" o->\n |\n/ \\",
	}
	art := variants[rng.Intn(len(variants))]
	return &EmoteASCII{Art: art, Width: 3, Height: 3}
}

// generateSaluteEmote creates a saluting ASCII art
func generateSaluteEmote(rng *rand.Rand) *EmoteASCII {
	variants := []string{
		" o7\n |\n/ \\",
		"(o7)\n |\n/ \\",
	}
	art := variants[rng.Intn(len(variants))]
	return &EmoteASCII{Art: art, Width: 3, Height: 3}
}

// generateShrugEmote creates a shrugging ASCII art
func generateShrugEmote(rng *rand.Rand) *EmoteASCII {
	variants := []string{
		"¯\\_(ツ)_/¯",
		"¯\\(◉◡◔)/¯",
		"¯\\_(⊙_ʖ⊙)_/¯",
	}
	art := variants[rng.Intn(len(variants))]
	height := 1
	for i := 0; i < len(art); i++ {
		if art[i] == '\n' {
			height++
		}
	}
	return &EmoteASCII{Art: art, Width: len(art), Height: height}
}

// generateThumbsUpEmote creates a thumbs up ASCII art
func generateThumbsUpEmote(rng *rand.Rand) *EmoteASCII {
	variants := []string{
		"(b^_^)b",
		"(^_^)b",
		"(°ロ°)☝",
		"d(^_^)b",
	}
	art := variants[rng.Intn(len(variants))]
	height := 1
	for i := 0; i < len(art); i++ {
		if art[i] == '\n' {
			height++
		}
	}
	return &EmoteASCII{Art: art, Width: len(art), Height: height}
}

// generateFacepalmEmote creates a facepalm ASCII art
func generateFacepalmEmote(rng *rand.Rand) *EmoteASCII {
	variants := []string{
		"(－‸ლ)",
		"(￢_￢;)",
		"(-_-;)",
		"(╯°□°）╯",
	}
	art := variants[rng.Intn(len(variants))]
	height := 1
	for i := 0; i < len(art); i++ {
		if art[i] == '\n' {
			height++
		}
	}
	return &EmoteASCII{Art: art, Width: len(art), Height: height}
}

// generateSleepEmote creates a sleeping ASCII art
func generateSleepEmote(rng *rand.Rand) *EmoteASCII {
	variants := []string{
		"(-_-)zzz",
		"(u_u) zzZ",
		"(-.-)Zzz",
	}
	art := variants[rng.Intn(len(variants))]
	height := 1
	for i := 0; i < len(art); i++ {
		if art[i] == '\n' {
			height++
		}
	}
	return &EmoteASCII{Art: art, Width: len(art), Height: height}
}

// FormatForChat formats the emote for display in chat.
// Single-line emotes are returned as-is, multi-line emotes are formatted with prefix.
func (e *EmoteASCII) FormatForChat(prefix string) string {
	if e.Height == 1 {
		return e.Art
	}
	// Multi-line: add prefix to each line
	result := ""
	currentLine := ""
	for i := 0; i < len(e.Art); i++ {
		if e.Art[i] == '\n' {
			result += prefix + currentLine + "\n"
			currentLine = ""
		} else {
			currentLine += string(e.Art[i])
		}
	}
	if currentLine != "" {
		result += prefix + currentLine
	}
	return result
}

// String returns the ASCII art as a string.
func (e *EmoteASCII) String() string {
	return fmt.Sprintf("EmoteASCII{Width: %d, Height: %d, Art: %s}", e.Width, e.Height, e.Art)
}
