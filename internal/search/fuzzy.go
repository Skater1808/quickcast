package search

import (
	"strings"
	"unicode"
)

// MatchResult holds the score and character indices that matched.
type MatchResult struct {
	Matched bool
	Score   float64
	Indices []int
}

// FuzzyMatch performs fuzzy matching of pattern against target.
// Returns MatchResult with score and matched index positions.
func FuzzyMatch(pattern, target string) MatchResult {
	if pattern == "" {
		return MatchResult{Matched: true, Score: 1.0, Indices: nil}
	}

	patternRunes := []rune(strings.ToLower(pattern))
	targetRunes := []rune(strings.ToLower(target))
	origTargetRunes := []rune(target)

	pLen := len(patternRunes)
	tLen := len(targetRunes)

	if pLen > tLen {
		return MatchResult{Matched: false, Score: 0}
	}

	// Exact match gets astronomical boost
	if strings.ToLower(pattern) == strings.ToLower(target) {
		indices := make([]int, pLen)
		for i := 0; i < pLen; i++ {
			indices[i] = i
		}
		return MatchResult{Matched: true, Score: 1000.0, Indices: indices}
	}

	// Substring match prefix gets huge boost
	if strings.HasPrefix(strings.ToLower(target), strings.ToLower(pattern)) {
		indices := make([]int, pLen)
		for i := 0; i < pLen; i++ {
			indices[i] = i
		}
		score := 500.0 + float64(pLen)/float64(tLen)*100.0
		return MatchResult{Matched: true, Score: score, Indices: indices}
	}

	// General fuzzy search
	pIdx := 0
	score := 0.0
	indices := make([]int, 0, pLen)
	consecutiveStreak := 0
	prevMatchedIdx := -2

	for tIdx := 0; tIdx < tLen && pIdx < pLen; tIdx++ {
		if patternRunes[pIdx] == targetRunes[tIdx] {
			indices = append(indices, tIdx)

			// Base point for char match
			charScore := 10.0

			// Consecutive match bonus
			if tIdx == prevMatchedIdx+1 {
				consecutiveStreak++
				charScore += float64(consecutiveStreak) * 15.0
			} else {
				consecutiveStreak = 0
			}

			// Word boundary bonus (start of string or preceded by non-letter)
			if tIdx == 0 {
				charScore += 40.0
			} else {
				prevChar := origTargetRunes[tIdx-1]
				currChar := origTargetRunes[tIdx]
				if !unicode.IsLetter(prevChar) && !unicode.IsDigit(prevChar) {
					charScore += 35.0
				} else if unicode.IsLower(prevChar) && unicode.IsUpper(currChar) {
					// CamelCase boundary
					charScore += 30.0
				}
			}

			score += charScore
			prevMatchedIdx = tIdx
			pIdx++
		}
	}

	if pIdx < pLen {
		return MatchResult{Matched: false, Score: 0}
	}

	// Penalty for longer target lengths to prefer tighter matches
	lengthPenalty := float64(tLen - pLen) * 0.5
	score -= lengthPenalty

	return MatchResult{
		Matched: true,
		Score:   score,
		Indices: indices,
	}
}
