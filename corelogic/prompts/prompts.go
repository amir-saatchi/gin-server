package prompts

import (
	// "bytes" // No longer needed for this function
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"path"

	"strings" // Import strings package
)

//go:embed templates/*.txt
var promptFS embed.FS

func init() {
	// ... (init function remains the same) ...
	log.Println("Initializing prompts package, checking embedded templates...")
	entries, err := promptFS.ReadDir("templates")
	if err != nil {
		log.Printf("ERROR: Could not read embedded 'templates' directory: %v", err)
		log.Println("       Ensure 'templates' directory exists inside 'corelogic/prompts/' and contains .txt files.")
		return
	}
	if len(entries) == 0 {
		log.Println("WARNING: Embedded 'templates' directory is empty or contains no matching *.txt files.")
		return
	}
	log.Println("Found embedded files in 'templates/':")
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			log.Printf("- %s", entry.Name())
			count++
		}
	}
    log.Printf("Total .txt files embedded: %d", count)
	log.Println("------------------------------------")
}


// LoadTemplate remains the same
func LoadTemplate(templateName string) (string, error) {
	templatePath := path.Join("templates", templateName)
	content, err := promptFS.ReadFile(templatePath)
	if err != nil {
		log.Printf("ERROR in LoadTemplate: failed to read embedded template %s: %v", templatePath, err)
		return "", fmt.Errorf("failed to read embedded template (tried path: %s): %w", templatePath, err)
	}
	return string(content), nil
}


// --- MODIFIED: FormatInstructionsPrompt using string replacement ---
// FormatInstructionsPrompt formats a template string using simple placeholder replacement.
// Placeholders should be in the format {key_name}.
func FormatInstructionsPrompt(templateContent string, contextData map[string]any) (string, error) {
	formattedPrompt := templateContent // Start with the original template

	for key, value := range contextData {
		placeholder := fmt.Sprintf("{%s}", key) // Create the {key} placeholder
		// Convert value to string for replacement
		// Handle potential nil values gracefully
		var valueStr string
		if value == nil {
			valueStr = "" // Replace nil with empty string, or choose other default
		} else {
			valueStr = fmt.Sprintf("%v", value) // Use generic %v formatting
		}
		formattedPrompt = strings.ReplaceAll(formattedPrompt, placeholder, valueStr)
	}

	// Optional: Check if any {placeholders} remain unreplaced
	if strings.Contains(formattedPrompt, "{") && strings.Contains(formattedPrompt, "}") {
		// Basic check, might yield false positives but good for debugging
		log.Printf("Warning: Potential unreplaced placeholders remain in formatted prompt starting with: %s", formattedPrompt[:200])
	}

	return formattedPrompt, nil
}
// --- END MODIFICATION ---


// FormatUserMessage remains the same
func FormatUserMessage(inputData map[string]any) (string, error) {
	// ... (implementation remains the same) ...
	jsonData, err := json.Marshal(inputData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal input data to JSON: %w", err)
	}
	return fmt.Sprintf("####%s####", string(jsonData)), nil
}

// getMapKeys helper function remains the same
func getMapKeys(m map[string]any) []string {
	// ... (implementation remains the same) ...
     keys := make([]string, 0, len(m))
     for k := range m {
         keys = append(keys, k)
     }
     return keys
}