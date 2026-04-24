package cmd

// generate is a shortcut for openai/gpt-image-2 (text-to-image).
//
// Usage:
//   fal generate "a cat wearing a hat"
//   fal generate "a cat" --quality high --resolution 4K --num 2
//
// Resolution/quality recommendations:
//   quality low  → prefer 4K (higher detail compensates for lower quality)
//   quality medium → 2K or 4K both work well

import (
	"github.com/spf13/cobra"
)

var (
	gptGenerateQuality    string
	gptGenerateResolution string
	gptGenerateNum        int
	gptGenerateFormat     string
	gptGenerateQueue      bool
	gptGenerateLogs       bool
)

var gptGenerateCmd = &cobra.Command{
	Use:   "generate <prompt>",
	Short: "Generate images with GPT Image 2 (openai/gpt-image-2)",
	Long: `Shortcut for openai/gpt-image-2 — high-quality text-to-image model.

Quality/resolution recommendations:
  quality low    → use 4K for best results
  quality medium → 2K or 4K both work well (default: 2K)
  quality high   → any resolution

Examples:
  fal generate "a cat wearing a hat" --queue --json
  fal generate "golden gate bridge at sunset" --quality high --queue --json
  fal generate "portrait of a woman" --resolution 4K --num 2 --queue --json
  fal generate "futuristic city" --format webp --queue --json
  fal generate "detailed artwork" --quality low --resolution 4K --queue --json`,
	Args: cobra.ExactArgs(1),
	RunE: runGptGenerate,
}

func init() {
	gptGenerateCmd.Flags().StringVar(&gptGenerateQuality, "quality", "medium",
		"Quality: low, medium, high (low → prefer 4K; medium → 2K or 4K)")
	gptGenerateCmd.Flags().StringVar(&gptGenerateResolution, "resolution", "2K",
		"Resolution: 2K (2048px), 4K (3840px)")
	gptGenerateCmd.Flags().IntVar(&gptGenerateNum, "num", 1,
		"Number of images to generate (1-4)")
	gptGenerateCmd.Flags().StringVar(&gptGenerateFormat, "format", "png",
		"Output format: jpeg, png, webp")
	gptGenerateCmd.Flags().BoolVar(&gptGenerateQueue, "queue", false,
		"Submit via queue instead of sync")
	gptGenerateCmd.Flags().BoolVar(&gptGenerateLogs, "logs", false,
		"Show model logs while polling queue (implies --queue)")
	rootCmd.AddCommand(gptGenerateCmd)
}

func resolutionToImageSize(resolution string) map[string]int {
	switch resolution {
	case "4K":
		return map[string]int{"width": 3840, "height": 3840}
	default: // 2K
		return map[string]int{"width": 2048, "height": 2048}
	}
}

func runGptGenerate(cmd *cobra.Command, args []string) error {
	prompt := args[0]

	payload := map[string]any{
		"prompt":        prompt,
		"image_size":    resolutionToImageSize(gptGenerateResolution),
		"quality":       gptGenerateQuality,
		"num_images":    gptGenerateNum,
		"output_format": gptGenerateFormat,
	}

	modelID := "openai/gpt-image-2"

	if gptGenerateQueue || gptGenerateLogs {
		return runViaQueue(cmd, modelID, payload, gptGenerateLogs)
	}
	return runViaSync(cmd, modelID, payload)
}
