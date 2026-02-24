package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/vincentmaurin/fal-cli/internal/output"
)

var runInputFlag string
var runQueueFlag bool
var runLogsFlag bool

var runCmd = &cobra.Command{
	Use:   "run <model-id>",
	Short: "Run a model (synchronous by default, use --queue for async)",
	Long: `Run a fal.ai model with a JSON input payload.

By default runs synchronously (connection stays open until result).
Use --queue to submit to the queue and poll until completion.

Examples:
  fal run fal-ai/nano-banana-pro --input '{"prompt":"a cat"}'
  fal run fal-ai/flux/dev --input '{"prompt":"a cat"}' --queue
  fal run fal-ai/nano-banana-pro/edit --input '{"prompt":"make it night","image_urls":["https://..."]}'`,
	Args: cobra.ExactArgs(1),
	RunE: runRunCmd,
}

func init() {
	runCmd.Flags().StringVar(&runInputFlag, "input", "", "JSON input payload (required)")
	runCmd.Flags().BoolVar(&runQueueFlag, "queue", false, "Use the queue (async) instead of sync")
	runCmd.Flags().BoolVar(&runLogsFlag, "logs", false, "Show model logs while polling queue (implies --queue)")
	_ = runCmd.MarkFlagRequired("input")
	rootCmd.AddCommand(runCmd)
}

func runRunCmd(cmd *cobra.Command, args []string) error {
	modelID := args[0]

	var payload map[string]any
	if err := json.Unmarshal([]byte(runInputFlag), &payload); err != nil {
		return fmt.Errorf("invalid --input JSON: %w", err)
	}

	if runQueueFlag || runLogsFlag {
		return runViaQueue(cmd, modelID, payload, runLogsFlag)
	}
	return runViaSync(cmd, modelID, payload)
}

func runViaSync(cmd *cobra.Command, modelID string, payload map[string]any) error {
	body, err := client.RunSync(modelID, payload)
	if err != nil {
		return err
	}
	return printResult(cmd, body)
}

func runViaQueue(cmd *cobra.Command, modelID string, payload map[string]any, withLogs bool) error {
	sub, err := client.QueueSubmit(modelID, payload)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Queued: %s\n", sub.RequestID)

	result, err := pollQueueUntilDone(modelID, sub.RequestID, withLogs)
	if err != nil {
		return err
	}
	return printResult(cmd, result)
}

// pollQueueUntilDone polls queue status until COMPLETED and returns the result bytes.
func pollQueueUntilDone(modelID, requestID string, withLogs bool) ([]byte, error) {
	seenLogs := map[string]bool{}
	delay := time.Second
	maxDelay := 10 * time.Second

	for {
		time.Sleep(delay)
		if delay < maxDelay {
			delay *= 2
			if delay > maxDelay {
				delay = maxDelay
			}
		}

		status, err := client.QueueStatus(modelID, requestID, withLogs)
		if err != nil {
			return nil, err
		}

		if withLogs {
			for _, log := range status.Logs {
				key := log.Timestamp + log.Message
				if !seenLogs[key] {
					seenLogs[key] = true
					fmt.Fprintf(os.Stderr, "[%s] %s\n", log.Level, log.Message)
				}
			}
		}

		switch status.Status {
		case "COMPLETED":
			return client.QueueResult(modelID, requestID)
		case "IN_QUEUE":
			if status.QueuePosition != nil {
				fmt.Fprintf(os.Stderr, "Queue position: %d\n", *status.QueuePosition)
			} else {
				fmt.Fprintf(os.Stderr, "In queue...\n")
			}
		case "IN_PROGRESS":
			fmt.Fprintf(os.Stderr, "In progress...\n")
		}
	}
}

// printResult writes result body to stdout in the right format.
func printResult(cmd *cobra.Command, body []byte) error {
	if output.IsJSON(cmd) {
		var v any
		if err := json.Unmarshal(body, &v); err != nil {
			_, err2 := os.Stdout.Write(body)
			return err2
		}
		return output.PrintJSON(v, output.IsPretty(cmd))
	}

	// Human-readable terminal output
	var v map[string]any
	if err := json.Unmarshal(body, &v); err != nil {
		_, err2 := os.Stdout.Write(body)
		return err2
	}
	printRunSummary(v)
	return nil
}

func printRunSummary(v map[string]any) {
	if images, ok := v["images"].([]any); ok {
		fmt.Printf("Generated %d image(s):\n", len(images))
		for i, img := range images {
			if m, ok := img.(map[string]any); ok {
				if u, ok := m["url"].(string); ok {
					fmt.Printf("  [%d] %s\n", i+1, u)
				}
			}
		}
	}
	if desc, ok := v["description"].(string); ok && desc != "" {
		fmt.Printf("Description: %s\n", desc)
	}
	if seed, ok := v["seed"]; ok && seed != nil {
		fmt.Printf("Seed: %v\n", seed)
	}
	// Fallback for non-image models: pretty-print full JSON
	if _, hasImages := v["images"]; !hasImages {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(v)
	}
}
