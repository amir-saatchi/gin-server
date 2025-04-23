package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Constants for Analysis Types
const (
	DamageScenarioAnalysis = "model_damage_scenario"
	ImpactScoresAnalysis   = "model_impact_scores"
	ThreatScenarioAnalysis = "model_threat_scenario"
	AttackStepsAnalysis    = "model_attack_steps"
	FeasibilityAnalysis    = "model_feasibility_assessment"
	AttackTreeAnalysis     = "model_attack_tree"
	// Add other constants as needed
)

// Constants for LLM Steps
const (
	StepBase     = "base"
	StepValidate = "validate"
)

// Constants for common data keys (mirroring Python constants.py)
// It's often better to define these where they are used or pass them explicitly,
// but for mirroring Python, we can put some common ones here.
const (
	Asset            = "Asset"
	Category         = "Category"
	Property         = "Property"
	AssetDescription = "Asset Description"
	DamageScenario   = "damage_scenario"
	ThreatScenario   = "threat_scenario"
	AttackSteps      = "attack_steps"
	AttackVector     = "attack_vector"
	SystemType       = "system_type"
	SystemDesc       = "system_desc"
	Shots            = "shots"
	ExpertsRes       = "experts_res"
	Threat           = "threat"
	
	// IDs
	AssetPropertyID = "ap_id"
	DamageID        = "damage_id"
	ThreatID        = "threat_id"
	AttackID        = "attack_id"

	// --- ADDED Constants for Impact Columns ---
	PrivacyImpact        = "privacy_impact"
	SafetyImpact         = "safety_impact"
	FinancialImpact      = "financial_impact"
	OperationalImpact    = "operational_impact" // Added
	OEMFinancialImpact   = "oem_financial_impact"
	OEMOperationalImpact = "oem_operational_impact"
	OEMIPImpact          = "oem_ip_impact"

	// --- ADDED Constants for Risk/Feasibility Columns ---
	ET   = "et"
	SE   = "se"
	KOIC = "koic" // Added
	WOO  = "woo"
	EQ   = "eq"
)

type ModelConfig struct {
	AnalysisType          string
	DataID                string   // Primary ID key for data lookup/grouping
	DataKeys              []string // ADDED: Keys defining the primary input data fields
	EmbedKeys             []string // Keys used for generating embeddings
	ShotsKeys             []string // Keys expected in the "shots" data
	ReferenceDataJSONFile string   // Filename of the reference JSON in the data dir
	PromptFiles           []string // List of prompt template filenames
	LLMStep               string   // e.g., StepBase or StepValidate
}

