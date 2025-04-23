package workflows

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/amir-saatchi/rest-api/corelogic/config"
	"github.com/amir-saatchi/rest-api/corelogic/llm"
	"github.com/amir-saatchi/rest-api/corelogic/prompts"
	"github.com/amir-saatchi/rest-api/corelogic/similarity"
)

// executeWorkflow handles the common steps of context prep, LLM calls (Base/Validate),
// and returns the RAW final response from the LLM for specific parsing by the caller.
func executeWorkflow(
	ctx context.Context,
	cfg *config.ModelConfig, // Configuration for the current analysis
	inputData similarity.InputData, // User's input data
	systemInfo map[string]string, // System context info
	shotsResult *similarity.SimilarityResult, // Results from similarity search
	apiKey string, // OpenAI API Key
	llmModel string, // Model ID (e.g., gpt-4o)
) (string, error) { // Returns the raw final LLM response string

	// --- 1. Prepare Shots Context ---
	var shotsForPrompt []map[string]any
	if shotsResult != nil {
		for _, shot := range shotsResult.Shots {
			shotBytes, _ := json.Marshal(shot)
			var shotMap map[string]any
			_ = json.Unmarshal(shotBytes, &shotMap)
			filteredShotMap := make(map[string]any)
			for _, key := range cfg.ShotsKeys {
				if val, ok := shotMap[key]; ok && val != nil {
					filteredShotMap[key] = val
				}
			}
			if len(filteredShotMap) > 0 {
				shotsForPrompt = append(shotsForPrompt, filteredShotMap)
			}
		}
	}
	formattedShots := config.FormatShotsForPrompt(shotsForPrompt)

	// --- 2. Prepare Base System Prompt Context ---
	// --- Prepare Base System Prompt Context ---
	baseSystemPromptContext := map[string]any{
    "system_type":       systemInfo[config.SystemType], // Use "system_type" key to match {system_type}
    "system_desc":       systemInfo[config.SystemDesc], // Use "system_desc" key to match {system_desc}
    "shots":            formattedShots,                // Use "shots" key to match {shots}
    "asset":            inputData.Asset,               // Use "asset" key? Check template
    "category":         inputData.Category,            // Use "category" key? Check template
    "property":         inputData.Property,            // Use "property" key? Check template
    "asset_description": inputData.AssetDescription,    // Use "asset_description" key? Check template
	"threat":           inputData.Threat,              // Use "threat" key? Check template
	"threat_scenario":   inputData.ThreatScenario,    // Use "threat_scenario" key? Check template
	"attack_steps":      inputData.AttackSteps,         // Use "attack_steps" key? Check template"
	"attack_vector":     inputData.AttackVector,        // Use "attack_vector" key? Check template
	"damage_scenario":   inputData.DamageScenario,      // Use "damage_scenario" key? Check template

	}



	// --- 3. Format Base User Message ---
	userInputMap := make(map[string]any)
	// Include only the keys relevant as direct input trigger, mirroring Python's approach more closely
	// This might vary slightly per prompt, but often includes asset/desc/property etc.
	// Let's use the keys specified in cfg.DataKeys as a basis for the user input context map.
	inputDataMapBytes, _ := json.Marshal(inputData)
	var inputDataMap map[string]any
	_ = json.Unmarshal(inputDataMapBytes, &inputDataMap)

	for _, key := range cfg.DataKeys { // Use DataKeys from config
		if val, ok := inputDataMap[key]; ok && val != "" {
			userInputMap[key] = val
		}
	}
	// Special handling for attack_vector definition if needed (as in Python api.py _create_user_msg)
	// This might need to be handled *before* this helper or passed in differently if critical
	// For now, we assume the base userInputMap is sufficient for FormatUserMessage

	userMessageContent, err := prompts.FormatUserMessage(userInputMap)
	if err != nil {
		return "", fmt.Errorf("failed to format base user message: %w", err)
	}

	// --- 4. Execute LLM Step (Base or Validate) ---
	var finalRawResponse string

	if cfg.LLMStep == config.StepBase {
		// --- BASE ---
		if len(cfg.PromptFiles) < 1 {
			return "", fmt.Errorf("base step requires at least 1 prompt file in config")
		}
		templateName := cfg.PromptFiles[0]
		log.Printf("Executing BASE step using template: %s", templateName)
		templateContent, err := prompts.LoadTemplate(templateName)
		if err != nil {
			return "", fmt.Errorf("base workflow error loading template %s: %w", templateName, err)
		}
		formattedSystemPrompt, err := prompts.FormatInstructionsPrompt(templateContent, baseSystemPromptContext)
		if err != nil {
			return "", fmt.Errorf("base workflow error formatting system prompt: %w", err)
		}

		llmResponse, err := llm.CallChatCompletion(ctx, formattedSystemPrompt, userMessageContent, apiKey, llmModel, 0.1)
		if err != nil {
			return "", fmt.Errorf("base workflow error during LLM call: %w", err)
		}
		finalRawResponse = llmResponse

	} else if cfg.LLMStep == config.StepValidate {
		// --- VALIDATE ---
		if len(cfg.PromptFiles) < 2 {
			return "", fmt.Errorf("validation step requires 2 prompt files in config, found %d", len(cfg.PromptFiles))
		}
		baseTemplateName := cfg.PromptFiles[0]
		validateTemplateName := cfg.PromptFiles[1]
		log.Printf("Executing VALIDATE step using base template: %s", baseTemplateName)

		// Load and Format BASE Prompt's System Instructions
		baseTemplateContent, err := prompts.LoadTemplate(baseTemplateName)
		if err != nil {
			return "", fmt.Errorf("validate workflow error loading base template %s: %w", baseTemplateName, err)
		}
		baseFormattedSystemPrompt, err := prompts.FormatInstructionsPrompt(baseTemplateContent, baseSystemPromptContext)
		if err != nil {
			return "", fmt.Errorf("validate workflow error formatting base system prompt: %w", err)
		}

		// --- Trials ---
		numTrials := 3
		expertResponses := []string{}
		for i := 0; i < numTrials; i++ {
			log.Printf("Validation trial %d/%d", i + 1, numTrials)
			// Use base system prompt and formatted user input for trials
			resp, err := llm.CallChatCompletion(ctx, baseFormattedSystemPrompt, userMessageContent, apiKey, llmModel, 0.5)
			if err != nil {
				log.Printf("Warning: Validation trial %d failed: %v", i+1, err)
				continue
			}

			// Parse the specific result from the trial response based on the base prompt's structure.
			// ASSUMPTION: All base prompts for VALIDATE workflows output their main result after the last '####'.
			// This might need adjustment if base prompts output dictionaries (like attack_steps_base).
			// For now, using parseFinalStepResponse as done for damage scenario.
			trialResult, err := parseFinalStepResponse(resp, "####")
			if err != nil {
				log.Printf("Warning: Failed to parse result from trial %d using '####': %v", i+1, err)
				continue
			}
			if trialResult != "" {
				expertResponses = append(expertResponses, trialResult)
				log.Printf("Trial %d result added.", i+1)
			} else {
				log.Printf("Warning: Parsed empty result from trial %d.", i+1)
			}
		}

		// --- Consolidation ---
		aggregatedExpertsRes := "No valid responses generated in trials."
		if len(expertResponses) > 0 {
			// Add numbering like Python
			var sb strings.Builder
			for i, resp := range expertResponses {
				sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, resp))
			}
			aggregatedExpertsRes = strings.TrimSpace(sb.String())
		} else {
			log.Println("Warning: All validation trials failed or produced no parsable result.")
		}

		// Prepare context specifically for the validation template
		// Include fields referenced by the specific validation template being used
		validationPromptContext := map[string]any{
			"system_type":       	systemInfo[config.SystemType], 	// Use "system_type" key
			"asset":            	inputData.Asset,               	// Use "asset" key
			"category":         	inputData.Category,            	// Use "category" key
			"property":         	inputData.Property,            	// Use "property" key
			"asset_description": 	inputData.AssetDescription,    	// Use "asset_description" key
			"threat":           	inputData.Threat,              	// Use "threat" key? Check template
			"threat_scenario":   	inputData.ThreatScenario,      	// Use "threat_scenario" key? Check template
			"attack_vector":     	inputData.AttackVector,        	// Use "attack_vector" key? Check template
			"experts_res":       	aggregatedExpertsRes,        	// Use "experts_res" key
		}

		// Load and format validation prompt (treating it as user message for the final call)
		log.Printf("Executing validation consolidation using template: %s", validateTemplateName)
		validateTemplateContent, err := prompts.LoadTemplate(validateTemplateName)
		if err != nil {
			return "", fmt.Errorf("validate workflow error loading validate template %s: %w", validateTemplateName, err)
		}
		validateFormattedUserPrompt, err := prompts.FormatInstructionsPrompt(validateTemplateContent, validationPromptContext)
		if err != nil {
			return "", fmt.Errorf("validate workflow error formatting validate prompt: %w", err)
		}

		// Define the system prompt for the final consolidation call
		finalSystemPrompt := "You are an automotive cybersecurity expert tasked with consolidating potential damage scenarios or attack steps based on expert input and providing the single most correct one formatted as instructed." // Make this more generic?

		log.Printf("Final User Prompt: %s", validateFormattedUserPrompt)
		// Final LLM call for consolidation
		llmResponse, err := llm.CallChatCompletion(ctx, finalSystemPrompt, validateFormattedUserPrompt, apiKey, llmModel, 0.1)
		if err != nil {
			return "", fmt.Errorf("validate workflow error during consolidation LLM call: %w", err)
		}
		finalRawResponse = llmResponse

	} else {
		return "", fmt.Errorf("unknown LLM step type in config: %s", cfg.LLMStep)
	}

	// Return the raw response string - specific parsing happens in the calling workflow func
	return finalRawResponse, nil
}

