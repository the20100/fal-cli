package cmd

// edit is a shortcut for fal-ai/nano-banana-2/edit (image-to-image).
// edit-old is a shortcut for fal-ai/nano-banana-pro/edit (image-to-image, previous model).
//
// Usage:
//   fal edit "make it night time" --image https://example.com/photo.jpg
//   fal edit "make it night time" --file /absolute/path/to/photo.jpg
//   fal edit "remove the car" --image https://... --file /path/to/other.jpg

import (
	"fmt"

	"github.com/spf13/cobra"
)

// --- edit (nano-banana-2/edit) ---

var (
	editImages       []string
	editFiles        []string
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

Provide image sources via --image (URL) and/or --file (local absolute path).
Local files are uploaded to fal.ai storage before the edit request is sent.
Both flags are repeatable and can be combined.

Examples:
  fal edit "make it night time" --image https://example.com/photo.jpg
  fal edit "make it night time" --file /path/to/photo.jpg
  fal edit "add snow" --file /path/to/city.jpg --aspect 16:9
  fal edit "remove background" --image https://... --file /path/to/other.jpg --num 2`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

// --- edit-old (nano-banana-pro/edit) ---

var (
	editOldImages       []string
	editOldFiles        []string
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

Provide image sources via --image (URL) and/or --file (local absolute path).
Local files are uploaded to fal.ai storage before the edit request is sent.
Both flags are repeatable and can be combined.

Examples:
  fal edit-old "make it night time" --image https://example.com/photo.jpg
  fal edit-old "make it night time" --file /path/to/photo.jpg
  fal edit-old "remove background" --image https://... --file /path/to/other.jpg --num 2`,
	Args: cobra.ExactArgs(1),
	RunE: runEditOld,
}

func init() {
	// edit flags
	editCmd.Flags().StringArrayVar(&editImages, "image", nil,
		"Image URL to edit (repeatable, combinable with --file)")
	editCmd.Flags().StringArrayVar(&editFiles, "file", nil,
		"Local image path to upload and edit (absolute path, repeatable)")
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
	rootCmd.AddCommand(editCmd)

	// edit-old flags
	editOldCmd.Flags().StringArrayVar(&editOldImages, "image", nil,
		"Image URL to edit (repeatable, combinable with --file)")
	editOldCmd.Flags().StringArrayVar(&editOldFiles, "file", nil,
		"Local image path to upload and edit (absolute path, repeatable)")
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
	rootCmd.AddCommand(editOldCmd)
}

// resolveImageSources uploads local files and merges them with URL images.
// Returns a combined slice of URLs ready to send to the API.
func resolveImageSources(cmd *cobra.Command, urls []string, files []string) ([]string, error) {
	if len(urls) == 0 && len(files) == 0 {
		return nil, fmt.Errorf("at least one --image <url> or --file <path> is required")
	}

	result := make([]string, 0, len(urls)+len(files))
	result = append(result, urls...)

	if len(files) > 0 {
		fmt.Fprintf(cmd.ErrOrStderr(), "Uploading %d local file(s)...\n", len(files))
		for _, path := range files {
			fmt.Fprintf(cmd.ErrOrStderr(), "  uploading %s\n", path)
			uploadedURL, err := client.UploadFile(path)
			if err != nil {
				return nil, fmt.Errorf("upload failed for %s: %w", path, err)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "  → %s\n", uploadedURL)
			result = append(result, uploadedURL)
		}
	}

	return result, nil
}

func runEdit(cmd *cobra.Command, args []string) error {
	prompt := args[0]

	imageURLs, err := resolveImageSources(cmd, editImages, editFiles)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"prompt":           prompt,
		"image_urls":       imageURLs,
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

	imageURLs, err := resolveImageSources(cmd, editOldImages, editOldFiles)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"prompt":           prompt,
		"image_urls":       imageURLs,
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