// --- configMap - ADDED DataKeys slices based on Python model_attrs.py ---
var configMap = map[string]ModelConfig{
	DamageScenarioAnalysis: {
		AnalysisType:          DamageScenarioAnalysis,
		DataID:                AssetPropertyID,
		DataKeys:              []string{Asset, Category, Property, AssetDescription}, // From Python
		EmbedKeys:             []string{Asset, Category, Property, AssetDescription},
		ShotsKeys:             []string{DamageID, Asset, Category, Property, AssetDescription, DamageScenario},
		ReferenceDataJSONFile: "damage_scenario_reference.json",
		PromptFiles:           []string{"damage_scenario_base.txt", "damage_scenario_validate.txt"},
		LLMStep:               StepValidate,
	},
	ImpactScoresAnalysis: {
		AnalysisType:          ImpactScoresAnalysis,
		DataID:                DamageID,
		DataKeys:              []string{Asset, Category, Property, DamageScenario, AssetDescription}, // From Python
		EmbedKeys:             []string{Asset, Category, Property, DamageScenario},
		ShotsKeys:             []string{DamageID, Asset, Property, AssetDescription, DamageScenario, SafetyImpact, FinancialImpact, OperationalImpact, PrivacyImpact, OEMFinancialImpact, OEMOperationalImpact, OEMIPImpact},
		ReferenceDataJSONFile: "impact_scores_reference.json",
		PromptFiles:           []string{"impact_scores_base.txt"},
		LLMStep:               StepBase,
	},
	ThreatScenarioAnalysis: {
		AnalysisType:          ThreatScenarioAnalysis,
		DataID:                DamageID,
		DataKeys:              []string{Asset, Category, Property, Threat, DamageScenario, AssetDescription}, // From Python
		EmbedKeys:             []string{Asset, Category, Property, DamageScenario},
		ShotsKeys:             []string{DamageID, Asset, Property, Threat, DamageScenario, ThreatScenario},
		ReferenceDataJSONFile: "threat_scenario_reference.json",
		PromptFiles:           []string{"threat_scenario_base.txt"},
		LLMStep:               StepBase,
	},
	AttackStepsAnalysis: {
		AnalysisType:          AttackStepsAnalysis,
		DataID:                AttackID,
		DataKeys:              []string{Asset, Category, Threat, ThreatScenario, AssetDescription, AttackVector}, // From Python
		EmbedKeys:             []string{Asset, Category, Threat, ThreatScenario},
		ShotsKeys:             []string{AttackID, Asset, Category, Threat, ThreatScenario, AttackSteps},
		ReferenceDataJSONFile: "attack_steps_reference.json",
		PromptFiles:           []string{"attack_steps_base.txt", "attack_steps_validate.txt"},
		LLMStep:               StepValidate,
	},
	FeasibilityAnalysis: {
		AnalysisType:          FeasibilityAnalysis,
		DataID:                AttackID,
		DataKeys:              []string{Asset, Category, Threat, ThreatScenario, AttackSteps, AssetDescription}, // From Python
		EmbedKeys:             []string{Asset, Category, Threat, AttackSteps},
		ShotsKeys:             []string{AttackID, Asset, Category, Threat, ThreatScenario, AttackSteps, ET, SE, KOIC, WOO, EQ},
		ReferenceDataJSONFile: "feasibility_reference.json",
		PromptFiles:           []string{"feasibility_base.txt"},
		LLMStep:               StepBase,
	},
	AttackTreeAnalysis: {
		AnalysisType:          AttackTreeAnalysis,
		DataID:                AttackID,
		DataKeys:              []string{Asset, Category, Property, ThreatScenario, AssetDescription, AttackVector}, // From Python
		EmbedKeys:             []string{Asset, Category, Threat, ThreatScenario},                                    // Matches Python
		ShotsKeys:             []string{AttackID, Asset, Category, Threat, ThreatScenario, AttackSteps},             // Matches Python
		ReferenceDataJSONFile: "attack_steps_reference.json",                                                       // Matches Python
		PromptFiles:           []string{"attack_tree.txt"},
		LLMStep:               StepBase,
	},
}

// GetConfig retrieves the configuration for a given analysis type.
func GetConfig(analysisType string, baseDataPath string) (*ModelConfig, error) {
	config, exists := configMap[analysisType]
	if !exists {
		return nil, fmt.Errorf("unknown analysis type: %s", analysisType)
	}

	// Construct full path for the reference file
	config.ReferenceDataJSONFile = filepath.Join(baseDataPath, config.ReferenceDataJSONFile)

	// Check if file exists
	if _, err := os.Stat(config.ReferenceDataJSONFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("reference data file not found for type %s at path %s", analysisType, config.ReferenceDataJSONFile)
	} else if err != nil {
		return nil, fmt.Errorf("error checking reference file %s: %w", config.ReferenceDataJSONFile, err)
	}

	return &config, nil
}

// FormatShotsForPrompt helper function remains the same
func FormatShotsForPrompt(shots []map[string]interface{}) string {
	// ... (implementation remains the same) ...
	var builder strings.Builder
	for _, shot := range shots {
		builder.WriteString("```\n")
		for key, value := range shot {
			builder.WriteString(fmt.Sprintf("%s: %v\n", key, value))
		}
		builder.WriteString("```\n")
	}
	return builder.String()
}