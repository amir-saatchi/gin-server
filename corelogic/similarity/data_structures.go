package similarity

// InputData matches the structure expected by your service
type InputData struct {
	Asset            string `json:"Asset"`
	Category         string `json:"Category"`
	Property         string `json:"Property"`
	AssetDescription string `json:"Asset Description"`
	AttackVector     string `json:"attack_vector,omitempty"` // ADDED for attack steps/tree
	// Add other fields that might be part of the input (e.g., ThreatScenario for feasibility)
	ThreatScenario   string `json:"threat_scenario,omitempty"`
	AttackSteps      string `json:"attack_steps,omitempty"` // Needed for feasibility input
	DamageScenario   string `json:"damage_scenario,omitempty"` // ADDED for threat/impact context
    Threat           string `json:"threat,omitempty"` // ADDED for attack steps context
}

// ReferenceData represents the structure stored in your reference file
// Now explicitly includes the embedding.
type ReferenceData struct {
	ID               	string    `json:"id"`
	Embedding        	[]float32 `json:"embedding"`
	Asset            	string    `json:"Asset"`
	Category         	string    `json:"Category"`
	Property         	string    `json:"Property"`
	AssetDescription 	string    `json:"AssetDescription"` // Ensure JSON tag matches output
	DamageScenario   	string    `json:"damage_scenario"`   // Ensure JSON tag matches output
	ThreatScenario   	string    `json:"threat_scenario"`   // Ensure JSON tag matches output
	AttackSteps      	string    `json:"attack_steps"`      // Ensure JSON tag matches output
	// Add other metadata fields from your JSON reference files if needed
	// Example: Impact scores if they are in the reference file
	SafetyImpact     	any `json:"safety_impact,omitempty"` // Use any if type varies (e.g., string/int)
	FinancialImpact  	any `json:"financial_impact,omitempty"`
	OperationalImpact 	any `json:"operational_impact,omitempty"`
	PrivacyImpact    	any `json:"privacy_impact,omitempty"`
	ET               	any `json:"et,omitempty"`
	SE               	any `json:"se,omitempty"`
	KOIC             	any `json:"koic,omitempty"`
	WOO              	any `json:"woo,omitempty"`
	EQ               	any `json:"eq,omitempty"`
}

// ResultWithScore combines ReferenceData with its calculated similarity score
type ResultWithScore struct {
	Data  ReferenceData
	Score float64
}

// SimilarityResult holds the top K results
type SimilarityResult struct {
	Shots []ReferenceData `json:"shots"`
}

// Define result structures for different workflows where applicable
// Example for dictionary results
type DictResult map[string]any

// Example for Threat Scenario result
type ThreatScenarioResult struct {
	ThreatScenario string   `json:"threat_scenario"`
	AttackVectors  []string `json:"attack_vectors"`
}