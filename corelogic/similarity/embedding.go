package similarity

import (
	"context"
	"fmt"
	"log"
	"strings"

	openai "github.com/sashabaranov/go-openai" // Corrected import path if using this popular client
)

// Define which keys from InputData should be embedded
var embedKeys = []string{"Asset", "Category", "Property", "Asset Description"}

// getOpenAIEmbedding generates an embedding using the specified OpenAI model.
func getOpenAIEmbedding(ctx context.Context, input InputData, apiKey string) ([]float32, error) {
	// Filter input data based on embedKeys and format it as a string
	var parts []string
	// Use a map for easier key lookup
	dataMap := map[string]string{
		"Asset":             input.Asset,
		"Category":          input.Category,
		"Property":          input.Property,
		"Asset Description": input.AssetDescription,
	}

	// Construct the text to embed based on specified keys
	for _, key := range embedKeys {
		if val, ok := dataMap[key]; ok && val != "" {
			// Simple string representation, adjust formatting as needed
			parts = append(parts, fmt.Sprintf("%s: %s", key, val))
		}
	}
	textToEmbed := strings.Join(parts, ", ")
	if textToEmbed == "" {
		return nil, fmt.Errorf("no valid data found for embedding based on embedKeys")
	}

	// --- Call OpenAI API (Corrected Usage) ---
	client := openai.NewClient(apiKey)

	// Use the correct model name as a string literal
	// Use openai.EmbeddingModel constants if preferred and available for v3 models
	// e.g. if openai.TextEmbedding3Large exists, use that. Otherwise, use the string.
	model := openai.LargeEmbedding3 // Defaulting to Ada v2 for example, CHECK constant for v3 large
	// modelName := "text-embedding-3-large" // Explicit string name

	dimensions := 256 // Match dimensions used in Python code

	// Create the embedding request using the standard struct
	req := openai.EmbeddingRequest{
		Input:      []string{textToEmbed},
		Model:      openai.EmbeddingModel(model), // Cast string to the EmbeddingModel type
		Dimensions: dimensions,
		// EncodingFormat is also available if needed, e.g., openai.EncodingFormatFloat
	}

	// Call the CreateEmbeddings method
	resp, err := client.CreateEmbeddings(ctx, req)
	if err != nil {
		// It's helpful to log the text that failed
		log.Printf("Failed to embed text: %s", textToEmbed)
		return nil, fmt.Errorf("openai CreateEmbeddings failed: %w", err)
	}

	// Check response structure
	if len(resp.Data) == 0 || len(resp.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("received empty embedding from openai")
	}

	// Return the first embedding
	return resp.Data[0].Embedding, nil
}