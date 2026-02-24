package cmd

// generate is a shortcut for fal-ai/nano-banana-pro (text-to-image).
//
// Usage:
//   fal generate "a cat wearing a hat"
//   fal generate "a cat" --aspect 16:9 --resolution 2K --num 4

import (
	"github.com/spf13/cobra"
)

var (
	generateAspect        string
	generateResolution    string
	generateNum           int
	generateFormat        string
	generateSafety        string
	generateSeed          int64
	generateWebSearch     bool
	generateGoogleSearch  bool
	generateQueue         bool
	generateLogs          bool
)

var generateCmd = &cobra.Command{
	Use:   "generate <prompt>",
	Short: "Generate images with nano-banana-pro (fal-ai/nano-banana-pro)",
	Long: `Shortcut for fal-ai/nano-banana-pro — Google's state-of-the-art image generation model.

Pricing: $0.15 per image (4K = $0.30/image, web search = +$0.015)

Examples:
  fal generate "a cat wearing a hat"
  fal generate "golden gate bridge at sunset" --aspect 16:9
  fal generate "portrait of a woman" --resolution 2K --num 2
  fal generate "futuristic city" --format webp --queue`,
	Args: cobra.ExactArgs(1),
	RunE: runGenerate,
}

func init() {
	generateCmd.Flags().StringVar(&generateAspect, "aspect", "1:1",
		"Aspect ratio: 21:9, 16:9, 3:2, 4:3, 5:4, 1:1, 4:5, 3:4, 2:3, 9:16, auto")
	generateCmd.Flags().StringVar(&generateResolution, "resolution", "1K",
		"Resolution: 1K, 2K, 4K (4K billed at 2x)")
	generateCmd.Flags().IntVar(&generateNum, "num", 1,
		"Number of images to generate (1-4)")
	generateCmd.Flags().StringVar(&generateFormat, "format", "png",
		"Output format: jpeg, png, webp")
	generateCmd.Flags().StringVar(&generateSafety, "safety", "4",
		"Safety tolerance 1 (strictest) to 6 (least strict)")
	generateCmd.Flags().Int64Var(&generateSeed, "seed", 0,
		"Random seed (0 = random)")
	generateCmd.Flags().BoolVar(&generateWebSearch, "web-search", false,
		"Enable web search grounding (+$0.015/image)")
	generateCmd.Flags().BoolVar(&generateGoogleSearch, "google-search", false,
		"Enable Google search grounding")
	generateCmd.Flags().BoolVar(&generateQueue, "queue", false,
		"Submit via queue instead of sync")
	generateCmd.Flags().BoolVar(&generateLogs, "logs", false,
		"Show model logs while polling queue (implies --queue)")

	rootCmd.AddCommand(generateCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
	prompt := args[0]

	payload := map[string]any{
		"prompt":           prompt,
		"aspect_ratio":     generateAspect,
		"resolution":       generateResolution,
		"num_images":       generateNum,
		"output_format":    generateFormat,
		"safety_tolerance": generateSafety,
	}

	if generateSeed != 0 {
		payload["seed"] = generateSeed
	}
	if generateWebSearch {
		payload["enable_web_search"] = true
	}
	if generateGoogleSearch {
		payload["enable_google_search"] = true
	}

	modelID := "fal-ai/nano-banana-pro"

	if generateQueue || generateLogs {
		return runViaQueue(cmd, modelID, payload, generateLogs)
	}
	return runViaSync(cmd, modelID, payload)
}
