package services

import (
	"regexp"
	"strings"
)

type scoredSentence struct {
	Text  string
	Score float64
	Index int
}

// ExtractPlainText extracts plain text from HTML content
func ExtractPlainText(html string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(html, " ")
	
	// Replace HTML entities
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	
	// Normalize whitespace
	re = regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")
	
	return strings.TrimSpace(text)
}

// TruncateText truncates text to a maximum number of characters
func TruncateText(text string, maxChars int) string {
	if len(text) <= maxChars {
		return text
	}
	
	// Try to break at word boundary
	truncated := text[:maxChars]
	lastSpace := strings.LastIndex(truncated, " ")
	
	if lastSpace > maxChars/2 { // Only break at word if it's not too far back
		return truncated[:lastSpace] + "..."
	}
	
	return truncated + "..."
}

// ExtractRelevantSections extracts text sections that contain query keywords
func ExtractRelevantSections(text, query string, maxChars int) string {
	if len(text) <= maxChars {
		return text
	}
	
	// Extract keywords from query
	keywords := extractQueryKeywords(query)
	if len(keywords) == 0 {
		// Fall back to simple truncation if no keywords
		return TruncateText(text, maxChars)
	}
	
	// Split text into sentences
	sentences := splitIntoSentences(text)
	
	// Score each sentence based on keyword matches
	
	var scoredSentences []scoredSentence
	for i, sentence := range sentences {
		score := scoreSentence(sentence, keywords)
		scoredSentences = append(scoredSentences, scoredSentence{
			Text:  sentence,
			Score: score,
			Index: i,
		})
	}
	
	// Sort by score (descending)
	sortSentencesByScore(scoredSentences)
	
	// Build result by adding highest-scoring sentences until we hit char limit
	var result strings.Builder
	var totalChars int
	usedSentences := make(map[int]bool)
	
	for _, scored := range scoredSentences {
		sentenceLen := len(scored.Text)
		if totalChars + sentenceLen + 2 <= maxChars { // +2 for spacing
			result.WriteString(scored.Text)
			result.WriteString(" ")
			totalChars += sentenceLen + 1
			usedSentences[scored.Index] = true
		}
	}
	
	extractedText := strings.TrimSpace(result.String())
	
	// If we got very little content, fall back to beginning of document
	if len(extractedText) < maxChars/3 {
		return TruncateText(text, maxChars)
	}
	
	return extractedText
}

// Helper functions
func extractQueryKeywords(query string) []string {
	// Remove common stop words and extract meaningful terms
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true, "do": true,
		"does": true, "did": true, "will": true, "would": true, "could": true, "should": true,
		"what": true, "when": true, "where": true, "who": true, "why": true, "how": true,
	}

	words := strings.Fields(strings.ToLower(query))
	var keywords []string
	
	for _, word := range words {
		// Clean word of punctuation
		cleaned := regexp.MustCompile(`[^\w]`).ReplaceAllString(word, "")
		if len(cleaned) > 2 && !stopWords[cleaned] {
			keywords = append(keywords, cleaned)
		}
	}
	
	return keywords
}

func splitIntoSentences(text string) []string {
	// Simple sentence splitting - could be improved with more sophisticated NLP
	sentences := regexp.MustCompile(`[.!?]+\s+`).Split(text, -1)
	
	var cleanSentences []string
	for _, sentence := range sentences {
		trimmed := strings.TrimSpace(sentence)
		if len(trimmed) > 10 { // Ignore very short fragments
			cleanSentences = append(cleanSentences, trimmed)
		}
	}
	
	return cleanSentences
}

func scoreSentence(sentence string, keywords []string) float64 {
	lowerSentence := strings.ToLower(sentence)
	score := 0.0
	
	for _, keyword := range keywords {
		// Count occurrences of keyword
		count := strings.Count(lowerSentence, keyword)
		score += float64(count)
		
		// Bonus for exact matches
		if strings.Contains(lowerSentence, keyword) {
			score += 0.5
		}
	}
	
	// Normalize by sentence length to avoid bias toward long sentences
	if len(sentence) > 0 {
		score = score / (float64(len(sentence)) / 100.0)
	}
	
	return score
}

func sortSentencesByScore(sentences []scoredSentence) {
	for i := 0; i < len(sentences)-1; i++ {
		for j := i + 1; j < len(sentences); j++ {
			if sentences[i].Score < sentences[j].Score {
				sentences[i], sentences[j] = sentences[j], sentences[i]
			}
		}
	}
}