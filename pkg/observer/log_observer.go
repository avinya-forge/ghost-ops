package observer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"ghost-ops/pkg/protocol"
)

// LogObserver ingests, sanitizes, and buffers logs for AI analysis.
type LogObserver struct {
	llm       protocol.LLMProvider
	eventBus  protocol.EventBus
	collector protocol.MetricsCollector

	batchSize     int
	flushInterval time.Duration

	mu      sync.Mutex
	buffers map[string][]string // Keyed by serviceID
}

// NewLogObserver creates a new LogObserver.
func NewLogObserver(llm protocol.LLMProvider, eventBus protocol.EventBus, collector protocol.MetricsCollector, batchSize int, flushInterval time.Duration) *LogObserver {
	return &LogObserver{
		llm:           llm,
		eventBus:      eventBus,
		collector:     collector,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		buffers:       make(map[string][]string),
	}
}

// Start begins a background routine to flush buffers periodically.
func (o *LogObserver) Start(ctx context.Context) {
	ticker := time.NewTicker(o.flushInterval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				// Use a new context for the final flush so it doesn't immediately cancel
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				o.flushAll(shutdownCtx)
				return
			case <-ticker.C:
				go o.flushAll(ctx)
			}
		}
	}()
}

// Observe ingests a single log line, sanitizes it, and adds it to the buffer.
// It will flush asynchronously if the buffer reaches the batch size.
func (o *LogObserver) Observe(ctx context.Context, serviceID string, logLine string) {
	sanitized := o.sanitize(logLine)
	if sanitized == "" {
		return
	}

	o.mu.Lock()
	o.buffers[serviceID] = append(o.buffers[serviceID], sanitized)
	currentLen := len(o.buffers[serviceID])
	o.mu.Unlock()

	if o.collector != nil {
		o.collector.Counter("logs_observed_total", 1, map[string]string{"service_id": serviceID})
	}

	if currentLen >= o.batchSize {
		// Flush asynchronously, avoiding the context cancellation of the caller if it's transient
		go o.flushService(context.WithoutCancel(ctx), serviceID)
	}
}

// sanitize removes potentially sensitive information (PII) from the log string.
// This is an O(n) operation relative to the log line length.
func (o *LogObserver) sanitize(logLine string) string {
	if !strings.Contains(logLine, "@") && !strings.Contains(strings.ToLower(logLine), "secret") && !strings.Contains(strings.ToLower(logLine), "password") && !strings.Contains(strings.ToLower(logLine), "token") {
		return logLine
	}

	words := strings.Fields(logLine)
	for i, word := range words {
		lowerWord := strings.ToLower(word)
		if strings.Contains(word, "@") && strings.Contains(word, ".") {
			words[i] = "[REDACTED_EMAIL]"
		} else if strings.Contains(lowerWord, "secret") || strings.Contains(lowerWord, "password") || strings.Contains(lowerWord, "token") {
			if parts := strings.SplitN(word, "=", 2); len(parts) == 2 {
				words[i] = parts[0] + "=[REDACTED]"
			} else if strings.HasSuffix(lowerWord, "token:") || strings.HasSuffix(lowerWord, "password:") || strings.HasSuffix(lowerWord, "secret:") {
				words[i] = word + "[REDACTED]"
				if i+1 < len(words) {
					words[i+1] = ""
				}
			} else {
				words[i] = "[REDACTED_SECRET]"
			}
		}
	}

	var filtered []string
	for _, w := range words {
		if w != "" {
			filtered = append(filtered, w)
		}
	}

	return strings.Join(filtered, " ")
}

func (o *LogObserver) flushAll(ctx context.Context) {
	o.mu.Lock()
	services := make([]string, 0, len(o.buffers))
	for serviceID := range o.buffers {
		services = append(services, serviceID)
	}
	o.mu.Unlock()

	var wg sync.WaitGroup
	for _, serviceID := range services {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			o.flushService(ctx, id)
		}(serviceID)
	}
	wg.Wait()
}

func (o *LogObserver) flushService(ctx context.Context, serviceID string) {
	o.mu.Lock()
	batch := o.buffers[serviceID]
	if len(batch) == 0 {
		o.mu.Unlock()
		return
	}
	o.buffers[serviceID] = nil // Reset buffer
	o.mu.Unlock()

	o.AnalyzeLogs(ctx, serviceID, batch)
}

// AnalyzeLogs uses the LLM to find anomalies in the batch of logs.
func (o *LogObserver) AnalyzeLogs(ctx context.Context, serviceID string, batch []string) {
	if o.llm == nil {
		return
	}

	logsStr := strings.Join(batch, "\n")
	blueprint := protocol.Blueprint{
		ServiceID: serviceID,
		Intent:    AnomalyDetectionPrompt,
		Constraints: map[string]interface{}{
			"logs": logsStr,
		},
	}

	respStr, _, err := o.llm.GenerateCode(ctx, blueprint)
	if err != nil {
		fmt.Printf("Error analyzing logs for %s: %v\n", serviceID, err)
		return
	}

	var result struct {
		AnomalyDetected bool   `json:"anomaly_detected"`
		Reason          string `json:"reason"`
	}

	respStr = strings.TrimPrefix(respStr, "```json\n")
	respStr = strings.TrimSuffix(respStr, "\n```")
	respStr = strings.TrimSpace(respStr)

	if err := json.Unmarshal([]byte(respStr), &result); err != nil {
		fmt.Printf("Error parsing anomaly detection result: %v, body: %s\n", err, respStr)
		return
	}

	if result.AnomalyDetected && o.eventBus != nil {
		event := protocol.Event{
			Type:      protocol.EventAnomalyDetected,
			ServiceID: serviceID,
			Payload: map[string]interface{}{
				"reason": result.Reason,
				"logs":   batch,
			},
		}
		_ = o.eventBus.Publish(ctx, event)
	}
}
