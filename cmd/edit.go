package cmd

// edit-banana is a shortcut for fal-ai/nano-banana-2/edit (image-to-image).
//
// Usage:
//   fal edit-banana "make it night time" --image https://example.com/photo.jpg
//   fal edit-banana "make it night time" --file /absolute/path/to/photo.jpg
//   fal edit-banana "remove the car" --image https://... --file /path/to/other.jpg

import (
	"encoding/base64"
	"fmt"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// execCommand runs a command and returns its combined output.
func execCommand(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return string(out), err
}

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
	editR2Bucket     string
	editR2Domain     string
)

var editCmd = &cobra.Command{
	Use:   "edit-banana <prompt>",
	Short: "Edit images with nano-banana-2 (fal-ai/nano-banana-2/edit)",
	Long: `Shortcut for fal-ai/nano-banana-2/edit — state-of-the-art image editing model.

Provide image sources via --image (URL) and/or --file (local absolute path).
Local files are encoded as base64 data URIs by default.
Use --r2-bucket + --r2-domain to upload to R2 instead (better for large files).
Both flags are repeatable and can be combined.

Examples:
  fal edit-banana "make it night time" --image https://example.com/photo.jpg
  fal edit-banana "make it night time" --file /path/to/photo.jpg
  fal edit-banana "add snow" --file /path/to/city.jpg --aspect 16:9
  fal edit-banana "add snow" --file /path/to/city.jpg --r2-bucket my-pub --r2-domain pub.example.com
  fal edit-banana "remove background" --image https://... --file /path/to/other.jpg --num 2`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

func init() {
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
	editCmd.Flags().StringVar(&editR2Bucket, "r2-bucket", "",
		"Upload files to this R2 bucket instead of encoding as base64 (requires r2 CLI)")
	editCmd.Flags().StringVar(&editR2Domain, "r2-domain", "",
		"Public domain for R2 bucket (e.g. pub.example.com); required with --r2-bucket")
	rootCmd.AddCommand(editCmd)
}

// resolveImageSources converts local files to data URIs (base64) or uploads
// them to R2 when r2Bucket is set, then merges with any remote URLs.
func resolveImageSources(cmd *cobra.Command, urls []string, files []string, r2Bucket, r2Domain string) ([]string, error) {
	if len(urls) == 0 && len(files) == 0 {
		return nil, fmt.Errorf("at least one --image <url> or --file <path> is required")
	}

	result := make([]string, 0, len(urls)+len(files))
	result = append(result, urls...)

	if len(files) > 0 {
		if r2Bucket != "" {
			// R2 upload path
			if r2Domain == "" {
				return nil, fmt.Errorf("--r2-domain is required when using --r2-bucket")
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "Uploading %d file(s) to R2 bucket %q...\n", len(files), r2Bucket)
			for _, path := range files {
				publicURL, err := uploadToR2(cmd, path, r2Bucket, r2Domain)
				if err != nil {
					return nil, fmt.Errorf("R2 upload failed for %s: %w", path, err)
				}
				result = append(result, publicURL)
			}
		} else {
			// Base64 data URI path (default)
			fmt.Fprintf(cmd.ErrOrStderr(), "Encoding %d file(s) as base64...\n", len(files))
			for _, path := range files {
				dataURI, err := fileToDataURI(path)
				if err != nil {
					return nil, fmt.Errorf("encoding failed for %s: %w", path, err)
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "  ✓ %s (%s)\n", filepath.Base(path), mimeFromPath(path))
				result = append(result, dataURI)
			}
		}
	}

	return result, nil
}

// mimeFromPath returns the MIME type for a file based on its extension.
func mimeFromPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	t := mime.TypeByExtension(ext)
	if t == "" {
		return "application/octet-stream"
	}
	return t
}

// fileToDataURI reads a local file and returns a base64 data URI.
func fileToDataURI(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}
	mimeType := mimeFromPath(path)
	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded), nil
}

// uploadToR2 uploads a file to an R2 bucket via the r2 CLI and returns its public URL.
func uploadToR2(cmd *cobra.Command, localPath, bucket, domain string) (string, error) {
	key := "fal-tmp/" + filepath.Base(localPath)
	fmt.Fprintf(cmd.ErrOrStderr(), "  uploading %s → %s/%s\n", filepath.Base(localPath), bucket, key)

	out, err := execCommand("r2", "objects", "put", key, "--bucket", bucket, "--file", localPath)
	if err != nil {
		return "", fmt.Errorf("r2 put: %s: %w", strings.TrimSpace(out), err)
	}

	domain = strings.TrimRight(domain, "/")
	publicURL := fmt.Sprintf("https://%s/%s", domain, key)
	fmt.Fprintf(cmd.ErrOrStderr(), "  → %s\n", publicURL)
	return publicURL, nil
}

func runEdit(cmd *cobra.Command, args []string) error {
	prompt := args[0]

	imageURLs, err := resolveImageSources(cmd, editImages, editFiles, editR2Bucket, editR2Domain)
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
