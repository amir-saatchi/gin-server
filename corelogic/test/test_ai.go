package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time" // Import time package

	// Use your actual module path here

	"github.com/amir-saatchi/rest-api/corelogic/config"
	"github.com/amir-saatchi/rest-api/corelogic/similarity"
	"github.com/amir-saatchi/rest-api/corelogic/workflows"
	"github.com/joho/godotenv"
)

// Helper function to define common inputs
func getTestData(analysisType string) (similarity.InputData, map[string]string) {
	// Define common system info
	systemInfo := map[string]string{
		config.SystemType: "Infotainment System", // Example system type
		config.SystemDesc: "IVI system with connectivity features, including Bluetooth, Wi-Fi, and cellular modem. Integrates with vehicle CAN bus.", // Example description
	}

	// Base input data - customize per analysis if needed
	inputData := similarity.InputData{
		Asset:            "DRAM",
		Category:         "Component",
		Property:         "Confidentiality", // Relevant for Damage/Threat
		AssetDescription: "Volatile memory for fast data access, stores temporary application data.",
		// --- Fields needed for specific workflows ---
		DamageScenario: "Unauthorized access to sensitive user data stored temporarily in DRAM CAUSED BY Information Disclosure",                   // Needed for Impact/Threat
		Threat:         "Spoofing",                                  // Needed for Attack Steps
		ThreatScenario: "Spoofing of Bluetooth connection leads to unauthorized access to sensitive user data stored temporarily in DRAM", // Needed for Attack Steps/Feasibility/Attack Tree
		AttackSteps:    "1. Attacker pairs malicious device via Bluetooth. 2. Attacker exploits vulnerability in Bluetooth stack. 3. Attacker gains memory read access.", // Needed for Feasibility
		AttackVector:   "Remote",                                    // Needed for Attack Steps / Attack Tree
	}

	// --- You might want specific inputs per analysis type ---
	// Example: Overwrite Property for a different test case
	// if analysisType == config.ImpactScoresAnalysis {
	//     inputData.Property = "Availability"
	// }

	return inputData, systemInfo
}

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Could not load .env file. Ensure OPENAI_API_KEY is set via environment.")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: OPENAI_API_KEY environment variable not set.")
	}

	baseDataPath := "./data" // Path to reference JSON files
	// topK := 5                // Number of shots to retrieve

	// List of analysis types to test
	analysisTypes := []string{
		config.DamageScenarioAnalysis,
		config.ImpactScoresAnalysis,
		config.ThreatScenarioAnalysis,
		config.AttackStepsAnalysis,
		config.FeasibilityAnalysis,
		config.AttackTreeAnalysis,
	}

	fmt.Println("--- Starting Go Core Logic Performance Test ---")

	for _, analysisType := range analysisTypes {
		fmt.Printf("\n--- Testing Analysis Type: %s ---\n", analysisType)

		// Get consistent test data
		inputData, systemInfo := getTestData(analysisType)

		start := time.Now() // Start timer
		var result interface{}
		var workflowErr error

		log.Printf("\n==================== STARTING GO Test: %s ====================", analysisType)

		// Call the appropriate workflow function
		switch analysisType {
		case config.DamageScenarioAnalysis:
			result, workflowErr = workflows.GenerateDamageScenario(context.Background(), inputData, systemInfo, baseDataPath)
		case config.ImpactScoresAnalysis:
			result, workflowErr = workflows.GenerateImpactScores(context.Background(), inputData, systemInfo, baseDataPath)
		case config.ThreatScenarioAnalysis:
			result, workflowErr = workflows.GenerateThreatScenario(context.Background(), inputData, systemInfo, baseDataPath)
		case config.AttackStepsAnalysis:
			result, workflowErr = workflows.GenerateAttackSteps(context.Background(), inputData, systemInfo, baseDataPath)
		case config.FeasibilityAnalysis:
			result, workflowErr = workflows.GenerateFeasibility(context.Background(), inputData, systemInfo, baseDataPath)
		case config.AttackTreeAnalysis:
			result, workflowErr = workflows.GenerateAttackTree(context.Background(), inputData, systemInfo, baseDataPath)
		default:
			log.Printf("Skipping unrecognized analysis type: %s", analysisType)
			continue // Skip to next iteration
		}

		duration := time.Since(start) // Calculate duration

		if workflowErr != nil {
			log.Printf("ERROR during %s: %v", analysisType, workflowErr)
			fmt.Printf("Duration: %s (encountered error)\n", duration)
		} else {
			fmt.Printf("Duration: %s\n", duration)
			// Print a snippet of the result for verification (limit length)
			resultStr := fmt.Sprintf("%v", result) // Convert result to string
            maxLen := 150
			if len(resultStr) > maxLen {
				resultStr = resultStr[:maxLen] + "..."
			}
			fmt.Printf("Result Snippet: %s\n", resultStr)
		}
		fmt.Println("---------------------------------------------")

		log.Printf("==================== FINISHED GO Test: %s ====================\n", analysisType)

	} // End loop through analysis types

	fmt.Println("\n--- Go Core Logic Performance Test Finished ---")
}