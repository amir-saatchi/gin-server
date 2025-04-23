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

// GenerateFeasibility performs the attack feasibility analysis workflow.
func GenerateFeasibility(ctx context.Context, inputData similarity.InputData, systemInfo map[string]string, baseDataPath string) (similarity.DictResult, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" { return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set") }

	// Check required inputs
	if inputData.ThreatScenario == "" || inputData.AttackSteps == "" {
		return nil, fmt.Errorf("missing required input fields 'ThreatScenario' or 'AttackSteps' for feasibility workflow")
	}

	analysisType := config.FeasibilityAnalysis
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
	// Feasibility prompt asks for dictionary delimited by !!!! in step 6 [cite: 106, 107]
	// Let's try the !!!! delimiter first for this one.
	feasibilityResult, err := parseDictResponse(rawLLMResponse, "!!!!")
	if err != nil {
		// Fallback to #### just in case
		log.Printf("Failed parsing feasibility with '!!!!', trying '####': %v", err)
		feasibilityResult, err = parseDictResponse(rawLLMResponse, "####")
		if err != nil {
			log.Printf("Failed to parse Feasibility response dictionary using '!!!!' or '####': %v", err)
			return nil, fmt.Errorf("could not parse dictionary from LLM for feasibility")
		}
	}

	log.Printf("Workflow completed for: %s", analysisType)
	return feasibilityResult, nil
}