package workflows

import (
	"context"
	"fmt"
	"log"
	"os"

	// "strconv" // Keep commented unless needed for type conversion

	"github.com/amir-saatchi/rest-api/corelogic/config"
	"github.com/amir-saatchi/rest-api/corelogic/similarity"
	openai "github.com/sashabaranov/go-openai"
)

// GenerateImpactScores performs the impact score analysis workflow.
func GenerateImpactScores(ctx context.Context, inputData similarity.InputData, systemInfo map[string]string, baseDataPath string) (similarity.DictResult, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" { return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set") }

	analysisType := config.ImpactScoresAnalysis
	log.Printf("Starting workflow for: %s", analysisType)

	// 1. Get Configuration
	cfg, err := config.GetConfig(analysisType, baseDataPath)
	if err != nil { return nil, fmt.Errorf("workflow error getting config: %w", err) }

	// 2. Find Similar Shots
	shotsResult, err := similarity.FindTopKShotsFile(ctx, inputData, cfg.ReferenceDataJSONFile, 5)
	if err != nil {
		log.Printf("Warning: Error finding shots for %s: %v. Proceeding without shots.", analysisType, err)
		shotsResult = &similarity.SimilarityResult{Shots: []similarity.ReferenceData{}}
	}

	// 3. Execute the core workflow steps using the helper
	llmModel := openai.GPT4o // Or get from config?
	rawLLMResponse, err := executeWorkflow(ctx, cfg, inputData, systemInfo, shotsResult, apiKey, llmModel)
	if err != nil {
		return nil, fmt.Errorf("core workflow execution failed for %s: %w", analysisType, err)
	}

	// 4. Perform final parsing specific to this workflow
	// Impact scores prompt asks for a dictionary in the last step [cite: 42]
	impactScores, err := parseDictResponse(rawLLMResponse, "####")
	if err != nil {
		log.Printf("Failed to parse Impact Scores response dictionary: %v", err)
		return nil, fmt.Errorf("could not parse dictionary from LLM for impact scores")
	}

	// Optional: Validate/convert score types here if needed before returning map[string]interface{}

	log.Printf("Workflow completed for: %s", analysisType)
	return impactScores, nil
}