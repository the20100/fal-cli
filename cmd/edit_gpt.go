package cmd

// edit is a shortcut for openai/gpt-image-2/edit (image-to-image).
//
// Usage:
//   fal edit "make it night time" --image https://example.com/photo.jpg
//   fal edit "make it night time" --file /absolute/path/to/photo.jpg
//   fal edit "remove the car" --image https://... --file /path/to/other.jpg
//
// Resolution/quality recommendations:
//   quality low  → prefer 4K (higher detail compensates for lower quality)
//   quality medium → 2K or 4K both work well

import (
	"github.com/spf13/cobra"
)

var (
	gptEditImages       []string
	gptEditFiles        []string
	gptEditMask         string
	gptEditQuality      string
	gptEditResolution   string
	gptEditNum          int
	gptEditFormat       string
	gptEditQueue        bool
	gptEditLogs         bool
	gptEditR2Bucket     string
	gptEditR2Domain     string
)

var gptEditCmd = &cobra.Command{
	Use:   "edit <prompt>",
	Short: "Edit images with GPT Image 2 (openai/gpt-image-2/edit)",
	Long: `Shortcut for openai/gpt-image-2/edit — high-quality image editing model.

Provide image sources via --image (URL) and/or --file (local absolute path).
Local files are encoded as base64 data URIs by default.
Use --r2-bucket + --r2-domain to upload to R2 instead (better for large files).
Both flags are repeatable and can be combined.

Quality/resolution recommendations:
  quality low    → use 4K for best results
  quality medium → 2K or 4K both work well (default: 2K)
  quality high   → any resolution

Examples:
  fal edit "make it night time" --image https://example.com/photo.jpg --queue --json
  fal edit "make it night time" --file /path/to/photo.jpg --queue --json
  fal edit "add snow" --file /path/to/city.jpg --resolution 4K --queue --json
  fal edit "add snow" --file /path/to/city.jpg --r2-bucket my-pub --r2-domain pub.example.com --queue --json
  fal edit "remove background" --image https://... --file /path/to/other.jpg --num 2 --queue --json
  fal edit "detailed retouch" --quality low --resolution 4K --file /path/to/photo.jpg --queue --json`,
	Args: cobra.ExactArgs(1),
	RunE: runGptEdit,
}

func init() {
	gptEditCmd.Flags().StringArrayVar(&gptEditImages, "image", nil,
		"Image URL to edit (repeatable, combinable with --file)")
	gptEditCmd.Flags().StringArrayVar(&gptEditFiles, "file", nil,
		"Local image path to upload and edit (absolute path, repeatable)")
	gptEditCmd.Flags().StringVar(&gptEditMask, "mask", "",
		"Mask image URL (optional, marks areas to edit)")
	gptEditCmd.Flags().StringVar(&gptEditQuality, "quality", "medium",
		"Quality: low, medium, high (low → prefer 4K; medium → 2K or 4K)")
	gptEditCmd.Flags().StringVar(&gptEditResolution, "resolution", "2K",
		"Resolution: 2K (2048px), 4K (3840px)")
	gptEditCmd.Flags().IntVar(&gptEditNum, "num", 1,
		"Number of images to generate (1-4)")
	gptEditCmd.Flags().StringVar(&gptEditFormat, "format", "png",
		"Output format: jpeg, png, webp")
	gptEditCmd.Flags().BoolVar(&gptEditQueue, "queue", false,
		"Submit via queue instead of sync")
	gptEditCmd.Flags().BoolVar(&gptEditLogs, "logs", false,
		"Show model logs while polling queue (implies --queue)")
	gptEditCmd.Flags().StringVar(&gptEditR2Bucket, "r2-bucket", "",
		"Upload files to this R2 bucket instead of encoding as base64 (requires r2 CLI)")
	gptEditCmd.Flags().StringVar(&gptEditR2Domain, "r2-domain", "",
		"Public domain for R2 bucket (e.g. pub.example.com); required with --r2-bucket")
	rootCmd.AddCommand(gptEditCmd)
}

func runGptEdit(cmd *cobra.Command, args []string) error {
	prompt := args[0]

	imageURLs, err := resolveImageSources(cmd, gptEditImages, gptEditFiles, gptEditR2Bucket, gptEditR2Domain)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"prompt":        prompt,
		"image_urls":    imageURLs,
		"image_size":    resolutionToImageSize(gptEditResolution),
		"quality":       gptEditQuality,
		"num_images":    gptEditNum,
		"output_format": gptEditFormat,
	}

	if gptEditMask != "" {
		payload["mask_url"] = gptEditMask
	}

	modelID := "openai/gpt-image-2/edit"

	if gptEditQueue || gptEditLogs {
		return runViaQueue(cmd, modelID, payload, gptEditLogs)
	}
	return runViaSync(cmd, modelID, payload)
}
