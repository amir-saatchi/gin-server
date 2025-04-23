// In corelogic/workflows/damage_workflow.go
package workflows

import (
	"context"
	"fmt"
	"log"
	"os"

	// "strings" // May only need utils now

	"github.com/amir-saatchi/rest-api/corelogic/config"
	"github.com/amir-saatchi/rest-api/corelogic/similarity"
	openai "github.com/sashabaranov/go-openai"
)

// GenerateDamageScenario performs the damage scenario analysis workflow.
func GenerateDamageScenario(ctx context.Context, inputData similarity.InputData, systemInfo map[string]string, baseDataPath string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" { return "", fmt.Errorf("OPENAI_API_KEY environment variable not set") }

	analysisType := config.DamageScenarioAnalysis
	log.Printf("Starting workflow for: %s", analysisType)

	// 1. Get Configuration
	cfg, err := config.GetConfig(analysisType, baseDataPath)
	if err != nil { return "", fmt.Errorf("workflow error getting config: %w", err) }

	// 2. Find Similar Shots
	shotsResult, err := similarity.FindTopKShotsFile(ctx, inputData, cfg.ReferenceDataJSONFile, 5)
	if err != nil {
		log.Printf("Warning: Error finding shots from file %s: %v. Proceeding without shots.", cfg.ReferenceDataJSONFile, err)
		shotsResult = &similarity.SimilarityResult{Shots: []similarity.ReferenceData{}}
	}

	// 3. Execute the core workflow steps using the helper
	llmModel := openai.GPT4o // Or get from config?
	rawLLMResponse, err := executeWorkflow(ctx, cfg, inputData, systemInfo, shotsResult, apiKey, llmModel)
	if err != nil {
		return "", fmt.Errorf("core workflow execution failed for %s: %w", analysisType, err)
	}

	// 4. Perform final parsing specific to this workflow
	// Damage scenario (validation) expects the result after the last '####'
	processedResponse, err := parseFinalStepResponse(rawLLMResponse, "####")
    if err != nil {
        log.Printf("Warning: could not parse final step for %s, returning raw response: %v", analysisType, err)
        processedResponse = rawLLMResponse // Fallback to raw
    }


	log.Printf("Workflow completed for: %s", analysisType)
	return processedResponse, nil
}