// Regex to find potential JSON objects or arrays in the text
var jsonObjectRegex = regexp.MustCompile(`(?s)\{.*?\}`) // Non-greedy match for {}
// var jsonArrayRegex = regexp.MustCompile(`(?s)\[.*?\]`)  // Non-greedy match for []

// cleanJSONString attempts to clean up common issues in LLM-generated JSON-like strings
func cleanJSONString(raw string) string {
	// Remove markdown code fences
	cleaned := strings.TrimPrefix(raw, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	// Replace Python-style None/True/False if necessary (use with caution)
	// cleaned = strings.ReplaceAll(cleaned, "None", "null")
	// cleaned = strings.ReplaceAll(cleaned, "True", "true")
	// cleaned = strings.ReplaceAll(cleaned, "False", "false")
	// Replace single quotes with double quotes (common LLM mistake)
	cleaned = strings.ReplaceAll(cleaned, "'", "\"")
	// Remove trailing commas before closing braces/brackets (more complex regex needed for reliability)
	// Basic attempt:
	re := regexp.MustCompile(`,\s*([}\]])`)
	cleaned = re.ReplaceAllString(cleaned, "$1")

	return strings.TrimSpace(cleaned)
}

// parseFinalStepResponse extracts the content after the last specified delimiter.
func parseFinalStepResponse(responseText string, delimiter string) (string, error) {
	parts := strings.Split(responseText, delimiter)
	if len(parts) < 2 { // Need at least one delimiter occurrence to have a part after it
		log.Printf("Warning: Delimiter '%s' not found or response format unexpected. Returning raw response. Raw: %s", delimiter, responseText)
		// Depending on strictness, you might return an error here instead
		return responseText, nil // Fallback to raw response
		// return "", fmt.Errorf("delimiter '%s' not found in response", delimiter)
	}
	// Return the last part, trimmed of whitespace
	return strings.TrimSpace(parts[len(parts)-1]), nil
}

// parseDictResponse tries to parse a dictionary from the last step of an LLM response.
// func parseDictResponse(responseText string, delimiter string) (map[string]any, error) {
// 	finalStep, err := parseFinalStepResponse(responseText, delimiter)
// 	if err != nil {
// 		// If parseFinalStepResponse returns an error (e.g., delimiter not found)
// 		log.Printf("Warning: Could not extract final step using delimiter '%s'. Trying to parse raw response.", delimiter)
// 		finalStep = responseText // Attempt to parse the whole response as JSON
// 		// return nil, err
// 	}

// 	cleanedJSON := cleanJSONString(finalStep)

// 	var resultDict map[string]any
// 	err = json.Unmarshal([]byte(cleanedJSON), &resultDict)
// 	if err != nil {
// 		log.Printf("Error unmarshalling JSON response. Cleaned text: '%s', Original final step: '%s', Raw Response: '%s'", cleanedJSON, finalStep, responseText)
// 		return nil, fmt.Errorf("failed to parse dictionary from LLM response: %w", err)
// 	}
// 	return resultDict, nil
// }

// parseListResponse tries to parse a list from the last step of an LLM response.
// func parseListResponse(responseText string, delimiter string) ([]any, error) {
// 	finalStep, err := parseFinalStepResponse(responseText, delimiter)
// 	if err != nil {
// 		log.Printf("Warning: Could not extract final step using delimiter '%s'. Trying to parse raw response.", delimiter)
// 		finalStep = responseText
// 	}

// 	cleanedJSON := cleanJSONString(finalStep)

// 	var resultList []any
// 	err = json.Unmarshal([]byte(cleanedJSON), &resultList)
// 	if err != nil {
// 		log.Printf("Error unmarshalling JSON list response. Cleaned text: '%s', Original final step: '%s', Raw Response: '%s'", cleanedJSON, finalStep, responseText)
// 		return nil, fmt.Errorf("failed to parse list from LLM response: %w", err)
// 	}
// 	return resultList, nil
// }

// --- NEW Helper: parseDelimitedJSON ---
// Attempts to find and parse JSON content explicitly wrapped by start and end delimiters.
func parseDelimitedJSON(responseText string, startDelimiter string, endDelimiter string) (string, error) {
    startIndex := strings.Index(responseText, startDelimiter)
    if startIndex == -1 {
        return "", fmt.Errorf("start delimiter '%s' not found", startDelimiter)
    }
    // Adjust startIndex to be *after* the delimiter
    startIndex += len(startDelimiter)

    endIndex := strings.Index(responseText[startIndex:], endDelimiter)
    if endIndex == -1 {
         // Maybe the end delimiter is missing? Try parsing from startDelimiter to end of string?
         // Or return error. Let's return error for now.
         return "", fmt.Errorf("end delimiter '%s' not found after start delimiter", endDelimiter)
    }
    // endIndex is relative to the substring after startIndex, adjust it
    endIndex += startIndex

    // Extract the content between delimiters
    jsonContent := responseText[startIndex:endIndex]
    return strings.TrimSpace(jsonContent), nil
}


// parseDictResponse - Modify to handle !!!! delimiter case more strictly
func parseDictResponse(responseText string, delimiter string) (map[string]any, error) {
	var jsonStr string
	var err error

	// Special handling for !!!! delimiter based on feasibility prompt
	if delimiter == "!!!!" {
		jsonStr, err = parseDelimitedJSON(responseText, "!!!!", "!!!!")
		if err != nil {
			// If we expect !!!!{...}!!!!, and can't find it, it's a parsing failure.
			// Don't fall back to other methods for this specific case.
			log.Printf("Failed to extract content between '!!!!' delimiters: %v. Raw Response: '%s'", err, responseText)
			return nil, fmt.Errorf("could not find content between required '!!!!' delimiters: %w", err)
		}
		// Proceed directly to cleaning and unmarshalling jsonStr if found between delimiters
	} else if delimiter == "####" { // Handle standard #### cases
		// Try last segment first
		jsonStr, err = parseFinalStepResponse(responseText, delimiter)
		if err != nil {
			log.Printf("Delimiter '%s' not found. Will search entire response for JSON object.", delimiter)
			// Fallback: Search the entire response for a JSON object
			jsonStr = jsonObjectRegex.FindString(responseText)
			if jsonStr == "" {
				log.Printf("Warning: No JSON object structure found in response via regex. Raw Response: '%s'", responseText)
				return nil, fmt.Errorf("no JSON object found in LLM response using any method")
			}
		}
	} else {
		// Default or unknown delimiter: just use regex search
		log.Printf("Unknown delimiter '%s', attempting regex search for JSON object.", delimiter)
		jsonStr = jsonObjectRegex.FindString(responseText)
		if jsonStr == "" {
			log.Printf("Warning: No JSON object structure found in response via regex. Raw Response: '%s'", responseText)
			return nil, fmt.Errorf("no JSON object found in LLM response using regex")
		}
	}

	// Clean and Unmarshal the extracted/found jsonStr
	cleanedJSON := cleanJSONString(jsonStr)
	if cleanedJSON == "" {
		return nil, fmt.Errorf("extracted JSON content was empty after cleaning (original extracted: '%s')", jsonStr)
	}

	var resultDict map[string]any
	err = json.Unmarshal([]byte(cleanedJSON), &resultDict)
	if err != nil {
		log.Printf("Error unmarshalling JSON response. Cleaned text: '%s', Original extracted: '%s', Raw Response: '%s'", cleanedJSON, jsonStr, responseText)
		return nil, fmt.Errorf("failed to parse dictionary from LLM response: %w", err)
	}
	return resultDict, nil
}
