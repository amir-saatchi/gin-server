package similarity

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
)

// --- Cosine Similarity Calculation ---

func dotProduct(a, b []float32) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vectors must have the same dimension")
	}
	var sum float64 = 0
	for i := range a {
		sum += float64(a[i]) * float64(b[i])
	}
	return sum, nil
}

func magnitude(vec []float32) float64 {
	var sumSq float64 = 0
	for _, val := range vec {
		sumSq += float64(val) * float64(val)
	}
	return math.Sqrt(sumSq)
}

func cosineSimilarity(a, b []float32) (float64, error) {
	dot, err := dotProduct(a, b)
	if err != nil {
		return 0, err
	}

	magA := magnitude(a)
	magB := magnitude(b)

	if magA == 0 || magB == 0 {
		// Handle zero vectors - similarity is arguably 0 or undefined.
		// Returning 0 is common practice in some libraries.
		return 0, nil
		// Alternatively: return 0, fmt.Errorf("cannot compute similarity with zero vector")
	}

	similarity := dot / (magA * magB)
	// Clamp similarity to [-1, 1] due to potential floating point inaccuracies
	return math.Max(-1.0, math.Min(1.0, similarity)), nil
}

// --- Data Loading ---

// loadReferenceData reads reference data from a JSON file.
// Assumes the JSON file contains a list of ReferenceData objects.
func loadReferenceData(filePath string) ([]ReferenceData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read reference data file %s: %w", filePath, err)
	}

	var referenceItems []ReferenceData
	err = json.Unmarshal(data, &referenceItems)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal reference data from %s: %w", filePath, err)
	}
	return referenceItems, nil
}

// --- Main Similarity Search Function ---

// FindTopKShotsFile finds the top K similar items from a reference data file.
func FindTopKShotsFile(ctx context.Context, input InputData, referenceDataPath string, topK int) (*SimilarityResult, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	// 1. Get Input Embedding
	inputEmbedding, err := getOpenAIEmbedding(ctx, input, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get input embedding: %w", err)
	}
	log.Println("Successfully retrieved input embedding")

	// 2. Load Reference Data from File
	// NOTE: 'referenceDataPath' based on the task type (e.g., damage scenario)
	referenceItems, err := loadReferenceData(referenceDataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load reference data: %w", err)
	}
	if len(referenceItems) == 0 {
		return nil, fmt.Errorf("no reference items loaded from %s", referenceDataPath)
	}
	log.Printf("Loaded %d reference items from %s", len(referenceItems), referenceDataPath)

	// 3. Calculate Similarities
	resultsWithScores := make([]ResultWithScore, 0, len(referenceItems))
	for _, item := range referenceItems {
		similarity, err := cosineSimilarity(inputEmbedding, item.Embedding)
		if err != nil {
			log.Printf("Warning: could not calculate similarity for item %s: %v", item.ID, err)
			continue // Skip item if similarity calculation fails
		}
		resultsWithScores = append(resultsWithScores, ResultWithScore{Data: item, Score: similarity})
	}

	// 4. Sort by Similarity Score (Descending)
	sort.Slice(resultsWithScores, func(i, j int) bool {
		return resultsWithScores[i].Score > resultsWithScores[j].Score // Higher score first
	})

	// 5. Get Top K Results
	actualTopK := topK
	if len(resultsWithScores) < topK {
		actualTopK = len(resultsWithScores)
	}

	finalShots := make([]ReferenceData, actualTopK)
	for i := 0; i < actualTopK; i++ {
		finalShots[i] = resultsWithScores[i].Data
		// Optionally add the score back to the result if needed elsewhere
		// finalShots[i].SimilarityScore = resultsWithScores[i].Score
	}

	log.Printf("Returning top %d shots", actualTopK)
	result := &SimilarityResult{
		Shots: finalShots,
	}

	return result, nil
}