package services

import (
	"regexp"
	"sort"
	"strings"
)

type RankingService struct{}

type RankedResult struct {
	Content SearchResult
	Score   float64
}

func NewRankingService() *RankingService {
	return &RankingService{}
}

// Simple keyword-based reranking - in production you'd use a proper reranking model
func (rs *RankingService) RerankResults(query string, results []SearchResult, topK int) []SearchResult {
	if len(results) == 0 {
		return results
	}

	// Calculate relevance scores
	var rankedResults []RankedResult
	queryTerms := extractKeywords(strings.ToLower(query))

	for _, result := range results {
		score := calculateRelevanceScore(queryTerms, result)
		rankedResults = append(rankedResults, RankedResult{
			Content: result,
			Score:   score,
		})
	}

	// Sort by score (descending)
	sort.Slice(rankedResults, func(i, j int) bool {
		return rankedResults[i].Score > rankedResults[j].Score
	})

	// Return top K results
	if topK > len(rankedResults) {
		topK = len(rankedResults)
	}

	var topResults []SearchResult
	for i := 0; i < topK; i++ {
		topResults = append(topResults, rankedResults[i].Content)
	}

	return topResults
}

func extractKeywords(text string) []string {
	// Remove common stop words and extract meaningful terms
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true, "do": true,
		"does": true, "did": true, "will": true, "would": true, "could": true, "should": true,
		"what": true, "when": true, "where": true, "who": true, "why": true, "how": true,
	}

	// Extract words (alphanumeric sequences)
	re := regexp.MustCompile(`\b\w+\b`)
	words := re.FindAllString(strings.ToLower(text), -1)

	var keywords []string
	for _, word := range words {
		if len(word) > 2 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

func calculateRelevanceScore(queryTerms []string, result SearchResult) float64 {
	score := 0.0
	
	// Combine title and content for scoring
	text := strings.ToLower(result.Title + " " + result.Content)
	
	// Calculate term frequency and position-based scoring
	for _, term := range queryTerms {
		// Title matches get higher weight
		titleMatches := strings.Count(strings.ToLower(result.Title), term)
		score += float64(titleMatches) * 2.0
		
		// Content matches
		contentMatches := strings.Count(strings.ToLower(result.Content), term)
		score += float64(contentMatches) * 1.0
		
		// Exact phrase bonus (if the term appears as part of a larger phrase)
		if strings.Contains(text, term) {
			score += 0.5
		}
	}
	
	// Normalize by text length to avoid bias toward longer documents
	textLength := float64(len(text))
	if textLength > 0 {
		score = score / (textLength / 100.0) // Normalize per 100 characters
	}
	
	return score
}