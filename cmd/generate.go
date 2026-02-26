package cmd

// generate is a shortcut for fal-ai/nano-banana-2 (text-to-image).
// generate-old is a shortcut for fal-ai/nano-banana-pro (text-to-image, previous model).
//
// Usage:
//   fal generate "a cat wearing a hat"
//   fal generate "a cat" --aspect 16:9 --resolution 2K --num 4
//   fal generate-old "a cat wearing a hat"

import (
	"github.com/spf13/cobra"
)

// --- generate (nano-banana-2) ---

var (
	generateAspect       string
	generateResolution   string
	generateNum          int
	generateFormat       string
	generateSafety       string
	generateSeed         int64
	generateWebSearch    bool
	generateGoogleSearch bool
	generateQueue        bool
	generateLogs         bool
)

var generateCmd = &cobra.Command{
	Use:   "generate <prompt>",
	Short: "Generate images with nano-banana-2 (fal-ai/nano-banana-2)",
	Long: `Shortcut for fal-ai/nano-banana-2 — state-of-the-art text-to-image model.

Examples:
  fal generate "a cat wearing a hat"
  fal generate "golden gate bridge at sunset" --aspect 16:9
  fal generate "portrait of a woman" --resolution 2K --num 2
  fal generate "futuristic city" --format webp --queue`,
	Args: cobra.ExactArgs(1),
	RunE: runGenerate,
}

// --- generate-old (nano-banana-pro) ---

var (
	generateOldAspect       string
	generateOldResolution   string
	generateOldNum          int
	generateOldFormat       string
	generateOldSafety       string
	generateOldSeed         int64
	generateOldWebSearch    bool
	generateOldGoogleSearch bool
	generateOldQueue        bool
	generateOldLogs         bool
)

var generateOldCmd = &cobra.Command{
	Use:   "generate-old <prompt>",
	Short: "Generate images with nano-banana-pro (fal-ai/nano-banana-pro, previous model)",
	Long: `Shortcut for fal-ai/nano-banana-pro — previous text-to-image model.

Examples:
  fal generate-old "a cat wearing a hat"
  fal generate-old "golden gate bridge at sunset" --aspect 16:9
  fal generate-old "portrait of a woman" --resolution 2K --num 2`,
	Args: cobra.ExactArgs(1),
	RunE: runGenerateOld,
}

func init() {
	// generate flags
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

	// generate-old flags
	generateOldCmd.Flags().StringVar(&generateOldAspect, "aspect", "1:1",
		"Aspect ratio: 21:9, 16:9, 3:2, 4:3, 5:4, 1:1, 4:5, 3:4, 2:3, 9:16, auto")
	generateOldCmd.Flags().StringVar(&generateOldResolution, "resolution", "1K",
		"Resolution: 1K, 2K, 4K (4K billed at 2x)")
	generateOldCmd.Flags().IntVar(&generateOldNum, "num", 1,
		"Number of images to generate (1-4)")
	generateOldCmd.Flags().StringVar(&generateOldFormat, "format", "png",
		"Output format: jpeg, png, webp")
	generateOldCmd.Flags().StringVar(&generateOldSafety, "safety", "4",
		"Safety tolerance 1 (strictest) to 6 (least strict)")
	generateOldCmd.Flags().Int64Var(&generateOldSeed, "seed", 0,
		"Random seed (0 = random)")
	generateOldCmd.Flags().BoolVar(&generateOldWebSearch, "web-search", false,
		"Enable web search grounding (+$0.015/image)")
	generateOldCmd.Flags().BoolVar(&generateOldGoogleSearch, "google-search", false,
		"Enable Google search grounding")
	generateOldCmd.Flags().BoolVar(&generateOldQueue, "queue", false,
		"Submit via queue instead of sync")
	generateOldCmd.Flags().BoolVar(&generateOldLogs, "logs", false,
		"Show model logs while polling queue (implies --queue)")
	rootCmd.AddCommand(generateOldCmd)
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

	modelID := "fal-ai/nano-banana-2"

	if generateQueue || generateLogs {
		return runViaQueue(cmd, modelID, payload, generateLogs)
	}
	return runViaSync(cmd, modelID, payload)
}

func runGenerateOld(cmd *cobra.Command, args []string) error {
	prompt := args[0]

	payload := map[string]any{
		"prompt":           prompt,
		"aspect_ratio":     generateOldAspect,
		"resolution":       generateOldResolution,
		"num_images":       generateOldNum,
		"output_format":    generateOldFormat,
		"safety_tolerance": generateOldSafety,
	}

	if generateOldSeed != 0 {
		payload["seed"] = generateOldSeed
	}
	if generateOldWebSearch {
		payload["enable_web_search"] = true
	}
	if generateOldGoogleSearch {
		payload["enable_google_search"] = true
	}

	modelID := "fal-ai/nano-banana-pro"

	if generateOldQueue || generateOldLogs {
		return runViaQueue(cmd, modelID, payload, generateOldLogs)
	}
	return runViaSync(cmd, modelID, payload)
}
