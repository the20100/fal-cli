package cmd

// edit is a shortcut for fal-ai/nano-banana-pro/edit (image-to-image).
//
// Usage:
//   fal edit "make it night time" --image https://example.com/photo.jpg
//   fal edit "remove the car" --image https://... --image https://...

import (
	"github.com/spf13/cobra"
)

var (
	editImages       []string
	editAspect       string
	editResolution   string
	editNum          int
	editFormat       string
	editSafety       string
	editSeed         int64
	editWebSearch    bool
	editGoogleSearch bool
	editQueue        bool
	editLogs         bool
)

var editCmd = &cobra.Command{
	Use:   "edit <prompt>",
	Short: "Edit images with nano-banana-pro (fal-ai/nano-banana-pro/edit)",
	Long: `Shortcut for fal-ai/nano-banana-pro/edit — Google's state-of-the-art image editing model.

Provide one or more image URLs to edit, and a prompt describing the transformation.

Pricing: $0.15 per image (4K = $0.30/image, web search = +$0.015)

Examples:
  fal edit "make it night time" --image https://example.com/photo.jpg
  fal edit "add snow" --image https://example.com/city.jpg --aspect 16:9
  fal edit "remove background" --image https://... --image https://... --num 2`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

func init() {
	editCmd.Flags().StringArrayVar(&editImages, "image", nil,
		"Image URL(s) to edit (required, repeatable)")
	editCmd.Flags().StringVar(&editAspect, "aspect", "auto",
		"Aspect ratio: 21:9, 16:9, 3:2, 4:3, 5:4, 1:1, 4:5, 3:4, 2:3, 9:16, auto")
	editCmd.Flags().StringVar(&editResolution, "resolution", "1K",
		"Resolution: 1K, 2K, 4K (4K billed at 2x)")
	editCmd.Flags().IntVar(&editNum, "num", 1,
		"Number of images to generate (1-4)")
	editCmd.Flags().StringVar(&editFormat, "format", "png",
		"Output format: jpeg, png, webp")
	editCmd.Flags().StringVar(&editSafety, "safety", "4",
		"Safety tolerance 1 (strictest) to 6 (least strict)")
	editCmd.Flags().Int64Var(&editSeed, "seed", 0,
		"Random seed (0 = random)")
	editCmd.Flags().BoolVar(&editWebSearch, "web-search", false,
		"Enable web search grounding (+$0.015/image)")
	editCmd.Flags().BoolVar(&editGoogleSearch, "google-search", false,
		"Enable Google search grounding")
	editCmd.Flags().BoolVar(&editQueue, "queue", false,
		"Submit via queue instead of sync")
	editCmd.Flags().BoolVar(&editLogs, "logs", false,
		"Show model logs while polling queue (implies --queue)")
	_ = editCmd.MarkFlagRequired("image")

	rootCmd.AddCommand(editCmd)
}

func runEdit(cmd *cobra.Command, args []string) error {
	prompt := args[0]

	payload := map[string]any{
		"prompt":           prompt,
		"image_urls":       editImages,
		"aspect_ratio":     editAspect,
		"resolution":       editResolution,
		"num_images":       editNum,
		"output_format":    editFormat,
		"safety_tolerance": editSafety,
	}

	if editSeed != 0 {
		payload["seed"] = editSeed
	}
	if editWebSearch {
		payload["enable_web_search"] = true
	}
	if editGoogleSearch {
		payload["enable_google_search"] = true
	}

	modelID := "fal-ai/nano-banana-pro/edit"

	if editQueue || editLogs {
		return runViaQueue(cmd, modelID, payload, editLogs)
	}
	return runViaSync(cmd, modelID, payload)
}
