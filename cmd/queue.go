package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vincentmaurin/fal-cli/internal/output"
)

var queueCmd = &cobra.Command{
	Use:   "queue",
	Short: "Manage queue requests",
}

var queueStatusCmd = &cobra.Command{
	Use:   "status <model-id> <request-id>",
	Short: "Check the status of a queued request",
	Long: `Check the status of a queued request.

Status values: IN_QUEUE, IN_PROGRESS, COMPLETED

Examples:
  fal queue status fal-ai/flux/dev abc123
  fal queue status fal-ai/flux/dev abc123 --logs`,
	Args: cobra.ExactArgs(2),
	RunE: runQueueStatus,
}

var queueResultCmd = &cobra.Command{
	Use:   "result <model-id> <request-id>",
	Short: "Get the result of a completed queued request",
	Long: `Retrieve the result of a completed queued request.

Examples:
  fal queue result fal-ai/flux/dev abc123`,
	Args: cobra.ExactArgs(2),
	RunE: runQueueResult,
}

var queueCancelCmd = &cobra.Command{
	Use:   "cancel <model-id> <request-id>",
	Short: "Cancel a queued request",
	Long: `Cancel a request that is waiting in the queue.

Only works for requests with status IN_QUEUE.

Examples:
  fal queue cancel fal-ai/flux/dev abc123`,
	Args: cobra.ExactArgs(2),
	RunE: runQueueCancel,
}

var queuePollCmd = &cobra.Command{
	Use:   "poll <model-id> <request-id>",
	Short: "Poll a queued request until completion and print the result",
	Long: `Poll a queued request until it completes, then print the result.

Examples:
  fal queue poll fal-ai/flux/dev abc123
  fal queue poll fal-ai/flux/dev abc123 --logs`,
	Args: cobra.ExactArgs(2),
	RunE: runQueuePoll,
}

var queueLogsFlag bool

func init() {
	queueStatusCmd.Flags().BoolVar(&queueLogsFlag, "logs", false, "Include model logs in output")
	queuePollCmd.Flags().BoolVar(&queueLogsFlag, "logs", false, "Show model logs while polling")

	queueCmd.AddCommand(queueStatusCmd, queueResultCmd, queueCancelCmd, queuePollCmd)
	rootCmd.AddCommand(queueCmd)
}

func runQueueStatus(cmd *cobra.Command, args []string) error {
	modelID, requestID := args[0], args[1]

	status, err := client.QueueStatus(modelID, requestID, queueLogsFlag)
	if err != nil {
		return err
	}

	if output.IsJSON(cmd) {
		return output.PrintJSON(status, output.IsPretty(cmd))
	}

	rows := [][]string{
		{"STATUS", status.Status},
	}
	if status.QueuePosition != nil {
		rows = append(rows, []string{"POSITION", fmt.Sprintf("%d", *status.QueuePosition)})
	}
	output.PrintKeyValue(rows)

	if queueLogsFlag && len(status.Logs) > 0 {
		fmt.Fprintln(os.Stderr, "\nLogs:")
		for _, log := range status.Logs {
			fmt.Fprintf(os.Stderr, "  [%s] %s\n", log.Level, log.Message)
		}
	}
	return nil
}

func runQueueResult(cmd *cobra.Command, args []string) error {
	modelID, requestID := args[0], args[1]

	body, err := client.QueueResult(modelID, requestID)
	if err != nil {
		return err
	}

	return printResult(cmd, body)
}

func runQueueCancel(cmd *cobra.Command, args []string) error {
	modelID, requestID := args[0], args[1]

	if err := client.QueueCancel(modelID, requestID); err != nil {
		return err
	}

	fmt.Printf("Cancellation requested for: %s\n", requestID)
	return nil
}

func runQueuePoll(cmd *cobra.Command, args []string) error {
	modelID, requestID := args[0], args[1]

	fmt.Fprintf(os.Stderr, "Polling: %s\n", requestID)

	result, err := pollQueueUntilDone(modelID, requestID, queueLogsFlag)
	if err != nil {
		return err
	}

	return printResult(cmd, result)
}
