package cmd

// edit is a shortcut for fal-ai/nano-banana-2/edit (image-to-image).
// edit-old is a shortcut for fal-ai/nano-banana-pro/edit (image-to-image, previous model).
//
// Usage:
//   fal edit "make it night time" --image https://example.com/photo.jpg
//   fal edit "remove the car" --image https://... --image https://...
//   fal edit-old "make it night time" --image https://example.com/photo.jpg

import (
	"github.com/spf13/cobra"
)

// --- edit (nano-banana-2/edit) ---

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
	Short: "Edit images with nano-banana-2 (fal-ai/nano-banana-2/edit)",
	Long: `Shortcut for fal-ai/nano-banana-2/edit — state-of-the-art image editing model.

Provide one or more image URLs to edit, and a prompt describing the transformation.

Examples:
  fal edit "make it night time" --image https://example.com/photo.jpg
  fal edit "add snow" --image https://example.com/city.jpg --aspect 16:9
  fal edit "remove background" --image https://... --image https://... --num 2`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

// --- edit-old (nano-banana-pro/edit) ---

var (
	editOldImages       []string
	editOldAspect       string
	editOldResolution   string
	editOldNum          int
	editOldFormat       string
	editOldSafety       string
	editOldSeed         int64
	editOldWebSearch    bool
	editOldGoogleSearch bool
	editOldQueue        bool
	editOldLogs         bool
)

var editOldCmd = &cobra.Command{
	Use:   "edit-old <prompt>",
	Short: "Edit images with nano-banana-pro (fal-ai/nano-banana-pro/edit, previous model)",
	Long: `Shortcut for fal-ai/nano-banana-pro/edit — previous image editing model.

Provide one or more image URLs to edit, and a prompt describing the transformation.

Examples:
  fal edit-old "make it night time" --image https://example.com/photo.jpg
  fal edit-old "add snow" --image https://example.com/city.jpg --aspect 16:9
  fal edit-old "remove background" --image https://... --image https://... --num 2`,
	Args: cobra.ExactArgs(1),
	RunE: runEditOld,
}

func init() {
	// edit flags
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

	// edit-old flags
	editOldCmd.Flags().StringArrayVar(&editOldImages, "image", nil,
		"Image URL(s) to edit (required, repeatable)")
	editOldCmd.Flags().StringVar(&editOldAspect, "aspect", "auto",
		"Aspect ratio: 21:9, 16:9, 3:2, 4:3, 5:4, 1:1, 4:5, 3:4, 2:3, 9:16, auto")
	editOldCmd.Flags().StringVar(&editOldResolution, "resolution", "1K",
		"Resolution: 1K, 2K, 4K (4K billed at 2x)")
	editOldCmd.Flags().IntVar(&editOldNum, "num", 1,
		"Number of images to generate (1-4)")
	editOldCmd.Flags().StringVar(&editOldFormat, "format", "png",
		"Output format: jpeg, png, webp")
	editOldCmd.Flags().StringVar(&editOldSafety, "safety", "4",
		"Safety tolerance 1 (strictest) to 6 (least strict)")
	editOldCmd.Flags().Int64Var(&editOldSeed, "seed", 0,
		"Random seed (0 = random)")
	editOldCmd.Flags().BoolVar(&editOldWebSearch, "web-search", false,
		"Enable web search grounding (+$0.015/image)")
	editOldCmd.Flags().BoolVar(&editOldGoogleSearch, "google-search", false,
		"Enable Google search grounding")
	editOldCmd.Flags().BoolVar(&editOldQueue, "queue", false,
		"Submit via queue instead of sync")
	editOldCmd.Flags().BoolVar(&editOldLogs, "logs", false,
		"Show model logs while polling queue (implies --queue)")
	_ = editOldCmd.MarkFlagRequired("image")
	rootCmd.AddCommand(editOldCmd)
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

	modelID := "fal-ai/nano-banana-2/edit"

	if editQueue || editLogs {
		return runViaQueue(cmd, modelID, payload, editLogs)
	}
	return runViaSync(cmd, modelID, payload)
}

func runEditOld(cmd *cobra.Command, args []string) error {
	prompt := args[0]

	payload := map[string]any{
		"prompt":           prompt,
		"image_urls":       editOldImages,
		"aspect_ratio":     editOldAspect,
		"resolution":       editOldResolution,
		"num_images":       editOldNum,
		"output_format":    editOldFormat,
		"safety_tolerance": editOldSafety,
	}

	if editOldSeed != 0 {
		payload["seed"] = editOldSeed
	}
	if editOldWebSearch {
		payload["enable_web_search"] = true
	}
	if editOldGoogleSearch {
		payload["enable_google_search"] = true
	}

	modelID := "fal-ai/nano-banana-pro/edit"

	if editOldQueue || editOldLogs {
		return runViaQueue(cmd, modelID, payload, editOldLogs)
	}
	return runViaSync(cmd, modelID, payload)
}
