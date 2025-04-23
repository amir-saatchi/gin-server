package llm

import (
	"context"
	"fmt"
	"log"
	"time" // Import time package for retry delay

	openai "github.com/sashabaranov/go-openai" // Or your official client import
)

const maxRetries = 3 // Number of retries for LLM calls

// CallChatCompletion sends distinct system and user prompts to OpenAI chat completion and returns the response.
// It handles basic retries on failure.
func CallChatCompletion(ctx context.Context, systemPrompt, userPrompt string, apiKey string, model string, temperature float32) (string, error) {
	client := openai.NewClient(apiKey)
	var lastErr error

	// Construct the messages slice based on provided prompts
	messages := []openai.ChatCompletionMessage{}
	if systemPrompt != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		})
	}
	if userPrompt == "" {
		// Should generally not happen if called correctly, but good to check
		return "", fmt.Errorf("userPrompt cannot be empty")
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userPrompt,
	})


		// --- ADDED: Log the messages being sent ---
		log.Printf("[Go LLM Call] Sending %d messages to LLM (model: %s):", len(messages), model)
		for i, msg := range messages {
			// Log role and start of content (limit length for readability)
			contentSnippet := msg.Content
			maxLogLen := 500 // Adjust as needed
			if len(contentSnippet) > maxLogLen {
				contentSnippet = contentSnippet[:maxLogLen] + "..."
			}
			log.Printf("[Go LLM Call]   Msg[%d] Role: %s, Content Snippet: %s", i, msg.Role, contentSnippet)
			// Optionally log the full content if needed for deep debugging, but it can be very long
			// fullMsgJson, _ := json.MarshalIndent(msg, "  ", "  ")
			// log.Printf("[Go LLM Call]   Msg[%d] Full: %s", i, string(fullMsgJson))
		}
		// --- END ADDED LOGGING ---

	for attempt := 0; attempt < maxRetries; attempt++ {
		req := openai.ChatCompletionRequest{
			Model:       model, // e.g., openai.GPT4o, openai.GPT4oMini, or specific string IDs
			Messages:    messages,
			Temperature: temperature,
			// Add other parameters like MaxTokens if needed
		}

		resp, err := client.CreateChatCompletion(ctx, req)
		if err != nil {
			// Append attempt number and model to the error context
			err = fmt.Errorf("attempt %d using model %s failed: %w", attempt+1, model, err)
			lastErr = err // Store the last error encountered
			log.Printf("Warning: %v. Retrying in %d seconds...", err, (attempt+1)*2)
			// Implement more sophisticated backoff if needed
			time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
			continue // Retry
		}

		if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
			lastErr = fmt.Errorf("openai attempt %d using model %s returned empty response choice", attempt+1, model)
			log.Printf("Warning: %v. Retrying in %d seconds...", lastErr, (attempt+1)*2)
			time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
			continue // Retry
		}

		// Success
		log.Printf("LLM call successful on attempt %d.", attempt+1)
		// Optionally log token usage: log.Printf("Usage: %+v", resp.Usage)
		return resp.Choices[0].Message.Content, nil
	}

	// All retries failed
	log.Printf("Error: LLM call failed after %d attempts.", maxRetries)
	// Return the *last* error encountered during retries
	return "", fmt.Errorf("llm call failed after %d attempts: %w", maxRetries, lastErr)
}