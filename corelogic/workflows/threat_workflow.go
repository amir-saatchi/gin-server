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

// GenerateThreatScenario performs the threat scenario analysis workflow.
func GenerateThreatScenario(ctx context.Context, inputData similarity.InputData, systemInfo map[string]string, baseDataPath string) (*similarity.ThreatScenarioResult, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" { return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set") }

	analysisType := config.ThreatScenarioAnalysis
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
	// Threat scenario prompt asks for a dictionary in the last step [cite: 66]
	parsedDict, err := parseDictResponse(rawLLMResponse, "####")
	if err != nil {
		log.Printf("Failed to parse Threat Scenario response dictionary: %v", err)
		return nil, fmt.Errorf("could not parse dictionary from LLM for threat scenario")
	}

	// Convert parsed map to struct
	result := &similarity.ThreatScenarioResult{}
	var parseOk bool
	if ts, ok := parsedDict["threat_scenario"].(string); ok {
		result.ThreatScenario = ts
		parseOk = true
	} else {
		log.Printf("Warning: 'threat_scenario' key missing or not a string in response dict")
	}

	if av, ok := parsedDict["attack_vectors"].([]interface{}); ok {
		for _, vector := range av {
			if vecStr, ok := vector.(string); ok {
				result.AttackVectors = append(result.AttackVectors, vecStr)
				parseOk = true // Mark as OK if at least attack vectors were parsed
			}
		}
	} else {
		log.Printf("Warning: 'attack_vectors' key missing or not a list in response dict")
	}

	if !parseOk {
		// If neither key was found or valid, return a more specific error or the raw response perhaps
		return nil, fmt.Errorf("failed to extract valid 'threat_scenario' or 'attack_vectors' from LLM response dictionary")
	}

	log.Printf("Workflow completed for: %s", analysisType)
	return result, nil
}