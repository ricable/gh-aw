package workflow

import (
	"encoding/json"
	"strings"

	"github.com/github/gh-aw/pkg/logger"
)

var mistralVibeLogsLog = logger.New("workflow:mistral_vibe_logs")

// ParseLogMetrics extracts metrics from Mistral Vibe log content
// Vibe outputs streaming JSON with session information
func (e *MistralVibeEngine) ParseLogMetrics(logContent string, verbose bool) LogMetrics {
	mistralVibeLogsLog.Printf("Parsing Mistral Vibe log metrics: log_size=%d bytes, verbose=%v", len(logContent), verbose)

	var metrics LogMetrics

	// Parse JSON streaming output from Vibe
	lines := strings.Split(logContent, "\n")

	var totalInputTokens, totalOutputTokens int
	var turnCount int
	var toolCallsMap = make(map[string]*ToolCallInfo)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "{") {
			continue
		}

		// Try to parse as JSON
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		// Look for session summary information
		if sessionData, hasSession := entry["session"].(map[string]any); hasSession {
			if usage, hasUsage := sessionData["usage"].(map[string]any); hasUsage {
				if inputTokens, ok := usage["input_tokens"].(float64); ok {
					totalInputTokens += int(inputTokens)
				}
				if outputTokens, ok := usage["output_tokens"].(float64); ok {
					totalOutputTokens += int(outputTokens)
				}
			}
			if turns, hasTurns := sessionData["turns"].(float64); hasTurns {
				turnCount = int(turns)
			}
		}

		// Count tool calls
		if toolCall, hasToolCall := entry["tool_call"].(map[string]any); hasToolCall {
			if toolName, hasName := toolCall["name"].(string); hasName && toolName != "" {
				if info, exists := toolCallsMap[toolName]; exists {
					info.CallCount++
				} else {
					toolCallsMap[toolName] = &ToolCallInfo{
						Name:      toolName,
						CallCount: 1,
					}
				}
			}
		}
	}

	// Convert map to slice for ToolCalls
	var toolCallsList []ToolCallInfo
	for _, info := range toolCallsMap {
		toolCallsList = append(toolCallsList, *info)
	}

	metrics.Turns = turnCount
	metrics.TokenUsage = totalInputTokens + totalOutputTokens
	metrics.ToolCalls = toolCallsList

	mistralVibeLogsLog.Printf("Parsed metrics: turns=%d, total_tokens=%d, tool_calls=%d",
		metrics.Turns, metrics.TokenUsage, len(metrics.ToolCalls))

	return metrics
}

// GetLogParserScriptId returns the name of the JavaScript script to parse Vibe logs
// For now, we use Go-based parsing, so no JavaScript parser is needed
func (e *MistralVibeEngine) GetLogParserScriptId() string {
	return ""
}

// GetLogFileForParsing returns the log file path to use for parsing
func (e *MistralVibeEngine) GetLogFileForParsing() string {
	return "/tmp/gh-aw/agent-stdio.log"
}
