package workflows

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/amir-saatchi/rest-api/corelogic/config"
	"github.com/amir-saatchi/rest-api/corelogic/similarity"
	openai "github.com/sashabaranov/go-openai"
)

// GenerateAttackTree performs the attack tree generation workflow.
// Returns the raw ASCII attack tree string or error.
func GenerateAttackTree(ctx context.Context, inputData similarity.InputData, systemInfo map[string]string, baseDataPath string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" { return "", fmt.Errorf("OPENAI_API_KEY environment variable not set") }

	// Check required inputs
	if inputData.ThreatScenario == "" || inputData.AttackVector == "" {
		return "", fmt.Errorf("missing required input fields 'ThreatScenario' or 'AttackVector' for attack tree workflow")
	}

	analysisType := config.AttackTreeAnalysis
	log.Printf("Starting workflow for: %s", analysisType)

	// 1. Get Configuration
	cfg, err := config.GetConfig(analysisType, baseDataPath)
	if err != nil { return "", fmt.Errorf("workflow error getting config: %w", err) }

	// 2. Find Similar Shots (Not needed for attack tree based on prompt analysis)
	// Attack tree prompt doesn't have {shots} placeholder [cite: 109]
	var shotsResult *similarity.SimilarityResult = nil // Pass nil if no shots needed

	// 3. Execute the core workflow steps using the helper
	llmModel := openai.GPT4o // Or get from config?
	rawLLMResponse, err := executeWorkflow(ctx, cfg, inputData, systemInfo, shotsResult, apiKey, llmModel)
	if err != nil {
		return "", fmt.Errorf("core workflow execution failed for %s: %w", analysisType, err)
	}

	// 4. Perform final parsing specific to this workflow
	// Attack tree prompt asks for ASCII format[cite: 122], no specific delimiter mentioned.
	// Return the raw response directly.
	processedResponse := rawLLMResponse

	log.Printf("Workflow completed for: %s", analysisType)
	return processedResponse, nil
}