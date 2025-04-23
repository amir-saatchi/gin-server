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

// GenerateAttackSteps performs the attack steps analysis workflow.
func GenerateAttackSteps(ctx context.Context, inputData similarity.InputData, systemInfo map[string]string, baseDataPath string) (similarity.DictResult, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" { return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set") }

	// Check for required input specific to this workflow
	if inputData.AttackVector == "" {
		return nil, fmt.Errorf("missing required input field 'AttackVector' for attack steps workflow")
	}
	// Validation prompt also needs threat scenario
	if inputData.ThreatScenario == "" {
		return nil, fmt.Errorf("missing required input field 'ThreatScenario' for attack steps validation workflow")
	}
    // Validation prompt also needs threat
	if inputData.Threat == "" {
		return nil, fmt.Errorf("missing required input field 'Threat' for attack steps validation workflow")
	}


	analysisType := config.AttackStepsAnalysis
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
	// Note: executeWorkflow handles the BASE vs VALIDATE logic internally
	llmModel := openai.GPT4o // Or get from config?
	rawLLMResponse, err := executeWorkflow(ctx, cfg, inputData, systemInfo, shotsResult, apiKey, llmModel)
	if err != nil {
		return nil, fmt.Errorf("core workflow execution failed for %s: %w", analysisType, err)
	}

	// 4. Perform final parsing specific to this workflow
	// Attack steps validation prompt asks for a dictionary in the last step [cite: 83]
	attackStepsResult, err := parseDictResponse(rawLLMResponse, "####")
	if err != nil {
		log.Printf("Failed to parse Attack Steps response dictionary: %v", err)
		return nil, fmt.Errorf("could not parse dictionary from LLM for attack steps")
	}

	log.Printf("Workflow completed for: %s", analysisType)
	return attackStepsResult, nil
}