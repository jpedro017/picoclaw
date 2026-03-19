package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sipeed/picoclaw/pkg/logger"
	"github.com/sipeed/picoclaw/pkg/providers/protocoltypes"
)

// UsageRecord represents a single LLM usage event logged to disk.
type UsageRecord struct {
	Timestamp                string `json:"timestamp"`
	AgentID                  string `json:"agent_id"`
	Model                    string `json:"model"`
	Channel                  string `json:"channel"`
	PromptTokens             int    `json:"prompt_tokens"`
	CompletionTokens         int    `json:"completion_tokens"`
	TotalTokens              int    `json:"total_tokens"`
	CacheReadInputTokens     int    `json:"cache_read_input_tokens,omitempty"`
	CacheCreationInputTokens int    `json:"cache_creation_input_tokens,omitempty"`
	ToolCalls                int    `json:"tool_calls"`
	Iteration                int    `json:"iteration"`
}

var (
	usageLogMu   sync.Mutex
	usageLogFile *os.File
	usageLogPath string
)

// logUsage appends a usage record to the JSONL file in the agent's workspace.
func logUsage(workspace string, agentID, model, channel string, usage *protocoltypes.UsageInfo, toolCalls, iteration int) {
	if usage == nil {
		return
	}

	record := UsageRecord{
		Timestamp:                time.Now().UTC().Format(time.RFC3339),
		AgentID:                  agentID,
		Model:                    model,
		Channel:                  channel,
		PromptTokens:             usage.PromptTokens,
		CompletionTokens:         usage.CompletionTokens,
		TotalTokens:              usage.TotalTokens,
		CacheReadInputTokens:     usage.CacheReadInputTokens,
		CacheCreationInputTokens: usage.CacheCreationInputTokens,
		ToolCalls:                toolCalls,
		Iteration:                iteration,
	}

	data, err := json.Marshal(record)
	if err != nil {
		logger.WarnCF("agent", "Failed to marshal usage record", map[string]any{"error": err.Error()})
		return
	}

	usageDir := filepath.Join(workspace, "usage")
	logPath := filepath.Join(usageDir, "usage.jsonl")

	usageLogMu.Lock()
	defer usageLogMu.Unlock()

	// Rotate or open file if needed
	if usageLogFile == nil || usageLogPath != logPath {
		if usageLogFile != nil {
			usageLogFile.Close()
		}
		if err := os.MkdirAll(usageDir, 0755); err != nil {
			logger.WarnCF("agent", "Failed to create usage dir", map[string]any{"error": err.Error()})
			return
		}
		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logger.WarnCF("agent", "Failed to open usage log", map[string]any{"error": err.Error()})
			return
		}
		usageLogFile = f
		usageLogPath = logPath
	}

	usageLogFile.Write(append(data, '\n'))
}
